package router

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"slices"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type RouterImpl struct {
	meta     *RouterMeta
	r_engine serviceapi.RouterEngine
}

// RawHandle implements Router.
func (r *RouterImpl) RawHandle(prefix string, stripPrefix bool, handler http.Handler) Router {
	r.meta.RawHandles = append(r.meta.RawHandles, &RawHandleMeta{
		Prefix:  prefix,
		Handler: handler,
		Strip:   stripPrefix,
	})
	return r
}

func NewListener(ctx registration.Context, config map[string]any) serviceapi.HttpListener {
	return NewListenerWithEngine(ctx, "", config)
}

func NewListenerWithEngine(ctx registration.Context, listenerType string,
	config map[string]any) serviceapi.HttpListener {

	lType := NormalizeListenerType(listenerType)

	factory, found := ctx.GetServiceFactory(lType)
	if !found {
		panic(fmt.Sprintf("Listener type %s not found", lType))
	}

	lsAny, err := factory(config)
	if err != nil {
		panic(fmt.Sprintf("failed to create listener for app %s: %v", lType, err))
	}
	ls, ok := lsAny.(serviceapi.HttpListener)
	if !ok {
		panic(fmt.Sprintf("listener for app %s is not of type serviceapi.HttpListener", lType))
	}
	return ls
}

func NewRouter(regCtx registration.Context, config map[string]any) Router {
	return NewRouterWithEngine(regCtx, "", config)
}

func NewRouterWithEngine(regCtx registration.Context, engineType string,
	config map[string]any) Router {

	serviceType := NormalizeRouterType(engineType)

	factory, found := regCtx.GetServiceFactory(serviceType)
	if !found {
		panic(fmt.Sprintf("Router engine %s not found", serviceType))
	}

	rtAny, err := factory(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create router engine %s: %v", engineType, err))
	}
	rt := rtAny.(serviceapi.RouterEngine)
	if rt == nil {
		panic(fmt.Sprintf("Router engine %s is not initialized", engineType))
	}
	return &RouterImpl{
		meta:     NewRouterMeta(),
		r_engine: rt,
	}
}

// GetMeta implements Router.
func (r *RouterImpl) GetMeta() *RouterMeta {
	return r.meta
}

// DELETE implements Router.
func (r *RouterImpl) DELETE(path string, handler any, mw ...any) Router {
	return r.Handle("DELETE", path, handler, mw...)
}

// DumpRoutes implements Router.
func (r *RouterImpl) DumpRoutes() {
	r.meta.DumpRoutes()
}

// AddRouter implements Router.
func (r *RouterImpl) AddRouter(other Router) Router {
	otherMeta := other.GetMeta()

	// Merge all routes
	r.meta.Routes = append(r.meta.Routes, otherMeta.Routes...)

	// Merge all groups
	r.meta.Groups = append(r.meta.Groups, otherMeta.Groups...)

	// Merge static mounts
	r.meta.StaticMounts = append(r.meta.StaticMounts, otherMeta.StaticMounts...)

	// Merge reverse proxies
	r.meta.ReverseProxies = append(r.meta.ReverseProxies, otherMeta.ReverseProxies...)

	// Merge raw handles
	r.meta.RawHandles = append(r.meta.RawHandles, otherMeta.RawHandles...)

	// Merge RPC handles
	r.meta.RPCHandles = append(r.meta.RPCHandles, otherMeta.RPCHandles...)

	// Middleware: TIDAK di-merge secara otomatis untuk menghindari efek samping
	// Middleware dari router lain tetap melekat pada route/group asalnya
	// Jika ingin merge middleware, gunakan Use() secara eksplisit setelah AddRouter()

	return r
}

// FastHttpHandler implements Router.
func (r *RouterImpl) FastHttpHandler() fasthttp.RequestHandler {
	return fasthttpadaptor.NewFastHTTPHandler(r.r_engine)
}

// GET implements Router.
func (r *RouterImpl) GET(path string, handler any, mw ...any) Router {
	return r.Handle("GET", path, handler, mw...)
}

// GetMiddleware implements Router.
func (r *RouterImpl) GetMiddleware() []*midware.Execution {
	mwf := make([]*midware.Execution, len(r.meta.Middleware))
	copy(mwf, r.meta.Middleware)
	return mwf
}

