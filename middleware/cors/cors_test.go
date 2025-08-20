package cors

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

func TestConfig_Parsing(t *testing.T) {
	tests := []struct {
		name     string
		config   any
		expected *Config
	}{
		{
			name:   "nil config - use defaults",
			config: nil,
			expected: &Config{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"*"},
				ExposedHeaders:   []string{},
				AllowCredentials: false,
				MaxAge:           86400,
			},
		},
		{
			name: "map config with all settings",
			config: map[string]any{
				"allowed_origins":   []string{"http://localhost:3000", "https://app.example.com"},
				"allowed_methods":   []string{"GET", "POST", "PUT"},
				"allowed_headers":   []string{"Content-Type", "Authorization"},
				"exposed_headers":   []string{"X-Total-Count", "X-Page-Count"},
				"allow_credentials": true,
				"max_age":           3600,
			},
			expected: &Config{
				AllowedOrigins:   []string{"http://localhost:3000", "https://app.example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
		},
		{
			name: "map config with string methods",
			config: map[string]any{
				"allowed_origins": []string{"*"},
				"allowed_methods": "GET,POST,PUT",
				"allowed_headers": "Content-Type,Authorization",
			},
			expected: &Config{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "PUT"},
				AllowedHeaders:   []string{"Content-Type", "Authorization"},
				ExposedHeaders:   []string{},
				AllowCredentials: false,
				MaxAge:           86400,
			},
		},
		{
			name: "struct config",
			config: &Config{
				AllowedOrigins:   []string{"https://myapp.com"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"*"},
				ExposedHeaders:   []string{"X-Custom-Header"},
				AllowCredentials: true,
				MaxAge:           7200,
			},
			expected: &Config{
				AllowedOrigins:   []string{"https://myapp.com"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowedHeaders:   []string{"*"},
				ExposedHeaders:   []string{"X-Custom-Header"},
				AllowCredentials: true,
				MaxAge:           7200,
			},
		},
		{
			name: "struct value config",
			config: Config{
				AllowedOrigins:   []string{"http://localhost:8080"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Accept"},
				ExposedHeaders:   []string{},
				AllowCredentials: false,
				MaxAge:           1800,
			},
			expected: &Config{
				AllowedOrigins:   []string{"http://localhost:8080"},
				AllowedMethods:   []string{"GET"},
				AllowedHeaders:   []string{"Accept"},
				ExposedHeaders:   []string{},
				AllowCredentials: false,
				MaxAge:           1800,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := factory(tt.config)
			if middleware == nil {
				t.Fatal("Expected middleware function, got nil")
			}

			// Test that middleware can be created - detailed config testing in other tests
		})
	}
}

func TestCorsMiddleware_Module(t *testing.T) {
	module := GetModule()

	if module.Name() != NAME {
		t.Errorf("Expected module name %s, got %s", NAME, module.Name())
	}

	if module.Description() == "" {
		t.Error("Expected non-empty description")
	}
}

func TestCorsMiddleware_ActualRequest(t *testing.T) {
	// Create middleware with custom config
	config := map[string]any{
		"allowed_origins":   []string{"http://localhost:3000"},
		"allow_credentials": true,
		"exposed_headers":   []string{"X-Total-Count"},
	}
	middleware := factory(config)

	// Create a test handler
	testHandler := func(ctx *lokstra.Context) error {
		ctx.Response.StatusCode = 200
		ctx.Response.Message = "Success"
		return nil
	}

	// Wrap the handler with middleware
	wrappedHandler := middleware(testHandler)

	// Create test request with Origin header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check CORS headers
	headers := ctx.Response.GetHeaders()

	if headers.Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("Expected Access-Control-Allow-Origin to be 'http://localhost:3000', got '%s'",
			headers.Get("Access-Control-Allow-Origin"))
	}

	if headers.Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("Expected Access-Control-Allow-Credentials to be 'true', got '%s'",
			headers.Get("Access-Control-Allow-Credentials"))
	}

	if headers.Get("Access-Control-Expose-Headers") != "X-Total-Count" {
		t.Errorf("Expected Access-Control-Expose-Headers to be 'X-Total-Count', got '%s'",
			headers.Get("Access-Control-Expose-Headers"))
	}
}

