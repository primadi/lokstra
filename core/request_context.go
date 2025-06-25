package core

import (
	"context"
	"net/http"
)

// NewRequestContext creates a new RequestContext instance by wrapping the original
// HTTP request with a cancellable context. It returns the created RequestContext
// and a cancel function that should be called when the request handling is complete.
//
// Usage:
//
//	ctx, cancel := NewRequestContext(w, r)
//	defer cancel()
//
// The cancel function is important to avoid context leaks and should be deferred
// or explicitly called when the request is done being processed.
func NewRequestContext(w http.ResponseWriter, r *http.Request) (*RequestContext, context.CancelFunc) {
	ctx, cancel := context.WithCancel(r.Context())
	req := r.WithContext(ctx)

	rc := &RequestContext{
		Context: ctx,
		Writer:  w,
		Request: req,
		values:  make(map[string]any),
	}

	return contextHelperBuilder(rc), cancel
}