// Group implements Router.
func (r *RouterImpl) Group(prefix string, mw ...any) Router {
	rm := NewRouterMeta()
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
func (r *RouterImpl) Handle(method request.HTTPMethod, path string, handler any,
	mw ...any) Router {
	r.meta.Handle(method, r.cleanPrefix(path), handler, false, mw...)
	return r
}

// HandleOverrideMiddleware implements Router.
func (r *RouterImpl) HandleOverrideMiddleware(method request.HTTPMethod, path string,
	handler any, mw ...any) Router {
	r.meta.Handle(method, r.cleanPrefix(path), handler, true, mw...)
	return r
}

// MountReverseProxy implements Router.
func (r *RouterImpl) MountReverseProxy(prefix string, target string,
	overrideMiddleware bool, mw ...any) Router {
	r.meta.MountReverseProxy(prefix, target, overrideMiddleware, mw...)
	return r
}

// MountStatic implements Router.
func (r *RouterImpl) MountStatic(prefix string, spa bool, sources ...fs.FS) Router {
	r.meta.MountStatic(prefix, spa, sources...)
	return r
}

// MountHtmx implements Router.
func (r *RouterImpl) MountHtmx(prefix string, sources ...fs.FS) Router {
	r.meta.MountHtmx(prefix, sources...)
	return r
}

// MountRpcService implements Router.
func (r *RouterImpl) MountRpcService(path string, svc any, overrideMiddleware bool, mw ...any) Router {
	r.meta.MountRpcService(path, svc, overrideMiddleware, mw...)
	return r
}

// OverrideMiddleware implements Router.
func (r *RouterImpl) OverrideMiddleware() bool {
	return r.meta.OverrideMiddleware
}

// PATCH implements Router.
func (r *RouterImpl) PATCH(path string, handler any, mw ...any) Router {
	return r.Handle("PATCH", path, handler, mw...)
}

// POST implements Router.
func (r *RouterImpl) POST(path string, handler any, mw ...any) Router {
	return r.Handle("POST", path, handler, mw...)
}

// PUT implements Router.
func (r *RouterImpl) PUT(path string, handler any, mw ...any) Router {
	return r.Handle("PUT", path, handler, mw...)
}

// Prefix implements Router.
func (r *RouterImpl) Prefix() string {
	return r.meta.Prefix
}

// RecurseAllHandler implements Router.
func (r *RouterImpl) RecurseAllHandler(callback func(rt *RouteMeta)) {
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
	r.meta.UseMiddleware(mw)
	return r
}

// WithOverrideMiddleware implements Router.
func (r *RouterImpl) WithOverrideMiddleware(enable bool) Router {
	r.meta.OverrideMiddleware = enable
	return r
}

// WithPrefix implements Router.
func (r *RouterImpl) WithPrefix(prefix string) Router {
	if prefix == "/" || prefix == "" {
		return r
	}

	if strings.HasPrefix(prefix, "/") {
		r.meta.Prefix = "/" + strings.Trim(prefix, "/") // replace absolute prefix
	} else {
		r.meta.Prefix = r.cleanPrefix(prefix) // add relative prefix
	}
	return r
}

var _ Router = (*RouterImpl)(nil)

func (r *RouterImpl) cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return r.meta.Prefix
	}

	cleaned := strings.Trim(prefix, "/")

	var result string
	if strings.HasSuffix(r.meta.Prefix, "/") {
		result = r.meta.Prefix + cleaned
	} else {
		result = r.meta.Prefix + "/" + cleaned
	}

	if strings.HasSuffix(prefix, "/") {
		result += "/"
	}

	return result
}

func (r *RouterImpl) handleRouteMeta(route *RouteMeta, mwParent []*midware.Execution) {
	var mwh []*midware.Execution

	if route.OverrideMiddleware {
		mwh = make([]*midware.Execution, len(route.Middleware))
		copy(mwh, route.Middleware)
	} else {
		mwh = utils.SlicesConcat(mwParent, route.Middleware)
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
			_ = ctx.ErrorInternal(err.Error())
		}
		if err := ctx.Err(); err != nil {
			_ = ctx.ErrorInternal("Request aborted")
		}
		_ = ctx.Response.WriteHttp(ctx.Writer)
	})

	r.r_engine.HandleMethod(string(route.Method), route.Path, finalHandler)
}

