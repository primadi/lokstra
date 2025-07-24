package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/meta"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

type RouteMeta struct {
	Path               string
	Method             request.HTTPMethod
	Handler            *meta.HandlerMeta
	OverrideMiddleware bool
	Middleware         []*midware.Execution
}

type StaticDirMeta struct {
	Prefix string
	Folder http.Dir
}

type SPADirMeta struct {
	Prefix       string
	FallbackFile string
}

type ReverseProxyMeta struct {
	Prefix string
	Target string
}

type RouterMeta struct {
	Prefix             string
	RouterEngineType   string
	OverrideMiddleware bool

	Routes         []*RouteMeta
	Middleware     []*midware.Execution
	StaticMounts   []*StaticDirMeta
	SPAMounts      []*SPADirMeta
	ReverseProxies []*ReverseProxyMeta
	Groups         []*RouterMeta
}

func NewRouterMeta() *RouterMeta {
	return &RouterMeta{
		Prefix:             "/",
		OverrideMiddleware: false,
		RouterEngineType:   DEFAULT_ROUTER_ENGINE_NAME,
		Routes:             []*RouteMeta{},
		Middleware:         []*midware.Execution{},
		StaticMounts:       []*StaticDirMeta{},
		SPAMounts:          []*SPADirMeta{},
		ReverseProxies:     []*ReverseProxyMeta{},
		Groups:             []*RouterMeta{},
	}
}

func (r *RouterMeta) DumpRoutes() {
	r.RecurseAllHandler(func(rt *RouteMeta) {
		fmt.Printf("[ROUTE] %s %s\n", rt.Method, rt.Path)
	})
}

func (r *RouterMeta) RecurseAllHandler(callback func(rt *RouteMeta)) {
	for _, route := range r.Routes {
		callback(route)
	}
	for _, group := range r.Groups {
		group.RecurseAllHandler(callback)
	}
}

func (r *RouterMeta) GetRouterEngineType() string {
	return r.RouterEngineType
}

func (r *RouterMeta) WithRouterEngineType(engineType string) *RouterMeta {
	r.RouterEngineType = engineType
	return r
}

func (r *RouterMeta) Handle(method request.HTTPMethod, path string, handler any, middleware ...any) *RouterMeta {
	return r.handle(method, path, handler, false, middleware...)
}

func (r *RouterMeta) HandleWithOverrideMiddleware(method request.HTTPMethod, path string, handler any,
	middleware ...any) *RouterMeta {
	return r.handle(method, path, handler, true, middleware...)
}

func (r *RouterMeta) handle(method request.HTTPMethod, path string, handler any,
	overrideMiddleware bool, middleware ...any) *RouterMeta {
	var handlerInfo *meta.HandlerMeta

	switch h := handler.(type) {
	case request.HandlerFunc:
		handlerInfo = &meta.HandlerMeta{HandlerFunc: h}
	case string:
		handlerInfo = &meta.HandlerMeta{Name: h}
	case *meta.HandlerMeta:
		handlerInfo = h
	default:
		fmt.Printf("Handler type: %T\n", handler)
		panic("Invalid handler type, must be a RequestHandler, string, or HandlerInfo")
	}

	mwp := make([]*midware.Execution, len(middleware))
	for i := range middleware {
		if middleware[i] == nil {
			continue
		}

		var mw *midware.Execution
		switch m := middleware[i].(type) {
		case midware.Func:
			mw = midware.NewExecution(m)
		case string:
			mw = midware.Named(m)
		case *midware.Execution:
			mw = m
		default:
			panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareExecution")
		}

		mwp[i] = mw
	}

	r.Routes = append(r.Routes, &RouteMeta{
		Path:               r.cleanPrefix(path),
		Method:             method,
		Handler:            handlerInfo,
		OverrideMiddleware: overrideMiddleware,
		Middleware:         mwp,
	})
	return r
}

func (r *RouterMeta) UseMiddleware(middleware any) *RouterMeta {
	var mw *midware.Execution

	switch m := middleware.(type) {
	case midware.Func:
		mw = midware.NewExecution(m)
	case string:
		mw = midware.Named(m)
	case *midware.Execution:
		mw = m
	default:
		panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareExecution")
	}

	r.Middleware = append(r.Middleware, mw)
	return r
}

func (r *RouterMeta) cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return r.Prefix
	}

	if r.Prefix == "/" {
		return "/" + strings.Trim(prefix, "/")
	}
	return r.Prefix + "/" + strings.Trim(prefix, "/")
}

func ResolveAllNamed(ctx registration.Context, r *RouterMeta) {
	for i, route := range r.Routes {
		if rpcServiceMeta, ok := route.Handler.Extension.(*meta.RpcServiceMeta); ok {
			svc := rpcServiceMeta.ServiceInst
			if svc == nil {
				svc = ctx.GetService(rpcServiceMeta.ServiceURI)
				if svc == nil {
					panic(fmt.Sprintf("Rpc Service '%s' not found", rpcServiceMeta.ServiceURI))
				}
				r.Routes[i].Handler.Extension.(*meta.RpcServiceMeta).ServiceInst = svc
			}
			rpcService := ctx.GetService("lokstra://rpc_server/default").(serviceapi.RpcServer)
			if rpcService != nil {
				r.Routes[i].Handler.HandlerFunc = func(ctx *request.Context) error {
					methodParam := ctx.GetPathParam(rpcServiceMeta.MethodParam)
					return rpcService.HandleRequest(ctx, svc, methodParam)
				}
			}
		} else if route.Handler.HandlerFunc == nil {
			handler := ctx.GetHandler(route.Handler.Name)
			if handler == nil {
				panic(fmt.Sprintf("Handler '%s' not found", route.Handler.Name))
			}
			route.Handler.HandlerFunc = handler.HandlerFunc
		}
		for _, mwExec := range route.Middleware {
			resolveMiddleware(ctx, mwExec)
		}
	}

	for _, mwExec := range r.Middleware {
		resolveMiddleware(ctx, mwExec)
	}

	for _, gr := range r.Groups {
		ResolveAllNamed(ctx, gr)
	}
}

func resolveMiddleware(ctx registration.Context, mw *midware.Execution) {
	if mw.MiddlewareFn == nil {
		mwFactory, priority, found := ctx.GetMiddlewareFactory(mw.Name)
		if !found {
			panic(fmt.Sprintf("Middleware factory '%s' not found", mw.Name))
		}
		mw.MiddlewareFn = mwFactory(mw.Config)
		mw.Priority = priority
	}
}
