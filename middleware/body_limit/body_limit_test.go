package body_limit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
)

func TestBodyLimitMiddleware_WithinLimit(t *testing.T) {
	// Create middleware with 1KB limit
	middleware := BodyLimit(1024)

	// Create handler that reads body
	handler := func(ctx *request.Context) error {
		body, err := ctx.GetRawRequestBody()
		if err != nil {
			return err
		}

		ctx.WithMessage("Body read successfully").WithData(map[string]any{
			"bodySize": len(body),
			"content":  string(body),
		})
		return nil
	}

	// Wrap handler with middleware
	wrappedHandler := middleware(handler)

	// Test with small body (within limit)
	smallBody := strings.Repeat("a", 512) // 512 bytes
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(smallBody))
	r.Header.Set("Content-Type", "text/plain")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error for small body, got: %v", err)
	}

	if ctx.Response.Message != "Body read successfully" {
		t.Errorf("Expected success message, got: %s", ctx.Response.Message)
	}
}

func TestBodyLimitMiddleware_ExceedsLimit_ContentLength(t *testing.T) {
	// Create middleware with 1KB limit
	middleware := BodyLimit(1024)

	handler := func(ctx *request.Context) error {
		ctx.WithMessage("Handler should not be called")
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test with large body (exceeds limit via Content-Length)
	largeBody := strings.Repeat("a", 2048) // 2KB
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
	r.Header.Set("Content-Type", "text/plain")

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected error for large body")
	}

	// Should use the default status code from config (413)
	if ctx.Response.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413, got: %d", ctx.Response.StatusCode)
	}

	if ctx.Response.Message != "Request body too large" {
		t.Errorf("Expected 'Request body too large' in message, got: %s", ctx.Response.Message)
	}

	if !strings.Contains(err.Error(), "Request body too large") {
		t.Errorf("Expected 'Request body too large' in error, got: %v", err)
	}
}

func TestBodyLimitMiddleware_ExceedsLimit_Reading(t *testing.T) {
	// Create middleware with small limit
	middleware := BodyLimit(100)

	handler := func(ctx *request.Context) error {
		// Try to read body - this should trigger the limit
		_, err := ctx.GetRawRequestBody()
		return err
	}

	wrappedHandler := middleware(handler)

	// Test with body that exceeds limit during reading
	largeBody := strings.Repeat("a", 200) // 200 bytes
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))
	// Don't set Content-Length to test reading limit

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected error for large body during reading")
	}
}

func TestBodyLimitMiddleware_SkipLargePayloads(t *testing.T) {
	config := &Config{
		MaxSize:           1024,
		SkipLargePayloads: true,
		Message:           "Body too large",
		StatusCode:        http.StatusRequestEntityTooLarge,
	}

	middleware := BodyLimitMiddleware(config)

	handler := func(ctx *request.Context) error {
		ctx.WithMessage("Handler executed successfully")
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test with large body (should skip and continue)
	largeBody := strings.Repeat("a", 2048) // 2KB
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(largeBody))

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error with SkipLargePayloads=true, got: %v", err)
	}

	if ctx.Response.Message != "Handler executed successfully" {
		t.Errorf("Expected handler to execute, got message: %s", ctx.Response.Message)
	}
}

func TestBodyLimitMiddleware_NoBody(t *testing.T) {
	middleware := BodyLimit(1024)

	handler := func(ctx *request.Context) error {
		ctx.WithMessage("No body request handled")
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test with no body
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)
	if err != nil {
		t.Errorf("Expected no error for no-body request, got: %v", err)
	}

	if ctx.Response.Message != "No body request handled" {
		t.Errorf("Expected success message, got: %s", ctx.Response.Message)
	}
}

func TestBodyLimitMiddleware_CustomConfig(t *testing.T) {
	config := &Config{
		MaxSize:    500,
		Message:    "Custom error: payload too big",
		StatusCode: http.StatusBadRequest, // 400 instead of 413
	}

	middleware := BodyLimitMiddleware(config)

	handler := func(ctx *request.Context) error {
		return nil
	}

	wrappedHandler := middleware(handler)

	// Test with body exceeding custom limit
	body := strings.Repeat("a", 600)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", strings.NewReader(body))

	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	err := wrappedHandler(ctx)
	if err == nil {
		t.Error("Expected error with custom config")
	}

	if ctx.Response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected custom status 400, got: %d", ctx.Response.StatusCode)
	}

	if !strings.Contains(err.Error(), "Custom error: payload too big") {
		t.Errorf("Expected custom error message, got: %v", err)
	}
}

func TestConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() midware.Func
		expected int64
	}{
		{"BodyLimit1MB", BodyLimit1MB, 1024 * 1024},
		{"BodyLimit5MB", BodyLimit5MB, 5 * 1024 * 1024},
		{"BodyLimit10MB", BodyLimit10MB, 10 * 1024 * 1024},
		{"BodyLimit50MB", BodyLimit50MB, 50 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := tt.fn()

			if middleware == nil {
				t.Error("Expected middleware function, got nil")
			}

			// Test with small body (should work)
			handler := func(ctx *request.Context) error {
				ctx.WithMessage("Success")
				return nil
			}

			wrappedHandler := middleware(handler)

			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/test", strings.NewReader("small body"))

			ctx, cancel := request.NewContext(w, r)
			defer cancel()

			err := wrappedHandler(ctx)
			if err != nil {
				t.Errorf("Expected no error for small body, got: %v", err)
			}
		})
	}
}

func TestLimitedReadCloser(t *testing.T) {
	// Test the internal limitedReadCloser directly
	config := DefaultConfig()
	config.MaxSize = 5 // Very small limit

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/test", nil)
	ctx, cancel := request.NewContext(w, r)
	defer cancel()

	limitedReader := &limitedReadCloser{
		reader:    http.NoBody,
		remaining: config.MaxSize,
		config:    config,
		ctx:       ctx,
	}

	// Test Close method
	err := limitedReader.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got: %v", err)
	}

	// Test Read with exceeding data
	limitedReader.remaining = 0 // Simulate limit reached
	buf := make([]byte, 10)

	n, err := limitedReader.Read(buf)
	if n != 0 {
		t.Errorf("Expected 0 bytes read when limit reached, got: %d", n)
	}

	if err == nil {
		t.Error("Expected error when limit exceeded")
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxSize != 10*1024*1024 {
		t.Errorf("Expected default max body size 10MB, got: %d", config.MaxSize)
	}

	if config.Message != "Request body too large" {
		t.Errorf("Expected default error message, got: %s", config.Message)
	}

	if config.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected default status 413, got: %d", config.StatusCode)
	}

	if config.SkipLargePayloads != false {
		t.Error("Expected default SkipLargePayloads to be false")
	}
}
