package router

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/iface"
	"lokstra/common/meta"
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

// GetMeta implements Router.
func (r *RouterImpl) GetMeta() *meta.RouterMeta {
	return r.meta
}

// DELETE implements Router.
func (r *RouterImpl) DELETE(path string, handler any, mw ...any) Router {
	return r.handle("DELETE", path, handler, false, mw...)
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
func (r *RouterImpl) GET(path string, handler any, mw ...any) Router {
	return r.handle("GET", path, handler, false, mw...)
}

// GetMiddleware implements Router.
func (r *RouterImpl) GetMiddleware() []*meta.MiddlewareMeta {
	mwf := make([]*meta.MiddlewareMeta, len(r.meta.Middleware))
	copy(mwf, r.meta.Middleware)
	return mwf
}

// Group implements Router.
func (r *RouterImpl) Group(prefix string, mw ...any) Router {
	r.mwLocked = true

	rm := meta.NewRouter()
	rm.Prefix = r.cleanPrefix(prefix)

	for _, m := range mw {
		rm.UseMiddleware(m)
	}

	r.meta.Groups = append(r.meta.Groups, rm)
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
func (r *RouterImpl) Handle(method iface.HTTPMethod, path string, handler any,
	mw ...any) Router {
	return r.handle(method, path, handler, false, mw...)
}

func (r *RouterImpl) handle(method iface.HTTPMethod, path string, handler any,
	overrideMiddleware bool, mw ...any) Router {
	r.mwLocked = true

	cleanPath := r.cleanPrefix(path)

	var handlerMeta *meta.HandlerMeta

	switch h := handler.(type) {
	case request.HandlerFunc:
		handlerMeta = &meta.HandlerMeta{HandlerFunc: h}
	case string:
		handlerMeta = &meta.HandlerMeta{Name: h}
	case *meta.HandlerMeta:
		handlerMeta = h
	default:
		fmt.Printf("Handler type: %T\n", handler)
		panic("Invalid handler type, must be a RequestHandler, string, or HandlerMeta")
	}

	if overrideMiddleware {
		r.meta.HandleWithOverrideMiddleware(method, cleanPath, handlerMeta, mw...)
	} else {
		r.meta.Handle(method, cleanPath, handlerMeta, mw...)
	}

	return r
}

// HandleOverrideMiddleware implements Router.
func (r *RouterImpl) HandleOverrideMiddleware(method iface.HTTPMethod, path string,
	handler any, mw ...any) Router {
	return r.handle(method, path, handler, true, mw...)
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
func (r *RouterImpl) PATCH(path string, handler any, mw ...any) Router {
	return r.handle("PATCH", path, handler, false, mw...)
}

// POST implements Router.
func (r *RouterImpl) POST(path string, handler any, mw ...any) Router {
	return r.handle("POST", path, handler, false, mw...)
}

// PUT implements Router.
func (r *RouterImpl) PUT(path string, handler any, mw ...any) Router {
	return r.handle("PUT", path, handler, false, mw...)
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
func (r *RouterImpl) Use(mw any) Router {
	if r.mwLocked {
		panic("Cannot add middleware after locking the router")
	}

	mwd := &meta.MiddlewareMeta{}

	switch m := mw.(type) {
	case iface.MiddlewareFunc:
		mwd.MiddlewareFunc = m
	case string:
		mwd.MiddlewareType = m
	case *meta.MiddlewareMeta:
		mwd = m
	default:
		panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareMeta")
	}

	r.meta.Middleware = append(r.meta.Middleware, mwd)
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

func (r *RouterImpl) handleRouteMeta(route *meta.RouteMeta, mwParent []iface.MiddlewareFunc) {
	mwh := make([]iface.MiddlewareFunc, len(route.Middleware))
	for i, m := range route.Middleware {
		mwh[i] = m.MiddlewareFunc
	}

	if !route.OverrideMiddleware {
		mwh = slices.Concat(mwParent, mwh)
	}

	handler_with_mw := composeMiddleware(mwh, route.Handler.HandlerFunc)
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx, ok := request.ContextFromRequest(req)

		var cancel func()
		if !ok {
			ctx, cancel = request.NewContext(w, req)
			defer cancel()
		}
		if err := handler_with_mw(ctx); err != nil {
			ctx.ErrorInternal(err.Error())
		}
		if err := ctx.Err(); err != nil {
			ctx.ErrorInternal("Request aborted")
		}
		ctx.Response.WriteHttp(ctx.Writer)
	})
	r.r_engine.HandleMethod(string(route.Method), route.Path, finalHandler)
}

func (r *RouterImpl) buildRouter(router *meta.RouterMeta, mwParent []iface.MiddlewareFunc) {
	mwh := make([]iface.MiddlewareFunc, len(router.Middleware))
	for i, m := range router.Middleware {
		mwh[i] = m.MiddlewareFunc
	}
	if !router.OverrideMiddleware {
		mwh = slices.Concat(mwParent, mwh)
	}

	for _, route := range router.Routes {
		r.handleRouteMeta(route, mwh)
	}

	for _, gr := range router.Groups {
		r.buildRouter(gr, mwh)
	}
}

func (r *RouterImpl) BuildRouter() {
	mw := make([]iface.MiddlewareFunc, len(r.meta.Middleware))

	for i, m := range r.meta.Middleware {
		mw[i] = m.MiddlewareFunc
	}
	r.buildRouter(r.meta, mw)
}

func composeMiddleware(mw []iface.MiddlewareFunc,
	finalHandler request.HandlerFunc) request.HandlerFunc {
	handler := finalHandler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i](handler)
	}
	return handler
}

func init() {
	mime.AddExtensionType(".wasm", "application/wasm")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	mime.AddExtensionType(".gz", "application/gzip")
	mime.AddExtensionType(".map", "application/json")
}
