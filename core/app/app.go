package app

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/meta"
	"lokstra/common/utils"
	"lokstra/core/router"
	"lokstra/modules/coreservice_module"
	"lokstra/serviceapi/core_service"
	"time"
)

type App struct {
	router.Router

	ctx      component.ComponentContext
	listener core_service.HttpListener

	meta *meta.AppMeta // Meta information about the app
}

func NewApp(ctx component.ComponentContext, name string, port int) *App {
	return NewAppFromMeta(ctx, meta.NewApp(name, port))
}

func NewAppFromMeta(ctx component.ComponentContext, meta *meta.AppMeta) *App {
	ctx.RegisterModule("coreservice_module", coreservice_module.Register)
	rtr := router.NewRouterWithEngine(ctx, meta.GetRouterEngineType())

	copyRouterMeta(rtr, meta.RouterMeta)

	return &App{
		ctx:      ctx,
		meta:     meta,
		listener: getListenerFromMeta(ctx, meta),
		Router:   rtr,
	}
}

func copyRouterMeta(rtr router.Router, meta *meta.RouterMeta) {
	for _, route := range meta.Routes {
		if route.OverrideMiddleware {
			rtr.HandleOverrideMiddleware(route.Method, route.Path, route.Handler,
				utils.ToAnySlice(route.Middleware)...)
		} else {
			rtr.Handle(route.Method, route.Path, route.Handler,
				utils.ToAnySlice(route.Middleware)...)
		}
	}

	for _, staticMount := range meta.StaticMounts {
		rtr.MountStatic(staticMount.Prefix, staticMount.Folder)
	}

	for _, spaMount := range meta.SPAMounts {
		rtr.MountSPA(spaMount.Prefix, spaMount.FallbackFile)
	}

	for _, reverseProxy := range meta.ReverseProxies {
		rtr.MountReverseProxy(reverseProxy.Prefix, reverseProxy.Target)
	}

	for _, gr := range meta.Groups {
		rtrGr := rtr.Group(gr.Prefix, utils.ToAnySlice(gr.Middleware)...).
			WithOverrideMiddleware(gr.OverrideMiddleware)
		copyRouterMeta(rtrGr, gr)

	}
}

func (a *App) GetName() string {
	return a.meta.GetName()
}

func (a *App) GetPort() int {
	return a.meta.GetPort()
}

func (a *App) Start() error {
	rtrImpl := a.Router.(*router.RouterImpl)
	meta.ResolveAllNamed(a.ctx, rtrImpl.Meta())
	rImp := a.Router.(*router.RouterImpl)
	rImp.BuildRouter()
	return a.listener.ListenAndServe(a.meta.Addr(), a.Router)
}

func (a *App) Shutdown(shutdownTimeout time.Duration) error {
	if a.listener == nil {
		return fmt.Errorf("listener is not initialized")
	}
	return a.listener.Shutdown(shutdownTimeout)
}

func (a *App) GetListener() core_service.HttpListener {
	return a.listener
}

func getListenerFromMeta(ctx component.ComponentContext, meta *meta.AppMeta) core_service.HttpListener {
	lsAny, err := ctx.NewService(meta.GetListenerType(), meta.GetName()+".listener")
	if err != nil {
		panic(fmt.Sprintf("failed to create listener for app %s: %v", meta.GetName(), err))
	}
	ls, ok := lsAny.(core_service.HttpListener)
	if !ok {
		panic(fmt.Sprintf("listener for app %s is not of type core_service.HttpListener", meta.GetName()))
	}
	return ls
}
