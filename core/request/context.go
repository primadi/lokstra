package request

import (
	"context"
	"lokstra/common/response"
	"net/http"
)

type Context struct {
	context.Context
	*response.Response
	W http.ResponseWriter
	R *http.Request
}

type HandlerFunc = func(ctx *Context) error
