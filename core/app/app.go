package app

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/meta"
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
	return &App{
		ctx:      ctx,
		meta:     meta,
		listener: getListenerFromMeta(ctx, meta),
		Router:   router.NewRouterWithEngine(ctx, meta.GetRouterEngineType()),
	}
}

func (a *App) GetName() string {
	return a.meta.GetName()
}

func (a *App) GetPort() int {
	return a.meta.GetPort()
}

func (a *App) Start() error {
	meta.ResolveAllNamed(a.ctx, a.meta.RouterMeta)
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
