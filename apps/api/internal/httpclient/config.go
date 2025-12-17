package httpclient

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds HTTP client configuration
type Config struct {
	AllowLiveFetch      bool
	UserAgent           string
	RobotsCacheTTLHours int
	ProviderRateLimits  map[string]RateLimitConfig
	DefaultRateLimit    RateLimitConfig
	HTTPTimeoutSeconds  int
	HTTPMaxRetries      int
}

// RateLimitConfig holds rate limit configuration for a provider
type RateLimitConfig struct {
	RPS   float64
	Burst int
}

// LoadConfig loads HTTP client configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		AllowLiveFetch:      getBoolEnv("ALLOW_LIVE_FETCH", false),
		UserAgent:           getEnv("USER_AGENT", "PriceCompareBot/1.0 (+contact@example.com)"),
		RobotsCacheTTLHours: getIntEnv("ROBOTS_CACHE_TTL_HOURS", 24),
		HTTPTimeoutSeconds:  getIntEnv("HTTP_TIMEOUT_SECONDS", 10),
		HTTPMaxRetries:      getIntEnv("HTTP_MAX_RETRIES", 3),
		ProviderRateLimits:  make(map[string]RateLimitConfig),
	}

	// Load provider-specific rate limits
	cfg.ProviderRateLimits["demo"] = RateLimitConfig{
		RPS:   getFloatEnv("PROVIDER_RATE_LIMIT_DEMO_RPS", 10),
		Burst: getIntEnv("PROVIDER_RATE_LIMIT_BURST", 2),
	}
	cfg.ProviderRateLimits["public_html"] = RateLimitConfig{
		RPS:   getFloatEnv("PROVIDER_RATE_LIMIT_PUBLIC_HTML_RPS", 10),
		Burst: getIntEnv("PROVIDER_RATE_LIMIT_BURST", 2),
	}
	cfg.ProviderRateLimits["live"] = RateLimitConfig{
		RPS:   getFloatEnv("PROVIDER_RATE_LIMIT_LIVE_RPS", 1),
		Burst: getIntEnv("PROVIDER_RATE_LIMIT_BURST", 2),
	}
	cfg.ProviderRateLimits["walmart"] = RateLimitConfig{
		RPS:   getFloatEnv("PROVIDER_RATE_LIMIT_WALMART_RPS", 5),
		Burst: getIntEnv("PROVIDER_RATE_LIMIT_BURST", 10),
	}
	cfg.ProviderRateLimits["amazon"] = RateLimitConfig{
		RPS:   getFloatEnv("PROVIDER_RATE_LIMIT_AMAZON_RPS", 1),
		Burst: getIntEnv("PROVIDER_RATE_LIMIT_BURST", 2),
	}

	// Default rate limit (fallback)
	cfg.DefaultRateLimit = RateLimitConfig{
		RPS:   getFloatEnv("PROVIDER_RATE_LIMIT_LIVE_RPS", 1),
		Burst: getIntEnv("PROVIDER_RATE_LIMIT_BURST", 2),
	}

	return cfg
}

// IsExternalURL checks if a URL is external (http/https with a host)
func IsExternalURL(targetURL string) (bool, error) {
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}

	// Must be http or https
	if u.Scheme != "http" && u.Scheme != "https" {
		return false, nil
	}

	// Must have a host
	if u.Host == "" {
		return false, nil
	}

	// Exclude localhost and internal addresses
	host := strings.ToLower(u.Host)
	if strings.Contains(host, "localhost") ||
		strings.Contains(host, "127.0.0.1") ||
		strings.Contains(host, "::1") ||
		strings.HasPrefix(host, "192.168.") ||
		strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.16.") ||
		strings.HasPrefix(host, "172.17.") ||
		strings.HasPrefix(host, "172.18.") ||
		strings.HasPrefix(host, "172.19.") ||
		strings.HasPrefix(host, "172.20.") ||
		strings.HasPrefix(host, "172.21.") ||
		strings.HasPrefix(host, "172.22.") ||
		strings.HasPrefix(host, "172.23.") ||
		strings.HasPrefix(host, "172.24.") ||
		strings.HasPrefix(host, "172.25.") ||
		strings.HasPrefix(host, "172.26.") ||
		strings.HasPrefix(host, "172.27.") ||
		strings.HasPrefix(host, "172.28.") ||
		strings.HasPrefix(host, "172.29.") ||
		strings.HasPrefix(host, "172.30.") ||
		strings.HasPrefix(host, "172.31.") {
		return false, nil
	}

	return true, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func getFloatEnv(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return floatValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes"
}

