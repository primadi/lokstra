package iface

import "github.com/primadi/lokstra/core/request"

type HandlerFunc = func(ctx *request.Context) error

type MiddlewareFunc = func(next HandlerFunc) HandlerFunc
type MiddlewareFactory = func(config any) MiddlewareFunc

type MiddlewareMeta struct {
	Priority    int // Lower number means higher priority (1-100)
	Description string
	Tags        []string // Tags for categorization
}

type MiddlewareModule interface {
	Name() string
	Factory(config any) MiddlewareFunc
	Meta() *MiddlewareMeta
}
