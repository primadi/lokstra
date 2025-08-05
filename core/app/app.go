package app

import (
	"fmt"
	"maps"
	"time"

	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/serviceapi"
)

type App struct {
	router.Router

	ctx      iface.RegistrationContext
	listener serviceapi.HttpListener

	name              string
	addr              string
	listenerType      string
	routingEngineType string
	settings          map[string]any
}

func NewApp(ctx iface.RegistrationContext, name string, addr string) *App {
	return NewAppCustom(ctx, name, addr, "", "", nil)
}

func NewAppCustom(ctx iface.RegistrationContext, name string, addr string,
	listenerType string, routerEngineType string, settings map[string]any) *App {

	if settings == nil {
		settings = make(map[string]any)
	}

	lType := router.NormalizeListenerType(listenerType)
	rType := router.NormalizeRouterType(routerEngineType)

	return &App{
		ctx:  ctx,
		name: name,
		addr: addr,

		listenerType: lType,
		listener:     router.NewListenerWithEngine(ctx, lType, settings),

		routingEngineType: rType,
		Router:            router.NewRouterWithEngine(ctx, rType, settings),

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
	router.ResolveAllNamed(a.ctx, a.Router.GetMeta())
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
