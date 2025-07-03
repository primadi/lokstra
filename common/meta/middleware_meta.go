package meta

import (
	"lokstra/common/iface"
)

// MiddlewareMeta holds information about a middleware component.
// it can be used to define middleware that can be registered by middlewareType
// or as a function directly.
type MiddlewareMeta struct {
	MiddlewareType string
	Config         any // Configuration for the middleware

	MiddlewareFunc iface.MiddlewareFunc
}

func NamedMiddleware(middlewareType string, config ...any) *MiddlewareMeta {
	return &MiddlewareMeta{MiddlewareType: middlewareType, Config: config}
}
