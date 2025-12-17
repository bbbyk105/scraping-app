package httpclient

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/pricecompare/api/internal/audit"
	"github.com/pricecompare/api/internal/compliance/robots"
	"github.com/pricecompare/api/internal/ratelimit"
)

// RedisClientOptional is an optional Redis client interface
type RedisClientOptional interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// Client is a compliant HTTP client with robots.txt checking, rate limiting, and audit logging
type Client struct {
	httpClient *http.Client
	robots     *robots.Checker
	limiter    *ratelimit.Manager
	cfg        *Config
	logger     *slog.Logger
}

// New creates a new HTTP client with compliance features
func New(cfg *Config, logger *slog.Logger, redisClient RedisClientOptional) *Client {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.HTTPTimeoutSeconds) * time.Second,
	}

	// Create robots.txt checker
	var robotsCache robots.Cache
	if redisClient != nil {
		// Use Redis cache if available
		robotsCache = &redisCacheAdapter{client: redisClient}
	}
	robotsChecker := robots.NewChecker(
		robotsCache,
		time.Duration(cfg.RobotsCacheTTLHours)*time.Hour,
		httpClient,
		logger,
	)

	// Create rate limiter
	rateLimitConfigs := make(map[string]ratelimit.RateLimitConfig)
	for k, v := range cfg.ProviderRateLimits {
		rateLimitConfigs[k] = ratelimit.RateLimitConfig{
			RPS:   v.RPS,
			Burst: v.Burst,
		}
	}
	defaultRateLimit := ratelimit.RateLimitConfig{
		RPS:   cfg.DefaultRateLimit.RPS,
		Burst: cfg.DefaultRateLimit.Burst,
	}
	limiter := ratelimit.NewManager(rateLimitConfigs, defaultRateLimit, logger)

	return &Client{
		httpClient: httpClient,
		robots:     robotsChecker,
		limiter:    limiter,
		cfg:        cfg,
		logger:     logger,
	}
}

