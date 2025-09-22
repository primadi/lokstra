package router

import (
	"io/fs"
	"net/http"

	"github.com/primadi/lokstra/common/htmx_fsmanager"
	"github.com/primadi/lokstra/common/static_files"
	"github.com/primadi/lokstra/core/request"

	"github.com/valyala/fasthttp"
)

type Router interface {
	// Prefix returns the router's base prefix string (routerPrefix)
	// This is the base path that gets prepended to all routes in this router
	// Example: if Prefix() returns "/api/v1", then GET("/users") becomes "/api/v1/users"
	Prefix() string

	// Use adds middleware to the router's middleware stack
	Use(any) Router
	// Handle registers a new route with the given method, path, handler, and optional middleware
	// Final endpoint = routerPrefix + path
	Handle(method request.HTTPMethod, path string, handler any, mw ...any) Router
	// HandleOverrideMiddleware registers a new route with the given method, path, handler, and optional middleware, overriding the router's middleware stack
	HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router

	// GET registers an exact-match GET route at: routerPrefix + path
	// Example: router.WithPrefix("/api").GET("/users") handles exactly "/api/users"
	GET(path string, handler any, mw ...any) Router

	// GETPrefix registers a catch-all GET route that handles: routerPrefix + pathPrefix + /*
	// Example: router.WithPrefix("/api").GETPrefix("/files") handles "/api/files", "/api/files/doc.pdf", etc.
	// Parameter 'pathPrefix' is the method-level prefix, different from router's base prefix
	GETPrefix(pathPrefix string, handler any, mw ...any) Router

	// POST registers an exact-match POST route at: routerPrefix + path
	POST(path string, handler any, mw ...any) Router
	// POSTPrefix registers a catch-all POST route that handles: routerPrefix + pathPrefix + /*
	POSTPrefix(pathPrefix string, handler any, mw ...any) Router

	// PUT registers an exact-match PUT route at: routerPrefix + path
	PUT(path string, handler any, mw ...any) Router
	// PUTPrefix registers a catch-all PUT route that handles: routerPrefix + pathPrefix + /*
	PUTPrefix(pathPrefix string, handler any, mw ...any) Router

	// PATCH registers an exact-match PATCH route at: routerPrefix + path
	PATCH(path string, handler any, mw ...any) Router
	// PATCHPrefix registers a catch-all PATCH route that handles: routerPrefix + pathPrefix + /*
	PATCHPrefix(pathPrefix string, handler any, mw ...any) Router

	// DELETE registers an exact-match DELETE route at: routerPrefix + path
	DELETE(path string, handler any, mw ...any) Router
	// DELETEPrefix registers a catch-all DELETE route that handles: routerPrefix + pathPrefix + /*
	DELETEPrefix(pathPrefix string, handler any, mw ...any) Router

	// WithOverrideMiddleware enables or disables middleware override for the router
	WithOverrideMiddleware(enable bool) Router

	// WithPrefix sets the base prefix for all routes in this router (routerPrefix)
	// This affects ALL routes registered after this call
	// Example: router.WithPrefix("/api/v1") makes GET("/users") handle "/api/v1/users"
	WithPrefix(prefix string) Router

	// RawHandle registers a standard http.Handler for the given path prefix
	RawHandle(prefix string, stripPrefix bool, handler http.Handler) Router

	// MountStatic serves static files from multiple sources at the specified prefix, using the first available file
	MountStatic(prefix string, spa bool, sources ...fs.FS) Router

	// MountHtmx serves HTMX pages with layout support at the specified prefix, using the provided sources
	// Assume sources has:
	//   - "/layouts" for HTML layout templates
	//   - "/pages" for HTML page templates
	//
	// All Request paths will be treated as page requests,
	// it will be depreceted in the future.
	MountHtmx(prefix string, si *static_files.ScriptInjection, sources ...fs.FS) Router

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

	// GetMeta returns the router's metadata
	GetMeta() *RouterMeta

	// HTMX Support
	AddHtmxPages(source fs.FS, dir ...string) Router
	AddHtmxLayouts(source fs.FS, dir ...string) Router
	AddHtmxStatics(source fs.FS, dir ...string) Router
	SetHtmxFSManager(manager *htmx_fsmanager.HtmxFsManager) Router
	SetHTMXLayoutScriptInjection(si *htmx_fsmanager.ScriptInjection) Router
}
