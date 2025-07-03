package meta

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/common/permission"
	"lokstra/common/registry"
	"lokstra/core"
	"lokstra/core/request"
	"lokstra/core/router"
	"net/http"

	"github.com/valyala/fasthttp"
)

type RouteMeta struct {
	path               string
	method             iface.HTTPMethod
	handler            *HandlerMeta
	overrideMiddleware bool
	middleware         []*MiddlewareMeta
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
	OverrideMiddleware bool
	Routes             []*RouteMeta
	Middleware         []*MiddlewareMeta
	StaticMounts       []*StaticDirMeta
	SPAMounts          []*SPADirMeta
	ReverseProxies     []*ReverseProxyMeta
	Groups             []*RouterMeta
}

func NewRouterInfo() *RouterMeta {
	return &RouterMeta{
		Prefix:             "",
		OverrideMiddleware: false,
		Routes:             []*RouteMeta{},
		Middleware:         []*MiddlewareMeta{},
		StaticMounts:       []*StaticDirMeta{},
		SPAMounts:          []*SPADirMeta{},
		ReverseProxies:     []*ReverseProxyMeta{},
		Groups:             []*RouterMeta{},
	}
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
		path:               path,
		method:             method,
		handler:            handlerInfo,
		overrideMiddleware: overrideMiddleware,
		middleware:         mwp,
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
	group := NewRouterInfo().WithPrefix(prefix)
	for _, mw := range middleware {
		group.UseMiddleware(mw)
	}

	r.Groups = append(r.Groups, group)
	return group
}

func (r *RouterMeta) GroupBlock(prefix string, fn func(gr *RouterMeta)) *RouterMeta {
	group := NewRouterInfo().WithPrefix(prefix)
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

func (r *RouterMeta) buildNetHttpRouter() router.Router {
	allAccessLic := createAllAccessLicense()

	// 1. Run All Registered Modules
	// Skip this step for now, as it is not implemented.
	// 2. Lock GlobalSettings

	permission.LockGlobalAccess()

	// 3. Run All Registered Plugins
	// Skip this step for now, as it is not implemented.
	// 4. Resolve all Named Handler, Middleware, Services
	resolveAllNamed(r, allAccessLic)
	// 5. Create Router based on RouterInfo
	return createRouter(r)
}

func (r *RouterMeta) CreateNetHttpRouter() router.Router {
	return r.buildNetHttpRouter()
}

func (r *RouterMeta) CreateFastHttpHandler() fasthttp.RequestHandler {
	return r.buildNetHttpRouter().FastHttpHandler()
}

// resolveAllNamed resolves all named handlers and middleware in the RouterInfo.
func resolveAllNamed(r *RouterMeta, lic *permission.PermissionLicense) {
	for _, route := range r.Routes {
		if route.handler.HandlerFunc == nil {
			handler, exists := registry.GetHandler(route.handler.Name, lic)
			if !exists {
				panic(fmt.Sprintf("Handler '%s' not found", route.handler.Name))
			}
			route.handler.HandlerFunc = handler.HandlerFunc
		}
	}

	// for _, mw := range r.Middleware {
	// 	if mw.MiddlewareFunc == nil {
	// 		middleware, exists := registry.GetMiddleware(mw.Name)
	// 		if !exists {
	// 			panic(fmt.Sprintf("Middleware '%s' not found", mw.Name))
	// 		}
	// 		mw.MiddlewareFunc = middleware.MiddlewareFunc
	// 	}
	// }

	for _, gr := range r.Groups {
		resolveAllNamed(gr, lic)
	}
}

func createAllAccessLicense() *permission.PermissionLicense {
	return permission.NewPermissionLicense(&permission.PermissionRegistration{
		ModuleName:              "lokstra",
		AllowRegisterHandler:    true,
		AllowRegisterMiddleware: true,
		AllowRegisterService:    true,
		WhitelistGetHandler:     []string{"*"}, // Allow all handlers to be accessed
		WhitelistGetServices:    []string{"*"}, // Allow all services to be accessed
	})
}

func createRouter(r *RouterMeta) router.Router {
	routerInstance := core.NewRouter().WithPrefix(r.Prefix).
		WithOverrideMiddleware(r.OverrideMiddleware)

	for _, m := range r.Middleware {
		routerInstance.Use(m.MiddlewareFunc)
	}

	for _, route := range r.Routes {
		mw := make([]iface.MiddlewareFunc, len(route.middleware))
		for i, m := range route.middleware {
			mw[i] = m.MiddlewareFunc
		}

		if route.overrideMiddleware {
			routerInstance.HandleOverrideMiddleware(route.method, route.path,
				route.handler.HandlerFunc, mw...)
		} else {
			routerInstance.Handle(route.method, route.path,
				route.handler.HandlerFunc, mw...)
		}
	}
	routerInstance.DumpRoutes()
	return routerInstance
}
