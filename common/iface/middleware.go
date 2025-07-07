package iface

import "lokstra/core/request"

type HandlerFunc func(ctx *request.Context) error
type MiddlewareFunc func(next HandlerFunc) HandlerFunc
type MiddlewareFactory = func(config any) MiddlewareFunc

type MiddlewareHandler interface {
	GetName() string
	GetMiddlewareFunc() MiddlewareFunc
}
