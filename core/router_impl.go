package core

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/core/request"
	"lokstra/core/router"
	"lokstra/core/router/listener"
	"lokstra/core/router/router_engine"
	"net/http"
	"slices"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type RouterImpl struct {
	prefix             string
	overrideMiddleware bool
	listener           listener.HttpListener
	rtre               router_engine.RouterEngine
	middleware         []iface.MiddlewareFunc
	routes             []*router.RouteHandlerData
	middlewareLocked   bool
}

// DumpRoutes implements router.Router.
func (r *RouterImpl) DumpRoutes() {
	r.RecurseAllHandler(func(rt *router.RouteHandlerData) {
		fmt.Printf("[ROUTE] %s %s\n", rt.Method, rt.Path)
	})
}

func (r *RouterImpl) registerRoute(data *router.RouteHandlerData) {
	r.routes = append(r.routes, data)
}

// LockMiddleware implements router.Router.
func (r *RouterImpl) LockMiddleware() {
	r.middlewareLocked = true
}

// FastHttpHandler implements router.Router.
func (r *RouterImpl) FastHttpHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(r.rtre)
}

// DELETE implements router.Router.
func (r *RouterImpl) DELETE(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	r.Handle("DELETE", path, handler, mw...)
	return r
}

// GET implements router.Router.
func (r *RouterImpl) GET(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	r.Handle("GET", path, handler, mw...)
	return r
}

// GetMiddleware implements router.Router.
func (r *RouterImpl) GetMiddleware() []iface.MiddlewareFunc {
	return r.middleware
}

// Group implements router.Router.
func (r *RouterImpl) Group(prefix string,
	mw ...iface.MiddlewareFunc) router.Router {
	r.LockMiddleware()

	return &GroupImpl{
		parent:     r,
		prefix:     r.cleanPrefix(prefix),
		middleware: mw,
	}
}

// GroupBlock implements router.Router.
func (r *RouterImpl) GroupBlock(prefix string, fn func(gr router.Router)) router.Router {
	gr := r.Group(prefix)
	fn(gr)
	return r
}

// Handle implements router.Router.
func (r *RouterImpl) Handle(method iface.HTTPMethod, path string,
	handler request.HandlerFunc, mw ...iface.MiddlewareFunc) router.Router {
	return r.handle(method, path, handler, false, mw...)
}

// HandleOverrideMiddleware implements router.Router.
func (r *RouterImpl) HandleOverrideMiddleware(method iface.HTTPMethod, path string,
	handler request.HandlerFunc, mw ...iface.MiddlewareFunc) router.Router {
	return r.handle(method, path, handler, true, mw...)
}

func (r *RouterImpl) handle(method iface.HTTPMethod, path string, handler request.HandlerFunc,
	overrideMiddleware bool, mw ...iface.MiddlewareFunc) router.Router {
	r.LockMiddleware()

	var mwh []iface.MiddlewareFunc
	if overrideMiddleware {
		mwh = mw
	} else {
		mwh = slices.Concat(r.GetMiddleware(), mw)
	}

	cleanPath := r.cleanPrefix(path)
	r.registerRoute(&router.RouteHandlerData{
		Path:           cleanPath,
		Method:         method,
		HandlerFunc:    handler,
		MiddlewareFunc: mwh,
	})

	handler_with_mw := composeMiddleware(mwh, handler)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, ok := contextFromRequest(req)

		var cancel func()
		if !ok {
			ctx, cancel = NewRequestContext(w, req)
			defer cancel()
		}
		if err := handler_with_mw(ctx); err != nil {
			ctx.ErrorInternal(err.Error())
		}
		if err := ctx.Err(); err != nil {
			ctx.ErrorInternal("Request aborted")
		}
		ctx.Response.WriteHttp(ctx.W)
	})
	r.rtre.HandleMethod(string(method), cleanPath, finalHandler)

	return r
}

// MountReverseProxy implements router.Router.
func (r *RouterImpl) MountReverseProxy(prefix string, target string) router.Router {
	r.rtre.ServeReverseProxy(prefix, target)
	return r
}

// MountSPA implements router.Router.
func (r *RouterImpl) MountSPA(prefix string, fallbackFile string) router.Router {
	r.rtre.ServeSPA(prefix, fallbackFile)
	return r
}

// MountStatic implements router.Router.
func (r *RouterImpl) MountStatic(prefix string, folder http.Dir) router.Router {
	r.rtre.ServeStatic(prefix, folder)
	return r
}

// OverrideMiddleware implements router.Router.
func (r *RouterImpl) OverrideMiddleware() router.Router {
	r.overrideMiddleware = true
	return r
}

// PATCH implements router.Router.
func (r *RouterImpl) PATCH(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	r.handle("PATCH", path, handler, false, mw...)
	return r
}

// POST implements router.Router.
func (r *RouterImpl) POST(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	r.handle("POST", path, handler, false, mw...)
	return r
}

// PUT implements router.Router.
func (r *RouterImpl) PUT(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	r.handle("PUT", path, handler, false, mw...)
	return r
}

// Prefix implements router.Router.
func (r *RouterImpl) Prefix() string {
	return r.prefix
}

// RecurseAllHandler implements router.Router.
func (r *RouterImpl) RecurseAllHandler(callback func(rt *router.RouteHandlerData)) {
	for _, route := range r.routes {
		callback(route)
	}
}

// ServeHTTP implements router.Router.
func (r *RouterImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.rtre.ServeHTTP(w, req)
}

// Use implements router.Router.
func (r *RouterImpl) Use(mw iface.MiddlewareFunc) router.Router {
	if r.middlewareLocked {
		panic("Cannot add middleware after router is locked")
	}
	r.middleware = append(r.middleware, mw)
	return r
}

// WithOverrideMiddleware implements router.Router.
func (r *RouterImpl) WithOverrideMiddleware(enable bool) router.Router {
	r.overrideMiddleware = enable
	return r
}

// WithPrefix implements router.Router.
func (r *RouterImpl) WithPrefix(prefix string) router.Router {
	r.prefix = prefix
	return r
}

func NewRouter() router.Router {
	return &RouterImpl{
		prefix:     "",
		rtre:       globalRuntime.newRouterEngineFunc(),
		middleware: []iface.MiddlewareFunc{},
		routes:     []*router.RouteHandlerData{},
	}
}

var _ router.Router = (*RouterImpl)(nil)

func (r *RouterImpl) cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return r.prefix
	}

	if r.prefix == "/" {
		return "/" + strings.Trim(prefix, "/")
	}
	return r.prefix + "/" + strings.Trim(prefix, "/")
}

func contextFromRequest(r *http.Request) (*request.Context, bool) {
	rc, ok := r.Context().(*request.Context)
	return rc, ok
}
