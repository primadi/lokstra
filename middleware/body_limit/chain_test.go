package body_limit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// TestFullMiddlewareChainBehavior tests how Lokstra handles middleware chain
// when middleware sets Message, Data, and returns Error
func TestFullMiddlewareChainBehavior(t *testing.T) {
	// Create body limit middleware
	bodyLimitMw := BodyLimit(100)

	// Create a logging middleware to see if it runs
	var logMessages []string
	loggingMw := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			logMessages = append(logMessages, "before")
			err := next(ctx)
			if ctx.ShouldStopMiddlewareChain(err) {
				logMessages = append(logMessages, "error-stop")
				return err
			}
			logMessages = append(logMessages, "after")
			return nil
		}
	}

	// Final handler
	var handlerCalled bool
	finalHandler := func(ctx *request.Context) error {
		handlerCalled = true
		logMessages = append(logMessages, "handler")
		return nil
	}

	// Create middleware executions
	mwExecs := []*midware.Execution{
		{
			Name:           "logging",
			MiddlewareFn:   loggingMw,
			Priority:       10,
			ExecutionOrder: 0,
		},
		{
			Name:           "body_limit",
			MiddlewareFn:   bodyLimitMw,
			Priority:       20,
			ExecutionOrder: 1,
		},
	}

	// Compose middleware chain using Lokstra's composer
	composedHandler := router.ComposeMiddlewareForTest(mwExecs, finalHandler)

	// Create test request with large payload
	largePayload := strings.Repeat("A", 200)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(largePayload))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(largePayload)))

	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute composed middleware chain
	err := composedHandler(ctx)

	// Verify results
	t.Logf("Error returned: %v", err)
	t.Logf("Log messages: %v", logMessages)
	t.Logf("Handler called: %v", handlerCalled)
	t.Logf("Context Status: %d", ctx.StatusCode)
	t.Logf("Context Message: %s", ctx.Response.Message)
	t.Logf("Context Data: %+v", ctx.Response.Data)

	// Write response to see final output
	_ = ctx.Response.WriteHttp(w)
	t.Logf("HTTP Status: %d", w.Code)
	t.Logf("HTTP Body: %s", w.Body.String())

	// Verify expectations
	if err == nil {
		t.Error("Expected error to be returned from middleware chain")
	}

	if handlerCalled {
		t.Error("Final handler should not be called when middleware returns error")
	}

	if !strings.Contains(logMessages[len(logMessages)-1], "error-stop") {
		t.Error("Expected logging middleware to detect error and stop")
	}

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected HTTP status 413, got %d", w.Code)
	}

	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "Request body too large") {
		t.Errorf("Expected response body to contain error message")
	}
}
