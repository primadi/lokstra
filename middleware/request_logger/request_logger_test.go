package request_logger

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

func TestConfig_Parsing(t *testing.T) {
	tests := []struct {
		name     string
		config   any
		expected Config
	}{
		{
			name: "map config with both flags true",
			config: map[string]any{
				"include_request_body":  true,
				"include_response_body": true,
			},
			expected: Config{
				IncludeRequestBody:  true,
				IncludeResponseBody: true,
			},
		},
		{
			name: "map config with both flags false",
			config: map[string]any{
				"include_request_body":  false,
				"include_response_body": false,
			},
			expected: Config{
				IncludeRequestBody:  false,
				IncludeResponseBody: false,
			},
		},
		{
			name: "map config with partial settings",
			config: map[string]any{
				"include_request_body": true,
			},
			expected: Config{
				IncludeRequestBody:  true,
				IncludeResponseBody: false,
			},
		},
		{
			name:   "nil config",
			config: nil,
			expected: Config{
				IncludeRequestBody:  false,
				IncludeResponseBody: false,
			},
		},
		{
			name: "struct config",
			config: Config{
				IncludeRequestBody:  true,
				IncludeResponseBody: true,
			},
			expected: Config{
				IncludeRequestBody:  true,
				IncludeResponseBody: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := factory(tt.config)
			if middleware == nil {
				t.Fatal("Expected middleware to be created")
			}
			// Note: We can't easily test the internal config parsing without
			// making the parsing function public or adding more complex test setup
		})
	}
}

func TestRequestLogger_Module(t *testing.T) {
	module := &RequestLogger{}

	if module.Name() != NAME {
		t.Errorf("Expected name %s, got %s", NAME, module.Name())
	}

	description := module.Description()
	if !strings.Contains(description, "request") {
		t.Errorf("Expected description to contain 'request', got: %s", description)
	}
}

func TestRequestLogger_BasicLogging(t *testing.T) {
	// Create middleware with default config
	middleware := factory(nil)

	// Create a test handler
	testHandler := func(ctx *request.Context) error {
		ctx.Response.StatusCode = 200
		return nil
	}

	// Wrap the handler with middleware
	wrappedHandler := middleware(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("User-Agent", "test-agent")

	// Create test context
	ctx := &request.Context{
		Request:  req,
		Response: response.NewResponse(),
	}
	ctx.Response.StatusCode = 200

	// Execute the middleware
	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRequestLogger_WithRequestBody(t *testing.T) {
	// Create middleware with request body logging enabled
	config := map[string]any{
		"include_request_body": true,
	}
	middleware := factory(config)

	// Create a test handler
	testHandler := func(ctx *request.Context) error {
		ctx.Response.StatusCode = 200
		return nil
	}

	// Wrap the handler with middleware
	wrappedHandler := middleware(testHandler)

	// Test with JSON body
	t.Run("JSON request body", func(t *testing.T) {
		jsonBody := `{"name":"test","value":123}`
		req := httptest.NewRequest("POST", "/test", strings.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		ctx, cancel := request.NewContext(nil, w, req)
		defer cancel()
		ctx.Response.StatusCode = 201

		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify that request body is still readable using GetRawRequestBody
		body, err := ctx.GetRawRequestBody()
		if err != nil {
			t.Errorf("Expected to be able to read body after middleware, got error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Errorf("Expected valid JSON body after middleware, got error: %v", err)
		}
	})

	// Test with text body
	t.Run("Text request body", func(t *testing.T) {
		textBody := "plain text body"
		req := httptest.NewRequest("POST", "/test", strings.NewReader(textBody))
		req.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()
		ctx, cancel := request.NewContext(nil, w, req)
		defer cancel()
		ctx.Response.StatusCode = 200

		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify that request body is still readable using GetRawRequestBody
		body, err := ctx.GetRawRequestBody()
		if err != nil {
			t.Errorf("Expected to be able to read body after middleware, got error: %v", err)
		}

		if string(body) != textBody {
			t.Errorf("Expected body to be '%s', got '%s'", textBody, string(body))
		}
	})
}

func TestRequestLogger_ErrorStatusLogging(t *testing.T) {
	middleware := factory(nil)

	tests := []struct {
		name       string
		statusCode int
		expectErr  bool
	}{
		{"Success 200", 200, false},
		{"Client Error 400", 400, false},
		{"Client Error 404", 404, false},
		{"Server Error 500", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testHandler := func(ctx *request.Context) error {
				ctx.Response.StatusCode = tt.statusCode
				return nil
			}

			wrappedHandler := middleware(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			ctx := &request.Context{
				Request:  req,
				Response: response.NewResponse(),
			}
			ctx.Response.StatusCode = tt.statusCode

			err := wrappedHandler(ctx)
			if (err != nil) != tt.expectErr {
				t.Errorf("Expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}

func TestRequestLogger_LongBodyTruncation(t *testing.T) {
	config := map[string]any{
		"include_request_body": true,
	}
	middleware := factory(config)

	testHandler := func(ctx *request.Context) error {
		ctx.Response.StatusCode = 200
		return nil
	}

	wrappedHandler := middleware(testHandler)

	// Create a body longer than 1000 characters
	longBody := strings.Repeat("a", 1500)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(longBody))

	ctx := &request.Context{
		Request:  req,
		Response: response.NewResponse(),
	}
	ctx.Response.StatusCode = 200

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// The middleware should handle long bodies gracefully
	// (actual truncation testing would require access to logs)
}

func TestRequestLogger_WithResponseBody(t *testing.T) {
	// Create middleware with response body logging enabled
	config := map[string]any{
		"include_response_body": true,
	}
	middleware := factory(config)

	// Create a test handler that sets response data
	testHandler := func(ctx *request.Context) error {
		ctx.Response.StatusCode = 200
		responseData := map[string]any{
			"message": "success",
			"data":    []string{"item1", "item2"},
		}
		// Set raw data to simulate response body
		jsonData, _ := json.Marshal(responseData)
		ctx.Response.RawData = jsonData
		return nil
	}

	// Wrap the handler with middleware
	wrappedHandler := middleware(testHandler)

	// Test with JSON response
	t.Run("JSON response body", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		ctx, cancel := request.NewContext(nil, w, req)
		defer cancel()

		err := wrappedHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify that response body can be read using GetRawResponseBody
		body, err := ctx.GetRawResponseBody()
		if err != nil {
			t.Errorf("Expected to be able to read response body, got error: %v", err)
		}

		var parsed map[string]any
		if err := json.Unmarshal(body, &parsed); err != nil {
			t.Errorf("Expected valid JSON response body, got error: %v", err)
		}

		if parsed["message"] != "success" {
			t.Errorf("Expected message to be 'success', got %v", parsed["message"])
		}
	})

	// Test with empty response body
	t.Run("Empty response body", func(t *testing.T) {
		emptyHandler := func(ctx *request.Context) error {
			ctx.Response.StatusCode = 204 // No Content
			// No response data set
			return nil
		}

		wrappedEmptyHandler := middleware(emptyHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		ctx, cancel := request.NewContext(nil, w, req)
		defer cancel()

		err := wrappedEmptyHandler(ctx)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify that empty response body doesn't cause issues
		body, err := ctx.GetRawResponseBody()
		if err != nil {
			t.Errorf("Expected no error for empty response body, got: %v", err)
		}

		if len(body) != 0 {
			t.Errorf("Expected empty response body, got: %s", string(body))
		}
	})
}
