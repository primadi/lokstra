package body_limit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/request"
)

// TestMiddlewareResponseBehavior tests what client receives when middleware
// sets Message, Data, and returns Error
func TestMiddlewareResponseBehavior(t *testing.T) {
	// Create middleware with 100 byte limit
	middleware := BodyLimit(100)

	// Create a large payload (exceeds limit)
	largePayload := strings.Repeat("A", 200)

	// Create request with large payload
	req := httptest.NewRequest("POST", "/test", strings.NewReader(largePayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(largePayload)))

	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(nil, w, req)
	defer cancel()

	// Dummy handler (should not be called)
	dummyHandler := func(ctx *request.Context) error {
		t.Error("Handler should not be called when middleware returns error")
		return nil
	}

	// Apply middleware
	middlewareWithHandler := middleware(dummyHandler)

	// Execute middleware
	err := middlewareWithHandler(ctx)

	// Verify middleware returned error
	if err == nil {
		t.Error("Expected middleware to return error for large payload")
	}

	// Check what was set in the context
	t.Logf("Context Status Code: %d", ctx.StatusCode)
	t.Logf("Context Response Message: %s", ctx.Response.Message)
	t.Logf("Context Response Data: %+v", ctx.Response.Data)
	t.Logf("Middleware Error: %v", err)

	// Write response to check what client would receive
	err = ctx.Response.WriteHttp(w)
	if err != nil {
		t.Errorf("Failed to write response: %v", err)
	}

	// Check HTTP response
	t.Logf("HTTP Status Code: %d", w.Code)
	t.Logf("HTTP Response Body: %s", w.Body.String())

	// Verify behavior
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected HTTP status 413, got %d", w.Code)
	}

	// Check if response contains the message and data
	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "Request body too large") {
		t.Errorf("Expected response to contain error message, got: %s", responseBody)
	}

	if !strings.Contains(responseBody, "maxSize") {
		t.Errorf("Expected response to contain maxSize data, got: %s", responseBody)
	}
}
