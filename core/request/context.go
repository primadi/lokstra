package request

import (
	"context"
	"io"
	"lokstra/core/response"
	"net/http"
	"sync"
)

type Context struct {
	context.Context
	*response.Response

	Writer   http.ResponseWriter
	Request  *http.Request
	rawBody  []byte
	bodyOnce sync.Once
	bodyErr  error
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

func (ctx *Context) GetHeader(name string) string {
	return ctx.Request.Header.Get(name)
}

func (ctx *Context) cacheBody() {
	ctx.bodyOnce.Do(func() {
		if ctx.Request.Body == nil {
			ctx.bodyErr = io.EOF
			return
		}
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.bodyErr = err
		} else {
			ctx.rawBody = body
			if len(ctx.rawBody) == 0 {
				ctx.bodyErr = io.EOF
			}
		}
	})
}

func (ctx *Context) GetRawBody() ([]byte, error) {
	ctx.cacheBody()
	return ctx.rawBody, ctx.bodyErr
}