// Get performs a GET request with compliance checks
func (c *Client) Get(ctx context.Context, providerKey, targetURL string) (*http.Response, error) {
	startTime := time.Now()
	var retryCount int
	var robotsAllowed bool
	var robotsGroup string
	var lastErr error

	// Check if URL is external
	isExternal, err := IsExternalURL(targetURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Block external URLs if live fetch is disabled
	if isExternal && !c.cfg.AllowLiveFetch {
		audit.LogRequest(c.logger, audit.Entry{
			Timestamp:     startTime,
			Provider:      providerKey,
			Method:        "GET",
			URL:           targetURL,
			Host:          getHost(targetURL),
			Path:          getPath(targetURL),
			Status:        0,
			DurationMs:    time.Since(startTime).Milliseconds(),
			UserAgent:     c.cfg.UserAgent,
			RobotsAllowed: false,
			RetryCount:    0,
			Error:         "ALLOW_LIVE_FETCH is false, external URL access blocked",
		})
		return nil, fmt.Errorf("live fetch is disabled (ALLOW_LIVE_FETCH=false), cannot access external URL: %s", targetURL)
	}

	// Check robots.txt for external URLs
	if isExternal {
		allowed, group, err := c.robots.CanFetch(ctx, targetURL, c.cfg.UserAgent)
		if err != nil {
			audit.LogRequest(c.logger, audit.Entry{
				Timestamp:     startTime,
				Provider:      providerKey,
				Method:        "GET",
				URL:           targetURL,
				Host:          getHost(targetURL),
				Path:          getPath(targetURL),
				Status:        0,
				DurationMs:    time.Since(startTime).Milliseconds(),
				UserAgent:     c.cfg.UserAgent,
				RobotsAllowed: false,
				RetryCount:    0,
				Error:         fmt.Sprintf("robots.txt check failed: %v", err),
			})
			return nil, fmt.Errorf("robots.txt check failed: %w", err)
		}
		if !allowed {
			audit.LogRequest(c.logger, audit.Entry{
				Timestamp:     startTime,
				Provider:      providerKey,
				Method:        "GET",
				URL:           targetURL,
				Host:          getHost(targetURL),
				Path:          getPath(targetURL),
				Status:        0,
				DurationMs:    time.Since(startTime).Milliseconds(),
				UserAgent:     c.cfg.UserAgent,
				RobotsAllowed: false,
				RobotsGroup:   group,
				RetryCount:    0,
				Error:         "robots.txt disallows this path",
			})
			return nil, fmt.Errorf("robots.txt disallows access to %s (matched rule: %s)", targetURL, group)
		}
		robotsAllowed = true
		robotsGroup = group
	}

	// Apply rate limiting
	if err := c.limiter.Wait(ctx, providerKey); err != nil {
		return nil, fmt.Errorf("rate limit wait failed: %w", err)
	}

	// Perform request with retries
	maxRetries := c.cfg.HTTPMaxRetries
	for attempt := 0; attempt <= maxRetries; attempt++ {
		retryCount = attempt

		req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
		if err != nil {
			lastErr = err
			break
		}

		req.Header.Set("User-Agent", c.cfg.UserAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			// Retry on network errors
			if attempt < maxRetries {
				backoff := exponentialBackoff(attempt)
				c.logger.Warn("HTTP request failed, retrying",
					"url", targetURL,
					"attempt", attempt+1,
					"max_retries", maxRetries,
					"backoff_ms", backoff.Milliseconds(),
					"error", err)
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(backoff):
					continue
				}
			}
			continue
		}

		// Check status code - retry on 429 and 5xx
		shouldRetry := resp.StatusCode == 429 || (resp.StatusCode >= 500 && resp.StatusCode < 600)
		if shouldRetry && attempt < maxRetries {
			resp.Body.Close()
			backoff := exponentialBackoff(attempt)
			c.logger.Warn("HTTP request returned retryable status, retrying",
				"url", targetURL,
				"status", resp.StatusCode,
				"attempt", attempt+1,
				"max_retries", maxRetries,
				"backoff_ms", backoff.Milliseconds())
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				continue
			}
		}

		// Success or non-retryable error
		duration := time.Since(startTime)
		audit.LogRequest(c.logger, audit.Entry{
			Timestamp:     startTime,
			Provider:      providerKey,
			Method:        "GET",
			URL:           targetURL,
			Host:          getHost(targetURL),
			Path:          getPath(targetURL),
			Status:        resp.StatusCode,
			DurationMs:    duration.Milliseconds(),
			UserAgent:     c.cfg.UserAgent,
			RobotsAllowed: robotsAllowed,
			RobotsGroup:   robotsGroup,
			RetryCount:    retryCount,
		})

		return resp, nil
	}

	// All retries exhausted
	duration := time.Since(startTime)
	errorMsg := "unknown error"
	if lastErr != nil {
		errorMsg = lastErr.Error()
	}
	audit.LogRequest(c.logger, audit.Entry{
		Timestamp:     startTime,
		Provider:      providerKey,
		Method:        "GET",
		URL:           targetURL,
		Host:          getHost(targetURL),
		Path:          getPath(targetURL),
		Status:        0,
		DurationMs:    duration.Milliseconds(),
		UserAgent:     c.cfg.UserAgent,
		RobotsAllowed: robotsAllowed,
		RobotsGroup:   robotsGroup,
		RetryCount:    retryCount,
		Error:         errorMsg,
	})

	return nil, fmt.Errorf("request failed after %d retries: %w", maxRetries, lastErr)
}

// exponentialBackoff calculates exponential backoff with jitter
func exponentialBackoff(attempt int) time.Duration {
	base := time.Second
	exponential := time.Duration(math.Pow(2, float64(attempt))) * base
	jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
	return exponential + jitter
}

func getHost(targetURL string) string {
	u, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}
	return u.Host
}

func getPath(targetURL string) string {
	u, err := url.Parse(targetURL)
	if err != nil {
		return ""
	}
	return u.Path
}

// redisCacheAdapter adapts RedisClientOptional to robots.Cache interface
type redisCacheAdapter struct {
	client RedisClientOptional
}

func (r *redisCacheAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	return r.client.Get(ctx, key)
}

func (r *redisCacheAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl)
}

