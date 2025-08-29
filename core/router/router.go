package router

import (
	"net/http"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"

	"github.com/valyala/fasthttp"
)

type Router interface {
	Prefix() string

	Use(any) Router
	Handle(method request.HTTPMethod, path string, handler any, mw ...any) Router
	HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router
	GET(path string, handler any, mw ...any) Router
	POST(path string, handler any, mw ...any) Router
	PUT(path string, handler any, mw ...any) Router
	PATCH(path string, handler any, mw ...any) Router
	DELETE(path string, handler any, mw ...any) Router

	WithOverrideMiddleware(enable bool) Router
	WithPrefix(prefix string) Router

	RawHandle(prefix string, handler http.Handler) Router
	RawHandleStripPrefix(prefix string, handler http.Handler) Router

	MountStatic(prefix string, folder http.Dir) Router
	// sources can be http.Dir, string, or fs.FS
	MountStaticWithFallback(prefix string, sources ...any) Router

	MountSPA(prefix string, fallbackFile string) Router
	MountReverseProxy(prefix string, target string, overrideMiddleware bool, mw ...any) Router
	MountRpcService(path string, service any, overrideMiddleware bool, mw ...any) Router

	Group(prefix string, mw ...any) Router
	GroupBlock(prefix string, fn func(gr Router)) Router

	RecurseAllHandler(callback func(rt *RouteMeta))
	DumpRoutes()

	ServeHTTP(w http.ResponseWriter, r *http.Request)
	FastHttpHandler() fasthttp.RequestHandler
	OverrideMiddleware() Router
	GetMiddleware() []*midware.Execution
	LockMiddleware()

	GetMeta() *RouterMeta
}
