package body_limit

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

func TestBodyLimit_SkipOnPath(t *testing.T) {
	tests := []struct {
		name       string
		config     *Config
		path       string
		bodySize   int
		shouldSkip bool
	}{
		{
			name: "Skip upload paths",
			config: &Config{
				MaxSize:    100,
				SkipOnPath: []string{"/upload/*", "/api/files/*"},
			},
			path:       "/upload/image",
			bodySize:   200,
			shouldSkip: true,
		},
		{
			name: "Skip with wildcard pattern",
			config: &Config{
				MaxSize:    100,
				SkipOnPath: []string{"/api/*/upload"},
			},
			path:       "/api/v1/upload",
			bodySize:   200,
			shouldSkip: true,
		},
		{
			name: "Skip with double wildcard",
			config: &Config{
				MaxSize:    100,
				SkipOnPath: []string{"/static/**"},
			},
			path:       "/static/images/large.jpg",
			bodySize:   200,
			shouldSkip: true,
		},
		{
			name: "Don't skip normal paths",
			config: &Config{
				MaxSize:    100,
				SkipOnPath: []string{"/upload/*"},
			},
			path:       "/api/users",
			bodySize:   200,
			shouldSkip: false,
		},
		{
			name: "Exact path match",
			config: &Config{
				MaxSize:    100,
				SkipOnPath: []string{"/webhook"},
			},
			path:       "/webhook",
			bodySize:   200,
			shouldSkip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := BodyLimitMiddleware(tt.config)

			handlerCalled := false
			handler := func(ctx *request.Context) error {
				handlerCalled = true
				return nil
			}

			wrappedHandler := middleware(handler)

			// Create request with large body
			body := strings.Repeat("a", tt.bodySize)
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", tt.path, strings.NewReader(body))

			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			err := wrappedHandler(ctx)

			if tt.shouldSkip {
				// Should not return error and handler should be called
				if err != nil {
					t.Errorf("Expected no error when path should be skipped, got: %v", err)
				}
				if !handlerCalled {
					t.Error("Expected handler to be called when path should be skipped")
				}
			} else {
				// Should return error and handler should not be called
				if err == nil {
					t.Error("Expected error when path should not be skipped")
				}
				if handlerCalled {
					t.Error("Expected handler not to be called when limit exceeded")
				}
			}
		})
	}
}

func TestMatchPath(t *testing.T) {
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		// Exact matches
		{"/api/users", "/api/users", true},
		{"/upload", "/upload", true},

		// Wildcard matches
		{"/upload/image.jpg", "/upload/*", true},
		{"/api/v1/users", "/api/*/users", true},
		{"/static/css/style.css", "/static/*", false}, // * doesn't match subdirectories

		// Double wildcard matches
		{"/static/css/style.css", "/static/**", true},
		{"/api/v1/users/123/profile", "/api/**", true},
		{"/api/v1/users/123/profile", "/api/**/profile", true},

		// Non-matches
		{"/api/users", "/upload/*", false},
		{"/upload", "/upload/*", false}, // exact path doesn't match wildcard
		{"/api/v1/users", "/api/v2/*", false},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_vs_"+tt.pattern, func(t *testing.T) {
			result := matchPath(tt.path, tt.pattern)
			if result != tt.expected {
				t.Errorf("matchPath(%q, %q) = %v, expected %v",
					tt.path, tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestFactory(t *testing.T) {
	tests := []struct {
		name   string
		config any
		verify func(t *testing.T, middleware func(request.HandlerFunc) request.HandlerFunc)
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			verify: func(t *testing.T, middleware func(request.HandlerFunc) request.HandlerFunc) {
				// Should use default 10MB limit
				handler := func(ctx *request.Context) error { return nil }
				wrappedHandler := middleware(handler)

				// Test with 5MB body (should pass)
				body := strings.Repeat("a", 5*1024*1024)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/test", strings.NewReader(body))
				ctx, cancel := request.NewContext(w, r)
				defer cancel()

				err := wrappedHandler(ctx)
				if err != nil {
					t.Errorf("Expected no error with default config, got: %v", err)
				}
			},
		},
		{
			name: "map config",
			config: map[string]any{
				"max_size":     int64(1024),
				"status_code":  400,
				"message":      "Too big!",
				"skip_on_path": []string{"/upload/*"},
			},
			verify: func(t *testing.T, middleware func(request.HandlerFunc) request.HandlerFunc) {
				handler := func(ctx *request.Context) error { return nil }
				wrappedHandler := middleware(handler)

				// Test with body exceeding limit
				body := strings.Repeat("a", 2048)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/test", strings.NewReader(body))
				ctx, cancel := request.NewContext(w, r)
				defer cancel()

				err := wrappedHandler(ctx)
				if err == nil {
					t.Error("Expected error with large body")
				}
				if ctx.Response.StatusCode != 400 {
					t.Errorf("Expected status code 400, got %d", ctx.Response.StatusCode)
				}
				if ctx.Response.Message != "Too big!" {
					t.Errorf("Expected custom message, got %s", ctx.Response.Message)
				}
			},
		},
		{
			name: "struct config",
			config: Config{
				MaxSize:    512,
				StatusCode: 422,
				Message:    "Payload too large",
				SkipOnPath: []string{"/webhook"},
			},
			verify: func(t *testing.T, middleware func(request.HandlerFunc) request.HandlerFunc) {
				handler := func(ctx *request.Context) error { return nil }
				wrappedHandler := middleware(handler)

				// Test skip path
				body := strings.Repeat("a", 1024)
				w := httptest.NewRecorder()
				r := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
				ctx, cancel := request.NewContext(w, r)
				defer cancel()

				err := wrappedHandler(ctx)
				if err != nil {
					t.Errorf("Expected no error for skip path, got: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := factory(tt.config)
			tt.verify(t, middleware)
		})
	}
}
