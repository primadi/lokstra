package meta

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/common/module"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

type RouteMeta struct {
	Path               string
	Method             iface.HTTPMethod
	Handler            *HandlerMeta
	OverrideMiddleware bool
	Middleware         []*MiddlewareExecution
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
	Middleware     []*MiddlewareExecution
	StaticMounts   []*StaticDirMeta
	SPAMounts      []*SPADirMeta
	ReverseProxies []*ReverseProxyMeta
	Groups         []*RouterMeta
}

func NewRouter() *RouterMeta {
	return &RouterMeta{
		Prefix:             "",
		OverrideMiddleware: false,
		RouterEngineType:   serviceapi.DEFAULT_ROUTER_ENGINE_NAME,
		Routes:             []*RouteMeta{},
		Middleware:         []*MiddlewareExecution{},
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

func (r *RouterMeta) Handle(method iface.HTTPMethod, path string, handler any, middleware ...any) *RouterMeta {
	return r.handle(method, path, handler, false, middleware...)
}

func (r *RouterMeta) HandleWithOverrideMiddleware(method iface.HTTPMethod, path string, handler any,
	middleware ...any) *RouterMeta {
	return r.handle(method, path, handler, true, middleware...)
}

func (r *RouterMeta) handle(method iface.HTTPMethod, path string, handler any,
	overrideMiddleware bool, middleware ...any) *RouterMeta {
	var handlerInfo *HandlerMeta

	switch h := handler.(type) {
	case request.HandlerFunc:
		handlerInfo = &HandlerMeta{HandlerFunc: h}
	case string:
		handlerInfo = &HandlerMeta{Name: h}
	case *HandlerMeta:
		handlerInfo = h
	default:
		fmt.Printf("Handler type: %T\n", handler)
		panic("Invalid handler type, must be a RequestHandler, string, or HandlerInfo")
	}

	mwp := make([]*MiddlewareExecution, len(middleware))
	for i := range middleware {
		if middleware[i] == nil {
			continue
		}

		var mw *MiddlewareExecution
		switch m := middleware[i].(type) {
		case iface.MiddlewareFunc:
			mw = MiddlewareFn(m)
		case string:
			mw = NamedMiddleware(m)
		case *MiddlewareExecution:
			mw = m
		default:
			panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareExecution")
		}

		mwp[i] = mw
	}

	r.Routes = append(r.Routes, &RouteMeta{
		Path:               path,
		Method:             method,
		Handler:            handlerInfo,
		OverrideMiddleware: overrideMiddleware,
		Middleware:         mwp,
	})
	return r
}

func (r *RouterMeta) UseMiddleware(middleware any) *RouterMeta {
	var mw *MiddlewareExecution

	switch m := middleware.(type) {
	case iface.MiddlewareFunc:
		mw = MiddlewareFn(m)
	case string:
		mw = NamedMiddleware(m)
	case *MiddlewareExecution:
		mw = m
	default:
		panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareExecution")
	}

	r.Middleware = append(r.Middleware, mw)
	return r
}

func ResolveAllNamed(ctx module.RegistrationContext, r *RouterMeta) {
	for _, route := range r.Routes {
		if route.Handler.HandlerFunc == nil {
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

func resolveMiddleware(ctx module.RegistrationContext, mw *MiddlewareExecution) {
	if mw.MiddlewareFn == nil {
		mwModule, found := ctx.GetMiddlewareModule(mw.Name)
		if !found {
			panic(fmt.Sprintf("Middleware factory '%s' not found", mw.Name))
		}
		mw.MiddlewareFn = mwModule.Factory(mw.Config)
		mw.Priority = mwModule.Meta().Priority
	}
}
