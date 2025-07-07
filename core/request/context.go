package request

import (
	"context"
	"lokstra/core/response"
	"net/http"
)

type Context struct {
	context.Context
	*response.Response
	W http.ResponseWriter
	R *http.Request
}

type HandlerFunc = func(ctx *Context) error