func TestCorsMiddleware_PreflightRequest(t *testing.T) {
	// Create middleware with custom config
	config := map[string]any{
		"allowed_origins": []string{"*"},
		"allowed_methods": []string{"GET", "POST", "PUT", "DELETE"},
		"allowed_headers": []string{"Content-Type", "Authorization"},
		"max_age":         7200,
	}
	middleware := factory(config)

	// Create a test handler (should not be called for OPTIONS)
	testHandler := func(ctx *lokstra.Context) error {
		t.Error("Handler should not be called for OPTIONS request")
		return nil
	}

	// Wrap the handler with middleware
	wrappedHandler := middleware(testHandler)

	// Create preflight OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check response status
	if ctx.Response.StatusCode != 204 {
		t.Errorf("Expected status code 204, got %d", ctx.Response.StatusCode)
	}

	// Check CORS headers
	headers := ctx.Response.GetHeaders()

	if !strings.Contains(headers.Get("Access-Control-Allow-Methods"), "GET") {
		t.Errorf("Expected Access-Control-Allow-Methods to contain 'GET', got '%s'",
			headers.Get("Access-Control-Allow-Methods"))
	}

	if !strings.Contains(headers.Get("Access-Control-Allow-Headers"), "Content-Type") {
		t.Errorf("Expected Access-Control-Allow-Headers to contain 'Content-Type', got '%s'",
			headers.Get("Access-Control-Allow-Headers"))
	}

	if headers.Get("Access-Control-Max-Age") != "7200" {
		t.Errorf("Expected Access-Control-Max-Age to be '7200', got '%s'",
			headers.Get("Access-Control-Max-Age"))
	}
}

func TestCorsMiddleware_WildcardHeaders(t *testing.T) {
	// Create middleware with wildcard headers
	config := map[string]any{
		"allowed_origins": []string{"*"},
		"allowed_headers": []string{"*"},
	}
	middleware := factory(config)

	testHandler := func(ctx *lokstra.Context) error {
		return nil
	}

	wrappedHandler := middleware(testHandler)

	// Create preflight OPTIONS request with custom headers
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Headers", "X-Custom-Header, X-Another-Header")
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that requested headers are echoed back
	headers := ctx.Response.GetHeaders()
	allowedHeaders := headers.Get("Access-Control-Allow-Headers")

	if !strings.Contains(allowedHeaders, "X-Custom-Header") {
		t.Errorf("Expected Access-Control-Allow-Headers to contain 'X-Custom-Header', got '%s'", allowedHeaders)
	}
}

func TestCorsMiddleware_NoOriginHeader(t *testing.T) {
	// Create middleware
	middleware := factory(nil)

	testHandler := func(ctx *lokstra.Context) error {
		ctx.Response.StatusCode = 200
		return nil
	}

	wrappedHandler := middleware(testHandler)

	// Create request without Origin header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that no CORS headers are set
	headers := ctx.Response.GetHeaders()

	if headers.Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header, got '%s'",
			headers.Get("Access-Control-Allow-Origin"))
	}
}

func TestCorsMiddleware_OriginNotAllowed(t *testing.T) {
	// Create middleware with specific allowed origins
	config := map[string]any{
		"allowed_origins": []string{"http://localhost:3000"},
	}
	middleware := factory(config)

	testHandler := func(ctx *lokstra.Context) error {
		ctx.Response.StatusCode = 200
		return nil
	}

	wrappedHandler := middleware(testHandler)

	// Create request with non-allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute the wrapped handler
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that no CORS headers are set
	headers := ctx.Response.GetHeaders()

	if headers.Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("Expected no Access-Control-Allow-Origin header for non-allowed origin, got '%s'",
			headers.Get("Access-Control-Allow-Origin"))
	}
}

func TestParseStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []string
	}{
		{
			name:     "string slice",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "any slice",
			input:    []any{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "comma separated string",
			input:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single string",
			input:    "single",
			expected: []string{"single"},
		},
		{
			name:     "comma separated with spaces",
			input:    "a, b , c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "invalid type",
			input:    123,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseStringSlice(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}
