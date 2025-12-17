package robots

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestChecker_CanFetch(t *testing.T) {
	// Test server that returns robots.txt
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/robots.txt" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`User-agent: *
Disallow: /admin/
Disallow: /private/

User-agent: PriceCompareBot
Allow: /products/
Disallow: /checkout/
`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	httpClient := &http.Client{Timeout: 5 * time.Second}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	checker := NewChecker(nil, 1*time.Hour, httpClient, logger)

	tests := []struct {
		name      string
		url       string
		userAgent string
		want      bool
		wantErr   bool
	}{
		{
			name:      "allowed path for wildcard user-agent",
			url:       server.URL + "/products/123",
			userAgent: "Mozilla/5.0",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "disallowed path for wildcard user-agent",
			url:       server.URL + "/admin/users",
			userAgent: "Mozilla/5.0",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "allowed path for specific user-agent",
			url:       server.URL + "/products/123",
			userAgent: "PriceCompareBot",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "disallowed path for specific user-agent",
			url:       server.URL + "/checkout/payment",
			userAgent: "PriceCompareBot",
			want:      false,
			wantErr:   false,
		},
		{
			name:      "root path allowed",
			url:       server.URL + "/",
			userAgent: "PriceCompareBot",
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, _, err := checker.CanFetch(context.Background(), tt.url, tt.userAgent)
			if (err != nil) != tt.wantErr {
				t.Errorf("CanFetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if allowed != tt.want {
				t.Errorf("CanFetch() allowed = %v, want %v", allowed, tt.want)
			}
		})
	}
}

func TestChecker_pathMatches(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	checker := NewChecker(nil, 1*time.Hour, &http.Client{}, logger)

	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		{"exact match", "/products", "/products", true},
		{"prefix match with slash", "/products/123", "/products/", true},
		{"prefix match without slash", "/products/123", "/products", true},
		{"no match", "/admin", "/products", false},
		{"empty pattern allows all", "/anything", "", false},
		{"root path", "/", "/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := checker.pathMatches(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("pathMatches(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.expected)
			}
		})
	}
}

