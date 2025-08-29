package request

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/primadi/lokstra/core/response"
)

type Context struct {
	context.Context
	*response.Response

	Writer          http.ResponseWriter
	Request         *http.Request
	rawRequestBody  []byte
	requestBodyOnce sync.Once
	requestBodyErr  error
}

func NewContext(w http.ResponseWriter, r *http.Request) (*Context, func()) {
	ctx, cancel := context.WithCancel(r.Context())
	req := r.WithContext(ctx)
	resp := response.NewResponse()

	return &Context{
		Context:  ctx,
		Response: resp,
		Writer:   w,
		Request:  req,
	}, cancel
}

func ContextFromRequest(r *http.Request) (*Context, bool) {
	rc, ok := r.Context().(*Context)
	return rc, ok
}

func (ctx *Context) GetPathParam(name string) string {
	return ctx.Request.PathValue(name)
}

func (ctx *Context) GetQueryParam(name string) string {
	return ctx.Request.URL.Query().Get(name)
}

// ShouldStopMiddlewareChain checks if middleware chain should stop due to error or HTTP error status.
// This helper ensures consistent error checking across all middleware implementations.
//
// Returns true if:
//   - err is not nil (any error occurred)
//   - ctx.StatusCode >= 400 (HTTP error status)
//
// Usage in middleware:
//
//	err := next(ctx)
//	if ctx.ShouldStopMiddlewareChain(err) {
//	    return err
//	}
//	// Continue with post-processing...
func (ctx *Context) ShouldStopMiddlewareChain(err error) bool {
	// Pure check function - no side effects
	return err != nil || ctx.StatusCode >= 400 || ctx.Err() != nil
}

func (ctx *Context) GetHeader(name string) string {
	return ctx.Request.Header.Get(name)
}

func (ctx *Context) IsHeaderContainValue(name, value string) bool {
	if hdr, ok := ctx.Request.Header[name]; ok {
		for _, v := range hdr {
			if strings.Contains(v, value) {
				return true
			}
		}
	}
	return false
}

func (ctx *Context) cacheRequestBody() {
	ctx.requestBodyOnce.Do(func() {
		if ctx.Request.Body == nil {
			return
		}
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.requestBodyErr = err
		} else {
			ctx.rawRequestBody = body
		}
	})
}

func (ctx *Context) GetRawRequestBody() ([]byte, error) {
	ctx.cacheRequestBody()
	return ctx.rawRequestBody, ctx.requestBodyErr
}

// GetRawResponseBody returns the raw response body data
// It accesses the RawData field from the Response object
func (ctx *Context) GetRawResponseBody() ([]byte, error) {
	if ctx.Response == nil {
		return nil, nil
	}
	return ctx.Response.RawData, nil
}

// HTML renders HTML content with the specified status code
func (ctx *Context) HTML(status int, html string) error {
	return ctx.Response.HTML(status, html)
}

// HTMX renders HTMX content with the specified status code
func (ctx *Context) HTMX(status int, html string) error {
	return ctx.Response.WithHeader("Vary", "HX-Request").HTML(status, html)
}

// ErrorHTML renders HTML content with error status and message
func (ctx *Context) ErrorHTML(status int, html string) error {
	return ctx.Response.ErrorHTML(status, html)
}
