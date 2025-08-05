package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
)

func anyArraytoMiddleware(middleware []any) []*midware.Execution {
	mwExec := make([]*midware.Execution, len(middleware))
	for i := range middleware {
		if middleware[i] == nil {
			continue
		}

		var mw *midware.Execution
		switch m := middleware[i].(type) {
		case midware.Func:
			mw = midware.NewExecution(m)
		case string:
			mw = midware.Named(m)
		case *midware.Execution:
			mw = m
		default:
			panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareExecution")
		}

		mwExec[i] = mw
	}
	return mwExec
}

func createReverseProxyHandler(target string) request.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		panic("invalid proxy target: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	return func(ctx *request.Context) error {
		fmt.Printf("Proxying request to %s\n", targetURL.String())
		r := ctx.Request
		w := ctx.Writer
		r.URL.Scheme = targetURL.Scheme
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host
		proxy.ServeHTTP(w, r)
		return nil
	}
}

func composeReverseProxyMw(rp *ReverseProxyMeta, mwParent []*midware.Execution) http.HandlerFunc {
	var mw []*midware.Execution

	if rp.OverrideMiddleware {
		mw = make([]*midware.Execution, len(rp.Middleware))
		copy(mw, rp.Middleware)
	} else {
		mw = utils.SlicesConcat(mwParent, rp.Middleware)
	}

	// Update execution order based on order of addition
	execOrder := 0
	for _, m := range mw {
		m.ExecutionOrder = execOrder
		execOrder++
	}

	// Sort middleware by priority and execution order
	slices.SortStableFunc(mw, func(a, b *midware.Execution) int {
		aOrder := a.Priority + a.ExecutionOrder
		bOrder := b.Priority + b.ExecutionOrder

		if aOrder < bOrder {
			return -1
		} else if aOrder > bOrder {
			return 1
		}

		return 0
	})

	// Create the final proxy handler
	proxyHandler := createReverseProxyHandler(rp.Target)

	// Wrap proxy handler with error-aware middleware composition
	// Start from the innermost handler (proxy) and wrap outward
	handler := proxyHandler
	for i := len(mw) - 1; i >= 0; i-- {
		currentMw := mw[i]
		handler = func(innerHandler request.HandlerFunc, middleware *midware.Execution) request.HandlerFunc {
			return middleware.MiddlewareFn(func(ctx *request.Context) error {
				// Check if previous middleware already set an error response (4xx or 5xx)
				if ctx.Response.StatusCode >= 400 {
					// Return error to stop middleware chain and prevent "after" logic
					return errors.New("previous middleware set error response")
				}
				return innerHandler(ctx)
			})
		}(handler, currentMw)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, deferFunc := request.NewContext(w, r)
		defer deferFunc()

		// Execute the wrapped middleware chain
		err := handler(ctx)

		// Handle response writing
		if err != nil || ctx.Response.StatusCode >= 400 {
			// Error case - write lokstra response
			ctx.Response.WriteHttp(ctx.Writer)
		}
		// Success case - proxy has already written response directly to http.ResponseWriter
		// No need to call WriteHttp for successful proxy responses
	})
}
