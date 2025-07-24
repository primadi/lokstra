package request_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

func TestNewContext(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?param=value", nil)
	r.Header.Set("X-Test-Header", "test-value")

	// Test
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Assertions
	if ctx == nil {
		t.Fatal("Expected context to be created")
	}

	if ctx.Writer != w {
		t.Error("Expected Writer to be set correctly")
	}

	if ctx.Request == nil {
		t.Error("Expected Request to be set")
	}

	if ctx.Context == nil {
		t.Error("Expected Context to be set")
	}

	if ctx.Response == nil {
		t.Error("Expected Response to be set")
	}
}

func TestContextFromRequest(t *testing.T) {
	// The ContextFromRequest function expects the request's context to be our custom Context type
	// This would typically be set up by middleware in a real application

	// Create a custom context that embeds our Context type
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Simulate how middleware would set the context
	// In reality, middleware would put the custom context into the request context
	reqWithCustomCtx := r.WithContext(ctx)

	// Test - ContextFromRequest should work with a request that has our context type in its context
	retrievedCtx, ok := request.ContextFromRequest(reqWithCustomCtx)

	// Assertions
	if !ok {
		t.Error("Expected to retrieve context from request with custom context")
	}

	if retrievedCtx == nil {
		t.Error("Expected retrieved context to not be nil")
	}

	// Test with regular request (should fail)
	regularReq := httptest.NewRequest("GET", "/test", nil)
	_, ok = request.ContextFromRequest(regularReq)
	if ok {
		t.Error("Expected to NOT retrieve context from regular request")
	}
}

func TestContext_GetPathParam(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123", nil)

	// Simulate path parameter (this would normally be set by the router)
	r.SetPathValue("id", "123")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	id := ctx.GetPathParam("id")

	// Assertions
	if id != "123" {
		t.Errorf("Expected path param 'id' to be '123', got '%s'", id)
	}

	// Test non-existent param
	missing := ctx.GetPathParam("nonexistent")
	if missing != "" {
		t.Errorf("Expected non-existent path param to be empty, got '%s'", missing)
	}
}

func TestContext_GetQueryParam(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test?name=john&age=25&tags=go,web", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test existing params
	name := ctx.GetQueryParam("name")
	if name != "john" {
		t.Errorf("Expected query param 'name' to be 'john', got '%s'", name)
	}

	age := ctx.GetQueryParam("age")
	if age != "25" {
		t.Errorf("Expected query param 'age' to be '25', got '%s'", age)
	}

	// Test non-existent param
	missing := ctx.GetQueryParam("nonexistent")
	if missing != "" {
		t.Errorf("Expected non-existent query param to be empty, got '%s'", missing)
	}
}

func TestContext_GetHeader(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer token123")
	r.Header.Set("X-Custom-Header", "custom-value")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test existing headers
	contentType := ctx.GetHeader("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", contentType)
	}

	auth := ctx.GetHeader("Authorization")
	if auth != "Bearer token123" {
		t.Errorf("Expected Authorization to be 'Bearer token123', got '%s'", auth)
	}

	// Test case-insensitive header access
	customHeader := ctx.GetHeader("x-custom-header")
	if customHeader != "custom-value" {
		t.Errorf("Expected X-Custom-Header to be 'custom-value', got '%s'", customHeader)
	}

	// Test non-existent header
	missing := ctx.GetHeader("Non-Existent")
	if missing != "" {
		t.Errorf("Expected non-existent header to be empty, got '%s'", missing)
	}
}

