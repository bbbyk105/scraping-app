package ratelimit

import (
	"context"
	"log/slog"
	"sync"

	"golang.org/x/time/rate"
)

// Manager manages rate limiters per provider
type Manager struct {
	limiters map[string]*rate.Limiter
	configs  map[string]RateLimitConfig
	defaultConfig RateLimitConfig
	mu       sync.RWMutex
	logger   *slog.Logger
}

// RateLimitConfig holds rate limit configuration
type RateLimitConfig struct {
	RPS   float64
	Burst int
}

// NewManager creates a new rate limit manager
func NewManager(configs map[string]RateLimitConfig, defaultConfig RateLimitConfig, logger *slog.Logger) *Manager {
	return &Manager{
		limiters:      make(map[string]*rate.Limiter),
		configs:       configs,
		defaultConfig: defaultConfig,
		logger:        logger,
	}
}

// Wait waits for rate limit token for the given provider
func (m *Manager) Wait(ctx context.Context, providerKey string) error {
	limiter := m.getLimiter(providerKey)
	return limiter.Wait(ctx)
}

// getLimiter gets or creates a limiter for the provider
func (m *Manager) getLimiter(providerKey string) *rate.Limiter {
	m.mu.RLock()
	limiter, ok := m.limiters[providerKey]
	m.mu.RUnlock()

	if ok {
		return limiter
	}

	// Create new limiter
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok := m.limiters[providerKey]; ok {
		return limiter
	}

	// Get config for provider or use default
	config, ok := m.configs[providerKey]
	if !ok {
		config = m.defaultConfig
	}

	// Create limiter: rate.Every converts RPS to interval
	limiter = rate.NewLimiter(rate.Limit(config.RPS), config.Burst)
	m.limiters[providerKey] = limiter

	return limiter
}

