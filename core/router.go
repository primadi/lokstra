package core

import "net/http"

type Router interface {
	Prefix() string

	UseNamedMiddleware(mwname string, params MiddlewareConfig) error

	Use(MiddlewareHandler) Router
	Handle(method HTTPMethod, path string, handler RequestHandler, mw ...MiddlewareHandler) Router
	HandleOverrideMiddleware(method HTTPMethod, path string, handler RequestHandler, mw ...MiddlewareHandler) Router
	GET(path string, handler RequestHandler, mw ...MiddlewareHandler) Router
	POST(path string, handler RequestHandler, mw ...MiddlewareHandler) Router
	PUT(path string, handler RequestHandler, mw ...MiddlewareHandler) Router
	PATCH(path string, handler RequestHandler, mw ...MiddlewareHandler) Router
	DELETE(path string, handler RequestHandler, mw ...MiddlewareHandler) Router

	MountStatic(prefix string, folder http.Dir) Router
	MountSPA(prefix string, fallbackFile string) Router
	MountReverseProxy(prefix string, target string) Router

	Group(prefix string) Router
	GroupBlock(prefix string, fn func(gr Router)) Router

	RecurseAllHandler(callback func(info RouteInfo))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	LockMiddleware()
	OverrideMiddleware() Router
	GetMiddleware() []MiddlewareHandler
}
