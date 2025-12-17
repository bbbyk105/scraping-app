package audit

import (
	"log/slog"
	"time"
)

// Entry represents an audit log entry
type Entry struct {
	Timestamp     time.Time `json:"ts"`
	Provider      string    `json:"provider"`
	Method        string    `json:"method"`
	URL           string    `json:"url"`
	Host          string    `json:"host"`
	Path          string    `json:"path"`
	Status        int       `json:"status"`
	DurationMs    int64     `json:"duration_ms"`
	UserAgent     string    `json:"user_agent"`
	RobotsAllowed bool      `json:"robots_allowed"`
	RobotsGroup   string    `json:"robots_group,omitempty"`
	RetryCount    int       `json:"retry_count"`
	Error         string    `json:"error,omitempty"`
}

// LogRequest logs an HTTP request to audit log
func LogRequest(logger *slog.Logger, entry Entry) {
	attrs := []any{
		slog.Time("ts", entry.Timestamp),
		slog.String("provider", entry.Provider),
		slog.String("method", entry.Method),
		slog.String("url", entry.URL),
		slog.String("host", entry.Host),
		slog.String("path", entry.Path),
		slog.Int("status", entry.Status),
		slog.Int64("duration_ms", entry.DurationMs),
		slog.String("user_agent", entry.UserAgent),
		slog.Bool("robots_allowed", entry.RobotsAllowed),
		slog.Int("retry_count", entry.RetryCount),
	}

	if entry.RobotsGroup != "" {
		attrs = append(attrs, slog.String("robots_group", entry.RobotsGroup))
	}

	if entry.Error != "" {
		attrs = append(attrs, slog.String("error", entry.Error))
	}

	logger.Info("HTTP request audit", attrs...)
}

