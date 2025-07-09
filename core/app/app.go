package app

import (
	"fmt"
	"lokstra/common/component"
	"lokstra/common/meta"
	"lokstra/core/router"
	"lokstra/modules/coreservice_module"
	"lokstra/serviceapi/core_service"
	"maps"
	"time"
)

type App struct {
	router.Router

	ctx      component.ComponentContext
	listener core_service.HttpListener

	name              string
	addr              string
	listenerType      string
	routingEngineType string
	settings          map[string]any
}

func NewApp(ctx component.ComponentContext, name string, addr string) *App {
	listenerType := core_service.DEFAULT_LISTENER_NAME
	routerEngineType := core_service.DEFAULT_ROUTER_ENGINE_NAME
	return NewAppCustom(ctx, name, addr, listenerType, routerEngineType, nil)
}

func NewAppCustom(ctx component.ComponentContext, name string, addr string,
	listenerType string, routerEngineType string, settings map[string]any) *App {
	ctx.RegisterModule("coreservice_module", coreservice_module.ModuleRegister) // no problem if already registered

	if settings == nil {
		settings = make(map[string]any)
	}

	return &App{
		ctx:  ctx,
		name: name,
		addr: addr,

		listenerType: listenerType,
		listener:     router.NewListenerWithEngine(ctx, listenerType, name, settings),

		routingEngineType: routerEngineType,
		Router:            router.NewRouterWithEngine(ctx, routerEngineType, name, settings),

		settings: maps.Clone(settings),
	}
}

func (a *App) GetName() string {
	return a.name
}

func (a *App) GetAddr() string {
	return a.addr
}

func (a *App) Start() error {
	meta.ResolveAllNamed(a.ctx, a.Router.GetMeta())
	rImp := a.Router.(*router.RouterImpl)
	rImp.BuildRouter()
	return a.listener.ListenAndServe(a.addr, a.Router)
}

func (a *App) Shutdown(shutdownTimeout time.Duration) error {
	if a.listener == nil {
		return fmt.Errorf("listener is not initialized")
	}
	return a.listener.Shutdown(shutdownTimeout)
}

func (a *App) GetSettings() map[string]any {
	return a.settings
}

func (a *App) GetSetting(key string) (any, bool) {
	val, found := a.settings[key]
	return val, found
}

func (a *App) SetSetting(key string, value any) {
	a.settings[key] = value
}
