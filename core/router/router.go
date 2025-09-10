package router

import (
	"io/fs"
	"net/http"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"

	"github.com/valyala/fasthttp"
)

type Router interface {
	// Prefix returns the router's prefix string
	Prefix() string

	// Use adds middleware to the router's middleware stack
	Use(any) Router
	// Handle registers a new route with the given method, path, handler, and optional middleware
	Handle(method request.HTTPMethod, path string, handler any, mw ...any) Router
	// HandleOverrideMiddleware registers a new route with the given method, path, handler, and optional middleware, overriding the router's middleware stack
	HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router

	// GET is a shortcut for router.Handle("GET", path, handler, mw...)
	GET(path string, handler any, mw ...any) Router
	// POST is a shortcut for router.Handle("POST", path, handler, mw...)
	POST(path string, handler any, mw ...any) Router
	// PUT is a shortcut for router.Handle("PUT", path, handler, mw...)
	PUT(path string, handler any, mw ...any) Router
	// PATCH is a shortcut for router.Handle("PATCH", path, handler, mw...)
	PATCH(path string, handler any, mw ...any) Router
	// DELETE is a shortcut for router.Handle("DELETE", path, handler, mw...)
	DELETE(path string, handler any, mw ...any) Router

	// WithOverrideMiddleware enables or disables middleware override for the router
	WithOverrideMiddleware(enable bool) Router

	// WithPrefix sets a prefix for all routes in the router
	WithPrefix(prefix string) Router

	// RawHandle registers a standard http.Handler for the given path prefix
	RawHandle(prefix string, stripPrefix bool, handler http.Handler) Router

	// MountStatic serves static files from multiple sources at the specified prefix, using the first available file
	MountStatic(prefix string, spa bool, sources ...fs.FS) Router

	// MountReverseProxy mounts a reverse proxy at the specified prefix, targeting the given URL, with optional middleware and override option
	MountReverseProxy(prefix string, target string, overrideMiddleware bool, mw ...any) Router
	// MountRpcService mounts an RPC service at the specified path, with optional middleware and override option
	MountRpcService(path string, service any, overrideMiddleware bool, mw ...any) Router

	// Group creates a sub-router with the given prefix and optional middleware
	Group(prefix string, mw ...any) Router
	// GroupBlock creates a sub-router with the given prefix and applies the provided function to it
	GroupBlock(prefix string, fn func(gr Router)) Router

	// RecurseAllHandler calls the given callback for each registered route, including those in sub-routers
	RecurseAllHandler(callback func(rt *RouteMeta))
	// DumpRoutes prints all registered routes and mounts (comprehensive view)
	DumpRoutes()

	// AddRouter merges another router into the current router
	AddRouter(r Router) Router

	// ServeHTTP makes the router implement the http.Handler interface
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	// FastHttpHandler returns a fasthttp.RequestHandler for the router
	FastHttpHandler() fasthttp.RequestHandler
	// OverrideMiddleware returns whether the router overrides middleware
	OverrideMiddleware() bool
	// GetMiddleware returns the router's middleware stack
	GetMiddleware() []*midware.Execution

	// GetMeta returns the router's metadata
	GetMeta() *RouterMeta
}
