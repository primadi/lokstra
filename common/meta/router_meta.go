package meta

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/iface"
	"lokstra/core/request"
	"lokstra/serviceapi/core_service"
	"net/http"
)

type RouteMeta struct {
	Path               string
	Method             iface.HTTPMethod
	Handler            *HandlerMeta
	OverrideMiddleware bool
	Middleware         []*MiddlewareMeta
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
	Middleware     []*MiddlewareMeta
	StaticMounts   []*StaticDirMeta
	SPAMounts      []*SPADirMeta
	ReverseProxies []*ReverseProxyMeta
	Groups         []*RouterMeta
}

func NewRouter() *RouterMeta {
	return &RouterMeta{
		Prefix:             "",
		OverrideMiddleware: false,
		RouterEngineType:   core_service.DEFAULT_ROUTER_ENGINE_NAME,
		Routes:             []*RouteMeta{},
		Middleware:         []*MiddlewareMeta{},
		StaticMounts:       []*StaticDirMeta{},
		SPAMounts:          []*SPADirMeta{},
		ReverseProxies:     []*ReverseProxyMeta{},
		Groups:             []*RouterMeta{},
	}
}

func (r *RouterMeta) GetRouterEngineType() string {
	return r.RouterEngineType
}

func (r *RouterMeta) WithRouterEngineType(engineType string) *RouterMeta {
	r.RouterEngineType = engineType
	return r
}

func (r *RouterMeta) WithPrefix(prefix string) *RouterMeta {
	r.Prefix = prefix
	return r
}

func (r *RouterMeta) WithOverrideMiddleware(override bool) *RouterMeta {
	r.OverrideMiddleware = override
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

	mwp := make([]*MiddlewareMeta, len(middleware))
	for i := range middleware {
		var mw *MiddlewareMeta
		switch m := middleware[i].(type) {
		case iface.MiddlewareFunc:
			mw = &MiddlewareMeta{MiddlewareFunc: m}
		case string:
			mw = &MiddlewareMeta{MiddlewareType: m}
		case *MiddlewareMeta:
			mw = m
		default:
			panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareInfo")
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
	mw := &MiddlewareMeta{}

	switch m := middleware.(type) {
	case iface.MiddlewareFunc:
		mw.MiddlewareFunc = m
	case string:
		mw.MiddlewareType = m
	case *MiddlewareMeta:
		mw = m
	default:
		panic("Invalid middleware type, must be a MiddlewareFunc, string, or *MiddlewareMeta")
	}

	r.Middleware = append(r.Middleware, mw)
	return r
}

func (r *RouterMeta) MountStatic(prefix string, folder http.Dir) *RouterMeta {
	r.StaticMounts = append(r.StaticMounts, &StaticDirMeta{
		Prefix: prefix,
		Folder: folder,
	})
	return r
}

func (r *RouterMeta) MountSPA(prefix string, fallbackFile string) *RouterMeta {
	r.SPAMounts = append(r.SPAMounts, &SPADirMeta{
		Prefix:       prefix,
		FallbackFile: fallbackFile,
	})
	return r
}

func (r *RouterMeta) MountReverseProxy(prefix string, target string) *RouterMeta {
	r.ReverseProxies = append(r.ReverseProxies, &ReverseProxyMeta{
		Prefix: prefix,
		Target: target,
	})
	return r
}

func (r *RouterMeta) Group(prefix string, middleware ...any) *RouterMeta {
	group := NewRouter().WithPrefix(prefix)
	for _, mw := range middleware {
		group.UseMiddleware(mw)
	}

	r.Groups = append(r.Groups, group)
	return group
}

func (r *RouterMeta) GroupBlock(prefix string, fn func(gr *RouterMeta)) *RouterMeta {
	group := NewRouter().WithPrefix(prefix)
	r.Groups = append(r.Groups, group)

	if fn != nil {
		fn(group)
	}

	return r
}

func (r *RouterMeta) GET(path string, handler any, middleware ...any) *RouterMeta {
	return r.Handle("GET", path, handler, middleware...)
}

func (r *RouterMeta) POST(path string, handler any, middleware ...any) *RouterMeta {
	return r.Handle("POST", path, handler, middleware...)
}

func (r *RouterMeta) PUT(path string, handler any, middleware ...any) *RouterMeta {
	return r.Handle("PUT", path, handler, middleware...)
}

func (r *RouterMeta) PATCH(path string, handler any, middleware ...any) *RouterMeta {
	return r.Handle("PATCH", path, handler, middleware...)
}

func (r *RouterMeta) DELETE(path string, handler any, middleware ...any) *RouterMeta {
	return r.Handle("DELETE", path, handler, middleware...)
}

func ResolveAllNamed(ctx component.ComponentContext, r *RouterMeta) {
	for _, route := range r.Routes {
		if route.Handler.HandlerFunc == nil {
			handler := ctx.GetHandler(route.Handler.Name)
			if handler == nil {
				panic(fmt.Sprintf("Handler '%s' not found", route.Handler.Name))
			}
			route.Handler.HandlerFunc = handler.HandlerFunc
		}
	}

	for _, mwMeta := range r.Middleware {
		if mwMeta.MiddlewareFunc == nil {
			mwfactory, found := ctx.GetMiddlewareFactory(mwMeta.MiddlewareType)
			if !found {
				panic(fmt.Sprintf("Middleware factory '%s' not found", mwMeta.MiddlewareType))
			}
			mwf := mwfactory(mwMeta.Config)
			mwMeta.MiddlewareFunc = mwf
		}
	}

	for _, gr := range r.Groups {
		ResolveAllNamed(ctx, gr)
	}
}
