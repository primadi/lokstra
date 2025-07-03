package router

import (
	"context"
	"lokstra/common/response"
	"net/http"
)

type RequestContext struct {
	// Inherit base context for cancellation, deadlines, etc.
	context.Context

	// Embed response to allow ctx.Ok(...), ctx.Error(...), etc.
	*response.Response

	// Raw HTTP writer and request
	Writer http.ResponseWriter
	Req    *http.Request
}

// Builder to create RequestContext from http.Handler
func NewRequestContext(w http.ResponseWriter, r *http.Request) *RequestContext {
	return &RequestContext{
		Context:  r.Context(), // inherit base context
		Writer:   w,
		Req:      r,
		Response: &response.Response{},
	}
}
