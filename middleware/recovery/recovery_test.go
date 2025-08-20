package recovery

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

// mockLogger implements serviceapi.Logger for testing
type mockLogger struct{}

func (m *mockLogger) Debugf(msg string, v ...any)                              {}
func (m *mockLogger) Infof(msg string, v ...any)                               {}
func (m *mockLogger) Warnf(msg string, v ...any)                               {}
func (m *mockLogger) Errorf(msg string, v ...any)                              {}
func (m *mockLogger) Fatalf(msg string, v ...any)                              {}
func (m *mockLogger) GetLogLevel() serviceapi.LogLevel                         { return serviceapi.LogLevelInfo }
func (m *mockLogger) SetLogLevel(level serviceapi.LogLevel)                    {}
func (m *mockLogger) WithField(key string, value any) serviceapi.Logger        { return m }
func (m *mockLogger) WithFields(fields serviceapi.LogFields) serviceapi.Logger { return m }
func (m *mockLogger) SetFormat(format string)                                  {}
func (m *mockLogger) SetOutput(output string)                                  {}

func setupLogger() {
	// Set a mock logger for testing
	if lokstra.Logger == nil {
		lokstra.Logger = &mockLogger{}
	}
}

func TestConfig_Parsing(t *testing.T) {
	testCases := []struct {
		name     string
		config   any
		expected bool
	}{
		{
			name:     "nil config - default to true",
			config:   nil,
			expected: true,
		},
		{
			name: "map config with enable_stack_trace true",
			config: map[string]any{
				"enable_stack_trace": true,
			},
			expected: true,
		},
		{
			name: "map config with enable_stack_trace false",
			config: map[string]any{
				"enable_stack_trace": false,
			},
			expected: false,
		},
		{
			name: "struct config",
			config: &Config{
				EnableStackTrace: false,
			},
			expected: false,
		},
		{
			name: "struct value config",
			config: Config{
				EnableStackTrace: true,
			},
			expected: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			middleware := factory(tt.config)
			if middleware == nil {
				t.Error("Expected middleware to be created")
			}
			// Note: We can't easily test the actual EnableStackTrace value
			// without exposing it, but we can test that the factory accepts the config
		})
	}
}

func TestRecoveryMiddleware_Module(t *testing.T) {
	module := GetModule()

	if module.Name() != NAME {
		t.Errorf("Expected module name to be '%s', got '%s'", NAME, module.Name())
	}

	description := module.Description()
	if description == "" {
		t.Error("Expected non-empty description")
	}

	if !strings.Contains(description, "panic") {
		t.Error("Expected description to mention panic recovery")
	}
}

func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	// Setup logger for test
	setupLogger()

	// Create middleware with stack trace enabled
	middleware := factory(map[string]any{
		"enable_stack_trace": true,
	})

	// Create a handler that panics
	panicHandler := func(ctx *lokstra.Context) error {
		panic("test panic message")
	}

	// Wrap the handler with recovery middleware
	wrappedHandler := middleware(panicHandler)

	// Create test context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)

	// Should not return an error (panic should be recovered)
	if err != nil {
		t.Errorf("Expected no error returned from recovered panic, got: %v", err)
	}

	// Should set internal server error status
	if ctx.Response.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", ctx.Response.StatusCode)
	}

	// Response should indicate internal server error
	if ctx.Response.Message != "Internal Server Error" {
		t.Errorf("Expected 'Internal Server Error' message, got '%s'", ctx.Response.Message)
	}

	// Success should be false
	if ctx.Response.Success != false {
		t.Errorf("Expected Success to be false, got %v", ctx.Response.Success)
	}
}

func TestRecoveryMiddleware_NormalExecution(t *testing.T) {
	// Create middleware
	middleware := factory(nil)

	// Create a normal handler that doesn't panic
	normalHandler := func(ctx *lokstra.Context) error {
		ctx.Response.StatusCode = 200
		ctx.Response.Message = "Success"
		return nil
	}

	// Wrap the handler with recovery middleware
	wrappedHandler := middleware(normalHandler)

	// Create test context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)

	// Should work normally
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if ctx.Response.StatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", ctx.Response.StatusCode)
	}

	if ctx.Response.Message != "Success" {
		t.Errorf("Expected 'Success' message, got '%s'", ctx.Response.Message)
	}
}

func TestRecoveryMiddleware_DisabledStackTrace(t *testing.T) {
	// Create middleware with stack trace disabled
	middleware := factory(map[string]any{
		"enable_stack_trace": false,
	})

	// Create a handler that panics
	panicHandler := func(ctx *lokstra.Context) error {
		panic("test panic without stack trace")
	}

	// Wrap the handler with recovery middleware
	wrappedHandler := middleware(panicHandler)

	// Create test context
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)

	// Should not return an error (panic should be recovered)
	if err != nil {
		t.Errorf("Expected no error returned from recovered panic, got: %v", err)
	}

	// Should still set internal server error status
	if ctx.Response.StatusCode != 500 {
		t.Errorf("Expected status code 500, got %d", ctx.Response.StatusCode)
	}

	// Note: We can't easily test that stack trace is not logged without
	// intercepting the logger, but we can verify the middleware still works
}

func TestRecoveryMiddleware_ConfigEdgeCases(t *testing.T) {
	// Test with invalid config type
	middleware := factory("invalid config")
	if middleware == nil {
		t.Error("Expected middleware to be created even with invalid config")
	}

	// Test with map containing wrong type
	middleware2 := factory(map[string]any{
		"enable_stack_trace": "not a boolean",
	})
	if middleware2 == nil {
		t.Error("Expected middleware to be created even with wrong type in config")
	}

	// Test with empty map
	middleware3 := factory(map[string]any{})
	if middleware3 == nil {
		t.Error("Expected middleware to be created with empty config map")
	}
}
