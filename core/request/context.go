package request

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/primadi/lokstra/common/htmx_fsmanager"
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
	hfmContainer    htmx_fsmanager.IContainer
}

func NewContext(hfmContainer htmx_fsmanager.IContainer, w http.ResponseWriter, r *http.Request) (*Context, func()) {
	ctx, cancel := context.WithCancel(r.Context())
	req := r.WithContext(ctx)
	resp := response.NewResponse()

	return &Context{
		Context:      ctx,
		Response:     resp,
		Writer:       response.NewResponseWriterWrapper(w),
		Request:      req,
		hfmContainer: hfmContainer,
	}, cancel
}

func ContextFromRequest(r *http.Request) (*Context, bool) {
	rc, ok := r.Context().(*Context)
	return rc, ok
}

func (ctx *Context) GetPathParam(name string) string {
	return ctx.Request.PathValue(name)
}

func (ctx *Context) GetPathParamWithDefault(name string, defaultValue string) string {
	if value := ctx.Request.PathValue(name); value != "" {
		return value
	}
	return defaultValue
}

func (ctx *Context) GetQueryParam(name string) string {
	return ctx.Request.URL.Query().Get(name)
}

func (ctx *Context) GetQueryParamWithDefault(name string, defaultValue string) string {
	if value := ctx.Request.URL.Query().Get(name); value != "" {
		return value
	}
	return defaultValue
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

func (ctx *Context) GetHeaderWithDefault(name, defaultValue string) string {
	if value := ctx.Request.Header.Get(name); value != "" {
		return value
	}
	return defaultValue
}

func (ctx *Context) GetHeaders(name string) []string {
	if hdr, ok := ctx.Request.Header[name]; ok {
		return hdr
	}
	return nil
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
