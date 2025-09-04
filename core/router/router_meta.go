package router

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

type RouteMeta struct {
	Path               string
	Method             request.HTTPMethod
	Handler            *request.HandlerMeta
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
	Prefix             string
	Target             string
	OverrideMiddleware bool
	Middleware         []*midware.Execution
}

type RouterMeta struct {
	Prefix             string
	OverrideMiddleware bool

	Routes     []*RouteMeta
	Middleware []*midware.Execution
	// StaticMounts   []*StaticDirMeta
	// SPAMounts      []*SPADirMeta
	ReverseProxies []*ReverseProxyMeta
	Groups         []*RouterMeta
}

func NewRouterMeta() *RouterMeta {
	return &RouterMeta{
		Prefix:             "/",
		OverrideMiddleware: false,
		Routes:             []*RouteMeta{},
		Middleware:         []*midware.Execution{},
		// StaticMounts:       []*StaticDirMeta{},
		// SPAMounts:          []*SPADirMeta{},
		ReverseProxies: []*ReverseProxyMeta{},
		Groups:         []*RouterMeta{},
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

func (r *RouterMeta) Handle(method request.HTTPMethod, path string, handler any, middleware ...any) *RouterMeta {
	return r.handle(method, path, handler, false, middleware...)
}

func (r *RouterMeta) HandleWithOverrideMiddleware(method request.HTTPMethod, path string, handler any,
	middleware ...any) *RouterMeta {
	return r.handle(method, path, handler, true, middleware...)
}

func (r *RouterMeta) handle(method request.HTTPMethod, path string, handler any,
	overrideMiddleware bool, middleware ...any) *RouterMeta {
	var handlerInfo *request.HandlerMeta

	switch h := handler.(type) {
	case request.HandlerFunc:
		handlerInfo = &request.HandlerMeta{HandlerFunc: h}
	case string:
		handlerInfo = &request.HandlerMeta{Name: h}
	case *request.HandlerMeta:
		handlerInfo = h
	default:
		// Try to match func(ctx *request.Context, params *T) error
		fnVal := reflect.ValueOf(handler)
		fnType := fnVal.Type()

		if fnType.Kind() == reflect.Func &&
			fnType.NumIn() == 2 &&
			fnType.NumOut() == 1 &&
			fnType.In(0) == reflect.TypeOf((*request.Context)(nil)) &&
			fnType.Out(0) == reflect.TypeOf((*error)(nil)).Elem() &&
			fnType.In(1).Kind() == reflect.Ptr &&
			fnType.In(1).Elem().Kind() == reflect.Struct {

			paramType := fnType.In(1)

			wrapped := func(ctx *request.Context) error {
				paramPtr := reflect.New(paramType.Elem()).Interface()
				if err := ctx.BindAll(paramPtr); err != nil {
					return ctx.ErrorBadRequest(err.Error())
				}
				out := fnVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(paramPtr)})
				if !out[0].IsNil() {
					return out[0].Interface().(error)
				}
				return nil
			}

			handlerInfo = &request.HandlerMeta{HandlerFunc: wrapped}
		} else {
			fmt.Printf("Handler type: %T\n", handler)
			panic("Invalid handler type, must be a HandlerFunc, string, HandlerMeta, or func(ctx, params)")
		}
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
		Path:               path,
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

func (r *RouterMeta) MountReverseProxy(prefix string, target string,
	overrideMiddleware bool, middleware ...any) *RouterMeta {
	mwp := anyArraytoMiddleware(middleware)
	r.ReverseProxies = append(r.ReverseProxies, &ReverseProxyMeta{
		Prefix:             prefix,
		Target:             target,
		OverrideMiddleware: overrideMiddleware,
		Middleware:         mwp,
	})
	return r
}

func ResolveAllNamed(ctx registration.Context, r *RouterMeta) {
	for i, route := range r.Routes {
		if rpcServiceMeta, ok := route.Handler.Extension.(*service.RpcServiceMeta); ok {
			svc := rpcServiceMeta.ServiceInst
			if svc == nil {
				var err error
				svc, err = ctx.GetService(rpcServiceMeta.ServiceName)
				if err != nil {
					panic(fmt.Sprintf("Rpc Service '%s' not found", rpcServiceMeta.ServiceName))
				}
				r.Routes[i].Handler.Extension.(*service.RpcServiceMeta).ServiceInst = svc
			}
			rpcSvr, err := getOrCreateService[serviceapi.RpcServer](ctx, "rpc_server.default",
				"rpc_service.rpc_server")
			if err != nil {
				panic(fmt.Sprintf("Failed to get RPC server: %s", err.Error()))
			}

			r.Routes[i].Handler.HandlerFunc = func(ctx *request.Context) error {
				methodParam := ctx.GetPathParam(rpcServiceMeta.MethodParam)
				return rpcSvr.HandleRequest(ctx, svc, methodParam)
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

	for _, rp := range r.ReverseProxies {
		for _, mwExec := range rp.Middleware {
			resolveMiddleware(ctx, mwExec)
		}
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

func getOrCreateService[T any](ctx registration.Context,
	serviceName string, factoryName string, config ...any) (T, error) {
	svc, err := ctx.GetService(serviceName)
	if err != nil {
		svc, err = ctx.CreateService(factoryName, serviceName, config...)
		if err != nil {
			var zero T
			return zero, errors.New("failed to create service: " + err.Error())
		}
	}
	if typedSvc, ok := svc.(T); ok {
		return typedSvc, nil
	}
	var zero T
	return zero, errors.New("service type mismatch: " + serviceName)
}
