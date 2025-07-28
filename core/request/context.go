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

func (ctx *Context) cacheBody() {
	ctx.bodyOnce.Do(func() {
		if ctx.Request.Body == nil {
			return
		}
		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.bodyErr = err
		} else {
			ctx.rawBody = body
		}
	})
}

func (ctx *Context) GetRawBody() ([]byte, error) {
	ctx.cacheBody()
	return ctx.rawBody, ctx.bodyErr
}
