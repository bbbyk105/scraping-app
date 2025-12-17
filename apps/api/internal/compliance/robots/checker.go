package robots

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Checker checks robots.txt compliance
type Checker struct {
	cache     Cache
	ttl       time.Duration
	httpClient *http.Client
	logger    *slog.Logger
	mu        sync.RWMutex
	memoryCache map[string]cacheEntry
}

type cacheEntry struct {
	content   []byte
	expiresAt time.Time
}

// NewChecker creates a new robots.txt checker
func NewChecker(cache Cache, ttl time.Duration, httpClient *http.Client, logger *slog.Logger) *Checker {
	checker := &Checker{
		cache:      cache,
		ttl:        ttl,
		httpClient: httpClient,
		logger:     logger,
		memoryCache: make(map[string]cacheEntry),
	}
	return checker
}

// CanFetch checks if a URL can be fetched according to robots.txt
// Returns: (allowed, ruleGroup, error)
// ruleGroup is the User-agent group that matched (e.g., "User-agent: *")
func (c *Checker) CanFetch(ctx context.Context, targetURL, userAgent string) (bool, string, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, "", fmt.Errorf("invalid URL: %w", err)
	}

	// Build robots.txt URL
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)
	cacheKey := fmt.Sprintf("robots:%s://%s", u.Scheme, u.Host)

	// Try to get from cache
	robotsContent, err := c.getRobotsTxt(ctx, cacheKey, robotsURL)
	if err != nil {
		// On error, fail safe (block access)
		// In production, you might want to make this configurable
		c.logger.Warn("Failed to fetch robots.txt, blocking access for safety",
			"url", targetURL,
			"robots_url", robotsURL,
			"error", err)
		return false, "", fmt.Errorf("robots.txt check failed: %w", err)
	}

	// Parse robots.txt
	rules := c.parseRobotsTxt(robotsContent, userAgent)

	// Check if path is allowed
	path := u.Path
	if path == "" {
		path = "/"
	}

	allowed, ruleGroup := c.checkPath(path, rules, userAgent)
	return allowed, ruleGroup, nil
}

func (c *Checker) getRobotsTxt(ctx context.Context, cacheKey, robotsURL string) ([]byte, error) {
	// Try cache first
	if c.cache != nil {
		content, err := c.cache.Get(ctx, cacheKey)
		if err == nil && len(content) > 0 {
			return content, nil
		}
	}

	// Try memory cache
	c.mu.RLock()
	if entry, ok := c.memoryCache[cacheKey]; ok {
		if time.Now().Before(entry.expiresAt) {
			content := entry.content
			c.mu.RUnlock()
			return content, nil
		}
	}
	c.mu.RUnlock()

	// Fetch from network
	content, err := c.fetchRobotsTxt(ctx, robotsURL)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if c.cache != nil {
		_ = c.cache.Set(ctx, cacheKey, content, c.ttl)
	}

	// Store in memory cache
	c.mu.Lock()
	c.memoryCache[cacheKey] = cacheEntry{
		content:   content,
		expiresAt: time.Now().Add(c.ttl),
	}
	c.mu.Unlock()

	return content, nil
}

func (c *Checker) fetchRobotsTxt(ctx context.Context, robotsURL string) ([]byte, error) {
	// Try HTTPS first if URL is HTTP
	var httpsURL string
	if strings.HasPrefix(robotsURL, "http://") {
		httpsURL = strings.Replace(robotsURL, "http://", "https://", 1)
	} else {
		httpsURL = robotsURL
	}

	req, err := http.NewRequestWithContext(ctx, "GET", httpsURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		content := make([]byte, 0, 8192)
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				content = append(content, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		return content, nil
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Fallback to HTTP if HTTPS failed and original was HTTP
	if strings.HasPrefix(robotsURL, "http://") && strings.HasPrefix(httpsURL, "https://") {
		req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("robots.txt returned status %d", resp.StatusCode)
		}

		content := make([]byte, 0, 8192)
		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				content = append(content, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		return content, nil
	}

	return nil, fmt.Errorf("failed to fetch robots.txt")
}

// RobotsRule represents a single rule from robots.txt
type RobotsRule struct {
	UserAgent string
	Disallow  []string
	Allow     []string
}

func (c *Checker) parseRobotsTxt(content []byte, userAgent string) []RobotsRule {
	lines := strings.Split(string(content), "\n")
	var rules []RobotsRule
	var currentRule *RobotsRule

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			// Save previous rule
			if currentRule != nil {
				rules = append(rules, *currentRule)
			}
			// Start new rule
			currentRule = &RobotsRule{
				UserAgent: value,
				Disallow:  []string{},
				Allow:     []string{},
			}
		case "disallow":
			if currentRule != nil {
				currentRule.Disallow = append(currentRule.Disallow, value)
			}
		case "allow":
			if currentRule != nil {
				currentRule.Allow = append(currentRule.Allow, value)
			}
		}
	}

	// Save last rule
	if currentRule != nil {
		rules = append(rules, *currentRule)
	}

	return rules
}

func (c *Checker) checkPath(path string, rules []RobotsRule, userAgent string) (bool, string) {
	// Find matching user-agent rules
	// Priority: exact match > * (wildcard)
	var matchedRules []RobotsRule
	var wildcardRule *RobotsRule

	for _, rule := range rules {
		if rule.UserAgent == userAgent {
			matchedRules = append(matchedRules, rule)
		} else if rule.UserAgent == "*" {
			wildcardRule = &rule
		}
	}

	// Use wildcard if no specific match
	if len(matchedRules) == 0 && wildcardRule != nil {
		matchedRules = append(matchedRules, *wildcardRule)
	}

	// If no rules match, allow by default (per robots.txt spec)
	if len(matchedRules) == 0 {
		return true, ""
	}

	// Check all matched rules
	// If any rule allows, it's allowed
	// If all rules disallow, it's disallowed
	for _, rule := range matchedRules {
		allowed, ruleGroup := c.checkPathAgainstRule(path, rule)
		if allowed {
			return true, ruleGroup
		}
	}

	// All rules disallow
	return false, matchedRules[0].UserAgent
}

func (c *Checker) checkPathAgainstRule(path string, rule RobotsRule) (bool, string) {
	// Check Allow rules first (they take precedence)
	for _, allowPath := range rule.Allow {
		if c.pathMatches(path, allowPath) {
			return true, rule.UserAgent
		}
	}

	// Check Disallow rules
	for _, disallowPath := range rule.Disallow {
		if c.pathMatches(path, disallowPath) {
			return false, rule.UserAgent
		}
	}

	// If no disallow matches, allow
	return true, rule.UserAgent
}

func (c *Checker) pathMatches(path, pattern string) bool {
	// Empty pattern means allow all
	if pattern == "" {
		return false
	}

	// Pattern must be a prefix of path
	if !strings.HasPrefix(path, pattern) {
		return false
	}

	// Exact match
	if path == pattern {
		return true
	}

	// Pattern ends with /, so any path starting with it matches
	if strings.HasSuffix(pattern, "/") {
		return true
	}

	// Pattern doesn't end with /, so only exact match or path starting with pattern/
	if strings.HasPrefix(path, pattern+"/") {
		return true
	}

	return false
}