func (r *RouterImpl) buildRouter(router *RouterMeta, mwParent []*midware.Execution) {
	var mwh []*midware.Execution

	if router.OverrideMiddleware {
		mwh = make([]*midware.Execution, len(router.Middleware))
		copy(mwh, router.Middleware)
	} else {
		mwh = utils.SlicesConcat(mwParent, router.Middleware)
	}

	for _, route := range router.Routes {
		r.handleRouteMeta(route, mwh)
	}

	for _, gr := range router.Groups {
		r.buildRouter(gr, mwh)
	}

	for _, rh := range router.RawHandles {
		if rh.Strip {
			r.r_engine.RawHandle(rh.Prefix, http.StripPrefix(rh.Prefix, rh.Handler))
		} else {
			r.r_engine.RawHandle(rh.Prefix, rh.Handler)
		}
	}

	for _, rpc := range router.RPCHandles {
		cleanPath := r.cleanPrefix(rpc.Path)
		if strings.HasSuffix(cleanPath, "/") {
			cleanPath += ":method"
		} else {
			cleanPath += "/:method"
		}

		rpcMeta := &service.RpcServiceMeta{
			MethodParam: "method",
		}
		switch s := rpc.Service.(type) {
		case string:
			rpcMeta.ServiceName = s
		case *service.RpcServiceMeta:
			rpcMeta = s
		case service.Service:
			rpcMeta.ServiceInst = s
		default:
			fmt.Printf("Service type: %T\n", rpc.Service)
			panic("Invalid service type, must be a string, *RpcServiceMeta, or iface.Service")
		}

		handlerMeta := &request.HandlerMeta{
			HandlerFunc: func(ctx *request.Context) error {
				return ctx.ErrorInternal("RpcService not yet resolved")
			},
			Extension: rpcMeta,
		}

		if rpc.OverrideMiddleware {
			r.meta.Handle("POST", cleanPath, handlerMeta, true, rpc.Middleware...)
		} else {
			r.Handle("POST", cleanPath, handlerMeta, rpc.Middleware...)
		}
	}

	for _, rp := range router.ReverseProxies {
		handler := composeReverseProxyMw(rp, mwh)
		r.r_engine.ServeReverseProxy(rp.Prefix, handler)
	}

	for _, sdf := range router.StaticMounts {
		r.r_engine.ServeStatic(sdf.Prefix, sdf.Spa, sdf.Sources...)
	}

	for _, htmx := range router.HTMXPages {
		r.r_engine.ServeHtmxPage(r.r_engine, htmx.Prefix, htmx.Sources...)
	}
}

func (r *RouterImpl) BuildRouter() {
	r.buildRouter(r.meta, nil)
}

func composeMiddleware(mw []*midware.Execution,
	finalHandler request.HandlerFunc) request.HandlerFunc {
	// Update execution order based on order of addition
	execOrder := 0
	for _, m := range mw {
		m.ExecutionOrder = execOrder
		execOrder++
	}

	// Sort middleware by priority and execution order
	slices.SortStableFunc(mw, func(a, b *midware.Execution) int {
		aOrder := a.Priority + a.ExecutionOrder
		bOrder := b.Priority + b.ExecutionOrder

		if aOrder < bOrder {
			return -1
		} else if aOrder > bOrder {
			return 1
		}

		return 0
	})

	// Compose middleware functions in reverse order
	handler := finalHandler
	for i := len(mw) - 1; i >= 0; i-- {
		handler = mw[i].MiddlewareFn(handler)
	}
	return handler
}

// ComposeMiddlewareForTest exposes composeMiddleware for testing
func ComposeMiddlewareForTest(mw []*midware.Execution, finalHandler request.HandlerFunc) request.HandlerFunc {
	return composeMiddleware(mw, finalHandler)
}

func NormalizeListenerType(listenerType string) string {
	if listenerType == "" {
		listenerType = "default"
	}

	if !strings.HasPrefix(listenerType, serviceapi.HTTP_LISTENER_PREFIX) {
		listenerType = serviceapi.HTTP_LISTENER_PREFIX + listenerType
	}

	return listenerType
}

func NormalizeRouterType(routerType string) string {
	if routerType == "" {
		routerType = "default"
	}

	if !strings.HasPrefix(routerType, serviceapi.HTTP_ROUTER_PREFIX) {
		routerType = serviceapi.HTTP_ROUTER_PREFIX + routerType
	}

	return routerType
}

func init() {
	_ = mime.AddExtensionType(".wasm", "application/wasm")
	_ = mime.AddExtensionType(".woff2", "font/woff2")
	_ = mime.AddExtensionType(".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	_ = mime.AddExtensionType(".gz", "application/gzip")
	_ = mime.AddExtensionType(".map", "application/json")
}
