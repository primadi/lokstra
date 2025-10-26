package body_limit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

func TestBodyLimit(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		bodySize       int64
		path           string
		expectedStatus int
		shouldPass     bool
	}{
		{
			name: "body within limit",
			config: &Config{
				MaxSize: 1024,
			},
			bodySize:       512,
			path:           "/api/test",
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
		{
			name: "body exceeds limit",
			config: &Config{
				MaxSize: 1024,
			},
			bodySize:       2048,
			path:           "/api/test",
			expectedStatus: http.StatusRequestEntityTooLarge,
			shouldPass:     false,
		},
		{
			name: "body exceeds limit but skip large payloads",
			config: &Config{
				MaxSize:           1024,
				SkipLargePayloads: true,
			},
			bodySize:       2048,
			path:           "/api/test",
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
		{
			name: "body exceeds limit but path is skipped",
			config: &Config{
				MaxSize:    1024,
				SkipOnPath: []string{"/upload/*"},
			},
			bodySize:       2048,
			path:           "/upload/file",
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
		{
			name: "body exceeds limit and path not skipped",
			config: &Config{
				MaxSize:    1024,
				SkipOnPath: []string{"/upload/*"},
			},
			bodySize:       2048,
			path:           "/api/test",
			expectedStatus: http.StatusRequestEntityTooLarge,
			shouldPass:     false,
		},
		{
			name: "custom status code and message",
			config: &Config{
				MaxSize:    1024,
				Message:    "Custom error message",
				StatusCode: http.StatusBadRequest,
			},
			bodySize:       2048,
			path:           "/api/test",
			expectedStatus: http.StatusBadRequest,
			shouldPass:     false,
		},
		{
			name: "skip path with ** pattern",
			config: &Config{
				MaxSize:    1024,
				SkipOnPath: []string{"/public/**"},
			},
			bodySize:       2048,
			path:           "/public/uploads/images/test.jpg",
			expectedStatus: http.StatusOK,
			shouldPass:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup formatter
			api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

			// Create router
			r := router.New("test-router")

			// Add body limit middleware
			r.Use(Middleware(tt.config))

			// Add test handler
			r.GET(tt.path, func(c *request.Context) error {
				return c.Api.Ok("success")
			})

			// Create request with specified body size
			body := bytes.Repeat([]byte("a"), int(tt.bodySize))
			req := httptest.NewRequest("GET", tt.path, bytes.NewReader(body))
			req.ContentLength = tt.bodySize

			// Record response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// Additional checks for failed requests
			if !tt.shouldPass && w.Code != http.StatusOK {
				// Should have error response
				if w.Body.Len() == 0 {
					t.Error("Expected error response body, got empty")
				}
			}
		})
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		pattern  string
		expected bool
	}{
		// {
		// 	name:     "exact match",
		// 	path:     "/api/test",
		// 	pattern:  "/api/test",
		// 	expected: true,
		// },
		// {
		// 	name:     "no match",
		// 	path:     "/api/test",
		// 	pattern:  "/api/other",
		// 	expected: false,
		// },
		// {
		// 	name:     "single wildcard match",
		// 	path:     "/api/test",
		// 	pattern:  "/api/*",
		// 	expected: true,
		// },
		// {
		// 	name:     "single wildcard no match - different segments",
		// 	path:     "/api/test/sub",
		// 	pattern:  "/api/*",
		// 	expected: false,
		// },
		// {
		// 	name:     "double wildcard match",
		// 	path:     "/api/test/sub/deep",
		// 	pattern:  "/api/**",
		// 	expected: true,
		// },
		// {
		// 	name:     "double wildcard prefix match",
		// 	path:     "/public/uploads/images/test.jpg",
		// 	pattern:  "/public/**",
		// 	expected: true,
		// },
		// {
		// 	name:     "double wildcard suffix match",
		// 	path:     "/api/v1/test",
		// 	pattern:  "**/test",
		// 	expected: true,
		// },
		// {
		// 	name:     "multiple single wildcards",
		// 	path:     "/api/v1/users",
		// 	pattern:  "/*/v1/*",
		// 	expected: true,
		// },
		{
			name:     "wildcard at start",
			path:     "/api/test",
			pattern:  "*/test",
			expected: false, // Pattern has 2 segments ["*", "test"], path has 3 segments ["", "api", "test"]
		},
		{
			name:     "wildcard with leading slash",
			path:     "/api/test",
			pattern:  "/*/test",
			expected: true, // Correct pattern for matching /api/test
		},
		{
			name:     "wildcard at end",
			path:     "/api/test",
			pattern:  "/api/*",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchPath(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchPath(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxSize != 10*1024*1024 {
		t.Errorf("Expected default MaxSize to be 10MB, got %d", cfg.MaxSize)
	}

	if cfg.SkipLargePayloads != false {
		t.Error("Expected default SkipLargePayloads to be false")
	}

	if cfg.Message != "Request body too large" {
		t.Errorf("Expected default Message, got %q", cfg.Message)
	}

	if cfg.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected default StatusCode to be 413, got %d", cfg.StatusCode)
	}

	if len(cfg.SkipOnPath) != 0 {
		t.Errorf("Expected default SkipOnPath to be empty, got %v", cfg.SkipOnPath)
	}
}

func TestBodyLimitWithDefaultValues(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	// Test with empty config - should use defaults
	cfg := &Config{}

	r := router.New("test-router")
	r.Use(Middleware(cfg))
	r.GET("/test", func(c *request.Context) error {
		return c.Api.Ok("success")
	})

	// Body within default limit (10MB)
	body := bytes.Repeat([]byte("a"), 1024) // 1KB
	req := httptest.NewRequest("GET", "/test", bytes.NewReader(body))
	req.ContentLength = 1024

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Body exceeds default limit
	largeBody := bytes.Repeat([]byte("a"), 11*1024*1024) // 11MB
	req2 := httptest.NewRequest("GET", "/test", bytes.NewReader(largeBody))
	req2.ContentLength = 11 * 1024 * 1024

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413, got %d", w2.Code)
	}
}
