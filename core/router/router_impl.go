package router

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/iface"
	"lokstra/common/meta"
	"lokstra/common/utils"
	"lokstra/core/request"
	"lokstra/serviceapi/core_service"
	"mime"
	"net/http"
	"slices"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type RouterImpl struct {
	mwLocked bool
	meta     *meta.RouterMeta
	r_engine core_service.RouterEngine
}

func NewRouter(ctx component.ComponentContext) Router {
	rtmt := meta.NewRouter()
	routerCtr++
	rtAny, err := ctx.NewService(rtmt.GetRouterEngineType(), fmt.Sprintf("r%d.router_engine", routerCtr))
	if err != nil {
		panic(fmt.Sprintf("Failed to create router engine %s: %v", rtmt.GetRouterEngineType(), err))
	}
	rt := rtAny.(core_service.RouterEngine)
	if rt == nil {
		panic(fmt.Sprintf("Router engine %s is not initialized", rtmt.GetRouterEngineType()))
	}
	return &RouterImpl{
		meta:     rtmt,
		r_engine: rt,
	}
}

var routerCtr = 0

func NewRouterWithEngine(ctx component.ComponentContext, engineType string) Router {
	rtmt := meta.NewRouter().WithRouterEngineType(engineType)
	routerCtr++
	rtAny, err := ctx.NewService(rtmt.GetRouterEngineType(), fmt.Sprintf("r%d.router_engine", routerCtr))
	if err != nil {
		panic(fmt.Sprintf("Failed to create router engine %s: %v", rtmt.GetRouterEngineType(), err))
	}
	rt := rtAny.(core_service.RouterEngine)
	if rt == nil {
		panic(fmt.Sprintf("Router engine %s is not initialized", rtmt.GetRouterEngineType()))
	}
	return &RouterImpl{
		meta:     rtmt,
		r_engine: rt,
	}
}

// DELETE implements Router.
func (r *RouterImpl) DELETE(path string, handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	return r.handle("DELETE", path, handler, false, true, mw...)
}

// DumpRoutes implements Router.
func (r *RouterImpl) DumpRoutes() {
	r.RecurseAllHandler(func(rt *meta.RouteMeta) {
		fmt.Printf("[ROUTE] %s %s\n", rt.Method, rt.Path)
	})
}

// FastHttpHandler implements Router.
func (r *RouterImpl) FastHttpHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(r.r_engine)
}

// GET implements Router.
func (r *RouterImpl) GET(path string, handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	return r.handle("GET", path, handler, false, true, mw...)
}

// GetMiddleware implements Router.
func (r *RouterImpl) GetMiddleware() []iface.MiddlewareFunc {
	mwf := make([]iface.MiddlewareFunc, len(r.meta.Middleware))
	for i, mw := range r.meta.Middleware {
		mwf[i] = mw.MiddlewareFunc
	}
	return mwf
}

// Group implements Router.
func (r *RouterImpl) Group(prefix string, mw ...iface.MiddlewareFunc) Router {
	r.mwLocked = true

	rm := meta.NewRouter()
	rm.Prefix = r.cleanPrefix(prefix)

	for _, m := range mw {
		rm.UseMiddleware(m)
	}

	return &GroupImpl{
		parent: r,
		meta:   rm,
	}
}

// GroupBlock implements Router.
func (r *RouterImpl) GroupBlock(prefix string, fn func(gr Router)) Router {
	gr := r.Group(prefix)
	fn(gr)
	return r
}

// Handle implements Router.
func (r *RouterImpl) Handle(method iface.HTTPMethod, path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	return r.handle(method, path, handler, false, true, mw...)
}

