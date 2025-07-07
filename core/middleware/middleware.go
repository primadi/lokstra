package middleware

import (
	"lokstra/common/iface"
	"lokstra/core/request"
)

func composeMiddleware(mw []iface.MiddlewareFunc,
	finalHandler request.HandlerFunc) request.HandlerFunc {
	handler := finalHandler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return handler
}
