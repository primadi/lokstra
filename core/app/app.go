package app

import (
	"fmt"
	"maps"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/serviceapi"
)

type App struct {
	router.Router

	ctx      registration.Context
	listener serviceapi.HttpListener

	name              string
	addr              string
	listenerType      string
	routingEngineType string
	settings          map[string]any
	merged            bool
	mergedRoutes      []router.Router
}

func NewApp(ctx registration.Context, name string, addr string) *App {
	return NewAppCustom(ctx, name, addr, "", "", nil)
}

func NewAppCustom(ctx registration.Context, name string, addr string,
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
	if a.merged {
		return nil
	}
	if err := a.BuildRouter(); err != nil {
		return err
	}

	return a.ListenAndServe()
}

func (a *App) BuildRouter() error {
	if a.merged {
		return nil
	}

	router.ResolveAllNamed(a.ctx, a.Router.GetMeta())
	rImp := a.Router.(*router.RouterImpl)
	rImp.BuildRouter()
	r_engine := rImp.GetEngine()
	for _, or := range a.mergedRoutes {
		router.ResolveAllNamed(a.ctx, or.GetMeta())
		roImp := or.(*router.RouterImpl)
		roImp.SetEngine(r_engine) // share the same routing engine
		roImp.BuildRouter()
	}

	return nil
}

func (a *App) PrintStartMessage() {
	if a.merged {
		return
	}
	fmt.Println(a.listener.GetStartMessage(a.addr))
	a.GetMeta().DumpRoutes()
	for _, or := range a.mergedRoutes {
		or.GetMeta().DumpRoutes()
	}
}

func (a *App) ListenAndServe() error {
	if a.merged {
		return nil
	}

	a.PrintStartMessage()

	return a.listener.ListenAndServe(a.addr, a.Router)
}

func (a *App) MergeOtherApp(otherApp *App) {
	if a.merged || otherApp.merged {
		return
	}

	otherApp.merged = true
	a.mergedRoutes = append(a.mergedRoutes, otherApp.Router)
}

// StartAndWaitForShutdown starts the app and waits for interrupt/terminate signal, then gracefully shuts down.
func (a *App) StartAndWaitForShutdown(shutdownTimeout time.Duration) error {
	if a.merged {
		return nil
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.Start()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		// Received shutdown signal
		fmt.Println("Received signal:", sig)
		shutdownErr := a.Shutdown(shutdownTimeout)
		if shutdownErr != nil {
			return shutdownErr
		}
		return nil
	case err := <-errCh:
		// Server exited with error
		return err
	}
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

func (a *App) IsMerged() bool {
	return a.merged
}