func TestContext_IsHeaderContainValue(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Accept", "application/json, text/html")
	r.Header.Set("Cache-Control", "no-cache, no-store")
	r.Header.Add("X-Multi", "value1")
	r.Header.Add("X-Multi", "value2")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test header contains value
	if !ctx.IsHeaderContainValue("Accept", "application/json") {
		t.Error("Expected Accept header to contain 'application/json'")
	}

	if !ctx.IsHeaderContainValue("Accept", "text/html") {
		t.Error("Expected Accept header to contain 'text/html'")
	}

	if !ctx.IsHeaderContainValue("Cache-Control", "no-cache") {
		t.Error("Expected Cache-Control header to contain 'no-cache'")
	}

	// Test multi-value header
	if !ctx.IsHeaderContainValue("X-Multi", "value1") {
		t.Error("Expected X-Multi header to contain 'value1'")
	}

	if !ctx.IsHeaderContainValue("X-Multi", "value2") {
		t.Error("Expected X-Multi header to contain 'value2'")
	}

	// Test header doesn't contain value
	if ctx.IsHeaderContainValue("Accept", "application/xml") {
		t.Error("Expected Accept header to NOT contain 'application/xml'")
	}

	// Test non-existent header
	if ctx.IsHeaderContainValue("Non-Existent", "value") {
		t.Error("Expected non-existent header to NOT contain any value")
	}
}

func TestContext_GetRawBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
		hasError bool
	}{
		{
			name:     "JSON body",
			body:     `{"name":"john","age":25}`,
			expected: `{"name":"john","age":25}`,
			hasError: false,
		},
		{
			name:     "Text body",
			body:     "Hello, World!",
			expected: "Hello, World!",
			hasError: false,
		},
		{
			name:     "Empty body",
			body:     "",
			expected: "",
			hasError: true, // io.EOF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			w := httptest.NewRecorder()
			var r *http.Request

			if tt.body == "" {
				r = httptest.NewRequest("POST", "/test", nil)
			} else {
				r = httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			}

			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			// Test
			body, err := ctx.GetRawBody()

			// Assertions
			if tt.hasError && err == nil {
				t.Error("Expected error for empty body")
			}

			if !tt.hasError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}

			if string(body) != tt.expected {
				t.Errorf("Expected body to be '%s', got '%s'", tt.expected, string(body))
			}

			// Test that subsequent calls return same data (caching)
			body2, err2 := ctx.GetRawBody()
			if string(body2) != string(body) {
				t.Error("Expected cached body to be the same")
			}
			if (err2 == nil) != (err == nil) {
				t.Error("Expected cached error to be the same")
			}
		})
	}
}

func TestContext_CacheBodyMultipleCalls(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	body := "test body content"
	r := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test multiple calls
	body1, err1 := ctx.GetRawBody()
	body2, err2 := ctx.GetRawBody()
	body3, err3 := ctx.GetRawBody()

	// All should return the same result
	if string(body1) != body || string(body2) != body || string(body3) != body {
		t.Error("Expected all calls to return same body content")
	}

	if err1 != nil || err2 != nil || err3 != nil {
		t.Error("Expected no errors on any call")
	}
}

func TestContext_EmbeddedTypes(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test that Context embeds context.Context
	if ctx.Context == nil {
		t.Error("Expected embedded context.Context to be available")
	}

	// Test that Context embeds response.Response
	if ctx.Response == nil {
		t.Error("Expected embedded response.Response to be available")
	}

	// Test that we can use context methods
	if ctx.Err() != nil {
		t.Error("Expected context to not be cancelled initially")
	}

	// Test that we can use response methods
	ctx.Response.WithMessage("test message")
	if ctx.Response.Message != "test message" {
		t.Error("Expected to be able to use response methods")
	}
}

func TestContext_ContextCancellation(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := request.NewContext(w, r)

	// Test that context is not cancelled initially
	if ctx.Err() != nil {
		t.Error("Expected context to not be cancelled initially")
	}

	// Cancel the context
	cancel()

	// Test that context is now cancelled
	if ctx.Err() != context.Canceled {
		t.Error("Expected context to be cancelled after calling cancel")
	}
}

func TestContext_NilBodyHandling(t *testing.T) {
	// Setup
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Body = nil // Explicitly set body to nil

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	// Test
	body, err := ctx.GetRawBody()

	// Assertions
	if err != io.EOF {
		t.Errorf("Expected io.EOF error for nil body, got: %v", err)
	}

	if len(body) != 0 {
		t.Errorf("Expected empty body for nil request body, got: %s", string(body))
	}
}
