package router_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

func TestComposeMiddleware_ErrorHandling(t *testing.T) {
	tests := []struct {
		name                    string
		middlewareSetStatusCode int
		middlewareReturnsError  bool
		expectInnerCall         bool
		expectError             bool
		description             string
	}{
		{
			name:                    "success_case",
			middlewareSetStatusCode: 200,
			middlewareReturnsError:  false,
			expectInnerCall:         true,
			expectError:             false,
			description:             "Normal success case should call inner handler",
		},
		{
			name:                    "middleware_sets_400_status",
			middlewareSetStatusCode: 400,
			middlewareReturnsError:  false,
			expectInnerCall:         false,
			expectError:             true,
			description:             "Middleware setting 400 status should stop chain",
		},
		{
			name:                    "middleware_sets_500_status",
			middlewareSetStatusCode: 500,
			middlewareReturnsError:  false,
			expectInnerCall:         false,
			expectError:             true,
			description:             "Middleware setting 500 status should stop chain",
		},
		{
			name:                    "middleware_returns_error",
			middlewareSetStatusCode: 500,
			middlewareReturnsError:  true,
			expectInnerCall:         false,
			expectError:             true,
			description:             "Middleware returning error should stop chain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track if inner handler was called
			innerCalled := false

			// Create final handler
			finalHandler := func(ctx *request.Context) error {
				innerCalled = true
				return nil
			}

			// Create middleware that sets status or returns error
			testMiddleware := func(next request.HandlerFunc) request.HandlerFunc {
				return func(ctx *request.Context) error {
					var err error
					// Pre-processing: Return error immediately if specified
					if tt.middlewareReturnsError {
						err = errors.New("middleware error")
					}
					if tt.middlewareSetStatusCode != 200 {
						ctx.StatusCode = tt.middlewareSetStatusCode
						err = errors.New("middleware set status code")
					}

					if ctx.ShouldStopMiddlewareChain(err) {
						return err
					}
					// Call next middleware/handler only if no pre-processing errors
					return next(ctx)
				}
			}

			// Create middleware execution
			mwExec := []*midware.Execution{
				{
					Name:           "test_middleware",
					MiddlewareFn:   testMiddleware,
					Priority:       50,
					ExecutionOrder: 0,
				},
			}

			// Compose middleware
			composedHandler := router.ComposeMiddlewareForTest(mwExec, finalHandler)

			// Create test context
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/test", nil)
			ctx, cancel := request.NewContext(nil, w, r)
			defer cancel()

			// Execute composed handler
			err := composedHandler(ctx)

			// Verify expectations
			if tt.expectInnerCall && !innerCalled {
				t.Errorf("Expected inner handler to be called, but it wasn't - %s", tt.description)
			}

			if !tt.expectInnerCall && innerCalled {
				t.Errorf("Expected inner handler NOT to be called, but it was - %s", tt.description)
			}

			if tt.expectError && err == nil {
				t.Errorf("Expected error to be returned, but got nil - %s", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got: %v - %s", err, tt.description)
			}
		})
	}
}

func TestComposeMiddleware_MultipleMiddleware(t *testing.T) {
	// Test multiple middleware with error in middle
	callOrder := []string{}

	// Create middlewares
	middleware1 := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			callOrder = append(callOrder, "mw1_before")
			err := next(ctx)
			if !ctx.ShouldStopMiddlewareChain(err) {
				callOrder = append(callOrder, "mw1_after")
			}
			return err
		}
	}

	middleware2 := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			callOrder = append(callOrder, "mw2_before")
			// This middleware sets error status and should stop chain
			ctx.StatusCode = 400
			// Don't call next when setting error status - stop the chain
			return errors.New("middleware2 set error status")
		}
	}

	middleware3 := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			callOrder = append(callOrder, "mw3_before")
			err := next(ctx)
			if !ctx.ShouldStopMiddlewareChain(err) {
				callOrder = append(callOrder, "mw3_after")
			}
			return err
		}
	}

	finalHandler := func(ctx *request.Context) error {
		callOrder = append(callOrder, "handler")
		return nil
	}

	// Create middleware executions
	mwExec := []*midware.Execution{
		{Name: "mw1", MiddlewareFn: middleware1, Priority: 10, ExecutionOrder: 0},
		{Name: "mw2", MiddlewareFn: middleware2, Priority: 20, ExecutionOrder: 1},
		{Name: "mw3", MiddlewareFn: middleware3, Priority: 30, ExecutionOrder: 2},
	}

	// Compose and execute
	composedHandler := router.ComposeMiddlewareForTest(mwExec, finalHandler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	ctx, cancel := request.NewContext(nil, w, r)
	defer cancel()

	err := composedHandler(ctx)

	// Verify error is returned
	if err == nil {
		t.Error("Expected error due to status 400, but got nil")
	}

	// Verify call order - middleware composition should stop early when status >= 400
	// The composition logic wraps from outermost to innermost, so when mw2 sets status 400,
	// mw3 should detect it and not proceed to inner handler
	expectedOrder := []string{"mw1_before", "mw2_before"}

	if len(callOrder) != len(expectedOrder) {
		t.Errorf("Expected %d calls, got %d: %v", len(expectedOrder), len(callOrder), callOrder)
		return
	}

	for i, expected := range expectedOrder {
		if callOrder[i] != expected {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expected, callOrder[i])
		}
	}

	// Verify handler was not called due to middleware chain stopping
	for _, call := range callOrder {
		if call == "handler" {
			t.Error("Handler should not have been called due to error status")
		}
	}

	// Verify neither mw2_after nor mw1_after were called (error stopped the chain)
	for _, call := range callOrder {
		if call == "mw2_after" || call == "mw1_after" {
			t.Errorf("After-processing should not have been called due to error: %s", call)
		}
	}
}
