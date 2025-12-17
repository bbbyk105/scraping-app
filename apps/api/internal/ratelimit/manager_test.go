package ratelimit

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestManager_Wait(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	configs := map[string]RateLimitConfig{
		"test": {
			RPS:   10.0, // 10 requests per second
			Burst: 2,
		},
	}
	defaultConfig := RateLimitConfig{
		RPS:   1.0,
		Burst: 1,
	}

	manager := NewManager(configs, defaultConfig, logger)

	ctx := context.Background()

	// First request should pass immediately
	start := time.Now()
	err := manager.Wait(ctx, "test")
	if err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}
	elapsed := time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("First Wait() took too long: %v", elapsed)
	}

	// Second request should also pass (within burst)
	start = time.Now()
	err = manager.Wait(ctx, "test")
	if err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}
	elapsed = time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Second Wait() took too long: %v", elapsed)
	}

	// Third request should be rate limited (beyond burst)
	start = time.Now()
	err = manager.Wait(ctx, "test")
	if err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}
	elapsed = time.Since(start)
	// Should wait at least some time due to rate limiting
	// With RPS=10, each request should be ~100ms apart after burst
	if elapsed < 50*time.Millisecond {
		t.Errorf("Third Wait() should have been rate limited, but elapsed = %v", elapsed)
	}
}

func TestManager_DefaultConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	configs := map[string]RateLimitConfig{}
	defaultConfig := RateLimitConfig{
		RPS:   1.0,
		Burst: 1,
	}

	manager := NewManager(configs, defaultConfig, logger)

	ctx := context.Background()

	// Unknown provider should use default config
	err := manager.Wait(ctx, "unknown")
	if err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}
}

