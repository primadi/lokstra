package body_limit

import (
	"net/http/httptest"
	"testing"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// TestPrematureStopCheck tests the bug in router composition where
// ShouldStopMiddlewareChain is called BEFORE middleware runs
func TestPrematureStopCheck(t *testing.T) {
	var callLog []string

	// Middleware 1: Sets status code
	middleware1 := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			callLog = append(callLog, "mw1-before")
			ctx.StatusCode = 400 // Set error status
			err := next(ctx)
			callLog = append(callLog, "mw1-after")
			return err
		}
	}

	// Middleware 2: Should detect the error status
	middleware2 := func(next request.HandlerFunc) request.HandlerFunc {
		return func(ctx *request.Context) error {
			callLog = append(callLog, "mw2-before")
			err := next(ctx)
			callLog = append(callLog, "mw2-after")
			return err
		}
	}

	// Final handler
	finalHandler := func(ctx *request.Context) error {
		callLog = append(callLog, "handler")
		return nil
	}

	// Create middleware chain
	mwExecs := []*midware.Execution{
		{Name: "mw1", MiddlewareFn: middleware1, Priority: 10},
		{Name: "mw2", MiddlewareFn: middleware2, Priority: 20},
	}

	// Use Lokstra's composition (with the bug)
	composedHandler := router.ComposeMiddlewareForTest(mwExecs, finalHandler)

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	ctx, cancel := request.NewContext(w, req)
	defer cancel()

	// Execute
	err := composedHandler(ctx)

	t.Logf("Call log: %v", callLog)
	t.Logf("Final status: %d", ctx.StatusCode)
	t.Logf("Error: %v", err)

	// With current buggy implementation, what happens?
	// The premature check prevents normal middleware flow
}
