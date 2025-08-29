package request

import "net/http"

type HandlerFunc = func(ctx *Context) error
type HandlerFuncWithParam[T any] = func(ctx *Context, params *T) error
type RawHandlerFunc = func(w http.ResponseWriter, r *http.Request)

type HandlerRegister struct {
	Name        string
	HandlerFunc HandlerFunc
}

type RawHandlerRegister struct {
	Name        string
	HandlerFunc RawHandlerFunc
}
