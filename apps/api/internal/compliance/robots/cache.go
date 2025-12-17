package robots

import (
	"context"
	"time"
)

// Cache interface for robots.txt caching
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

