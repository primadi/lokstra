package request

import (
	"context"
	"net/http"

	"github.com/primadi/lokstra/core/response"
)

type Context struct {
	// Embedding standard context for easy access
	context.Context

	// Helper to access request methods and fields
	Req *RequestHelper
	// Helper to access response methods and fields
	Resp *response.Response
	// Helper for opinionated API responses (wraps data in ApiResponse)
	Api *response.ApiHelper

	// Direct access to primitives (for advanced usage)
	W *writerWrapper
	R *http.Request

	index    int
	handlers []HandlerFunc
}

func NewContext(w http.ResponseWriter, r *http.Request, handlers []HandlerFunc) *Context {
	resp := &response.Response{}

	ctx := &Context{
		W:        newWriterWrapper(w),
		R:        r,
		handlers: handlers,
		Resp:     resp,                        // Direct assignment to Resp
		Api:      response.NewApiHelper(resp), // Initialize API helper
	}

	// Initialize request helper
	ctx.Req = newRequestHelper(ctx)

	return ctx
}

// Call inside middleware
func (c *Context) Next() error {
	if c.index >= len(c.handlers) {
		return nil
	}
	h := c.handlers[c.index]
	c.index++
	return h(c)
}

// Finalizes the response, writing status code and body if not already written
func (c *Context) FinalizeResponse(err error) {
	if c.W.ManualWritten() {
		// User already wrote directly to ResponseWriter, do nothing
		return
	}

	if err != nil {
		st := c.Resp.RespStatusCode
		if st == 0 || st < http.StatusBadRequest {
			c.Resp.WithStatus(http.StatusInternalServerError).Json(map[string]string{"error": err.Error()})
		}
	}

	c.Resp.WriteHttp(c.W)
}

func (c *Context) executeHandler() error {
	return c.Next()
}
