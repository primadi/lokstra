package request

import (
	"context"
	"net/http"

	"github.com/primadi/lokstra/core/response"
)

type Context struct {
	// Embedding standard context and response for easy access
	context.Context
	*response.Response

	// for internal use only
	W *writerWrapper
	// for internal use only
	R *http.Request

	index    int
	handlers []HandlerFunc
}

func NewContext(w http.ResponseWriter, r *http.Request, handlers []HandlerFunc) *Context {
	return &Context{
		W:        newWriterWrapper(w),
		R:        r,
		handlers: handlers,
		Response: &response.Response{},
	}
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
		st := c.RespStatusCode
		if st == 0 || st < http.StatusBadRequest {
			c.ErrorInternal(err)
		}
	}

	c.WriteHttp(c.W)
}

func (c *Context) executeHandler() error {
	return c.Next()
}
