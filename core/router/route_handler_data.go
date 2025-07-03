package router

import (
	"lokstra/common/iface"
	"lokstra/core/request"
)

type RouteHandlerData struct {
	Path        string
	Method      iface.HTTPMethod
	HandlerFunc request.HandlerFunc

	OverrideMiddleware bool
	MiddlewareFunc     []iface.MiddlewareFunc
}