func (r *RouterImpl) handle(method iface.HTTPMethod, path string, handler request.HandlerFunc,
	overrideMiddleware bool, updateMeta bool, mw ...iface.MiddlewareFunc) Router {
	r.mwLocked = true

	var mwh []iface.MiddlewareFunc
	if overrideMiddleware {
		mwh = mw
	} else {
		mwh = slices.Concat(r.GetMiddleware(), mw)
	}

	cleanPath := r.cleanPrefix(path)
	if updateMeta {
		if overrideMiddleware {
			r.meta.HandleWithOverrideMiddleware(method, cleanPath, handler, utils.ToAnySlice(mwh)...)
		} else {
			r.meta.Handle(method, cleanPath, handler, utils.ToAnySlice(mwh)...)
		}
	}

	handler_with_mw := composeMiddleware(mwh, handler)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, ok := contextFromRequest(req)

		var cancel func()
		if !ok {
			ctx, cancel = NewContext(w, req)
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
	r.r_engine.HandleMethod(string(method), cleanPath, finalHandler)

	return r
}

// HandleOverrideMiddleware implements Router.
func (r *RouterImpl) HandleOverrideMiddleware(method iface.HTTPMethod, path string,
	handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	return r.handle(method, path, handler, true, true, mw...)
}

// LockMiddleware implements Router.
func (r *RouterImpl) LockMiddleware() {
	r.mwLocked = true
}

// MountReverseProxy implements Router.
func (r *RouterImpl) MountReverseProxy(prefix string, target string) Router {
	r.r_engine.ServeReverseProxy(prefix, target)
	return r
}

// MountSPA implements Router.
func (r *RouterImpl) MountSPA(prefix string, fallbackFile string) Router {
	r.r_engine.ServeSPA(prefix, fallbackFile)
	return r
}

// MountStatic implements Router.
func (r *RouterImpl) MountStatic(prefix string, folder http.Dir) Router {
	r.r_engine.ServeStatic(prefix, folder)
	return r
}

// OverrideMiddleware implements Router.
func (r *RouterImpl) OverrideMiddleware() Router {
	r.meta.OverrideMiddleware = true
	return r
}

// PATCH implements Router.
func (r *RouterImpl) PATCH(path string, handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	return r.handle("PATCH", path, handler, false, true, mw...)
}

// POST implements Router.
func (r *RouterImpl) POST(path string, handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	return r.handle("POST", path, handler, false, true, mw...)
}

// PUT implements Router.
func (r *RouterImpl) PUT(path string, handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	return r.handle("PUT", path, handler, false, true, mw...)
}

// Prefix implements Router.
func (r *RouterImpl) Prefix() string {
	return r.meta.Prefix
}

// RecurseAllHandler implements Router.
func (r *RouterImpl) RecurseAllHandler(callback func(rt *meta.RouteMeta)) {
	for _, route := range r.meta.Routes {
		callback(route)
	}
}

// ServeHTTP implements Router.
func (r *RouterImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.r_engine.ServeHTTP(w, req)
}

// Use implements Router.
func (r *RouterImpl) Use(mw iface.MiddlewareFunc) Router {
	if r.mwLocked {
		panic("Cannot add middleware after locking the router")
	}
	r.meta.Middleware = append(r.meta.Middleware, &meta.MiddlewareMeta{
		MiddlewareFunc: mw,
	})
	return r
}

// WithOverrideMiddleware implements Router.
func (r *RouterImpl) WithOverrideMiddleware(enable bool) Router {
	r.meta.OverrideMiddleware = enable
	return r
}

// WithPrefix implements Router.
func (r *RouterImpl) WithPrefix(prefix string) Router {
	r.meta.Prefix = r.cleanPrefix(prefix)
	return r
}

var _ Router = (*RouterImpl)(nil)

func (r *RouterImpl) cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return r.meta.Prefix
	}

	if r.meta.Prefix == "/" {
		return "/" + strings.Trim(prefix, "/")
	}
	return r.meta.Prefix + "/" + strings.Trim(prefix, "/")
}

func composeMiddleware(mw []iface.MiddlewareFunc,
	finalHandler request.HandlerFunc) request.HandlerFunc {
	handler := finalHandler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return handler
}

func contextFromRequest(r *http.Request) (*request.Context, bool) {
	rc, ok := r.Context().(*request.Context)
	return rc, ok
}

func init() {
	mime.AddExtensionType(".wasm", "application/wasm")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	mime.AddExtensionType(".gz", "application/gzip")
	mime.AddExtensionType(".map", "application/json")
}
