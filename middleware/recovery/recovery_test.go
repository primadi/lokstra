package recovery

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
)

func TestRecovery(t *testing.T) {
	tests := []struct {
		name             string
		config           *Config
		panicValue       any
		expectStatus     int
		expectStackTrace bool
	}{
		{
			name: "recover from string panic",
			config: &Config{
				EnableStackTrace: false,
				EnableLogging:    false,
			},
			panicValue:       "something went wrong",
			expectStatus:     500,
			expectStackTrace: false,
		},
		{
			name: "recover with stack trace enabled (logged only)",
			config: &Config{
				EnableStackTrace: true,
				EnableLogging:    true, // Stack trace logged to console, not in response
			},
			panicValue:       "error with trace",
			expectStatus:     500,
			expectStackTrace: false, // Stack trace logged, not returned in response
		},
		{
			name: "recover from nil panic",
			config: &Config{
				EnableStackTrace: false,
				EnableLogging:    false,
			},
			panicValue:       nil,
			expectStatus:     500,
			expectStackTrace: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup formatter
			api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

			// Create router
			r := router.New("test-router")

			// Add recovery middleware
			r.Use(Middleware(tt.config))

			// Add handler that panics
			r.GET("/panic", func(c *request.Context) error {
				panic(tt.panicValue)
			})

			// Create request
			req := httptest.NewRequest("GET", "/panic", nil)

			// Record response
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Check status code
			if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, w.Code)
			}

			// Check response body
			body := w.Body.String()
			if body == "" {
				t.Error("Expected error response body, got empty")
			}

			// Check stack trace presence
			hasStackTrace := strings.Contains(body, "stack_trace")
			if hasStackTrace != tt.expectStackTrace {
				t.Errorf("Expected stack trace present=%v, got %v", tt.expectStackTrace, hasStackTrace)
			}

			t.Logf("Recovery response: %s", body)
		})
	}
}

func TestRecoveryWithCustomHandler(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	customHandlerCalled := false

	cfg := &Config{
		EnableStackTrace: false,
		EnableLogging:    false,
		CustomHandler: func(c *request.Context, recovered any, stack []byte) error {
			customHandlerCalled = true
			return c.Api.Ok(map[string]any{
				"custom_recovery": true,
				"panic_value":     recovered,
			})
		},
	}

	r := router.New("test-router")
	r.Use(Middleware(cfg))

	r.GET("/panic", func(c *request.Context) error {
		panic("custom handled panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if !customHandlerCalled {
		t.Error("Expected custom handler to be called")
	}

	body := w.Body.String()
	if !strings.Contains(body, "custom_recovery") {
		t.Errorf("Expected custom recovery response, got: %s", body)
	}
}

func TestRecoveryDoesNotAffectNormalRequests(t *testing.T) {
	// Setup formatter
	api_formatter.SetGlobalFormatter(api_formatter.NewApiResponseFormatter())

	r := router.New("test-router")
	r.Use(Middleware(&Config{
		EnableStackTrace: false,
		EnableLogging:    false,
	}))

	r.GET("/normal", func(c *request.Context) error {
		return c.Api.Ok("success")
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "success") {
		t.Errorf("Expected normal response, got: %s", body)
	}
}

func TestRecoveryFactory(t *testing.T) {
	// Test with nil params
	middleware1 := MiddlewareFactory(nil)
	if middleware1 == nil {
		t.Error("Expected middleware with nil params")
	}

	// Test with custom params
	params := map[string]any{
		PARAMS_ENABLE_STACK_TRACE: true,
		PARAMS_ENABLE_LOGGING:     false,
	}
	middleware2 := MiddlewareFactory(params)
	if middleware2 == nil {
		t.Error("Expected middleware with custom params")
	}
}
