package app

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/core/app/listener"
	"github.com/primadi/lokstra/core/router"
)

type App struct {
	http.Handler

	name           string
	router         router.Router
	listenerConfig map[string]any

	listener listener.AppListener
}

// Create a new App instance with default listener configuration
func New(name string, addr string, routers ...router.Router) *App {
	return NewWithConfig(name, addr, "default", nil, routers...)
}

// Create a new App instance with custom listener configuration
func NewWithConfig(name string, addr string, listenerType string,
	cfg map[string]any, routers ...router.Router) *App {
	if cfg == nil {
		cfg = make(map[string]any)
	}
	cfg["addr"] = addr
	cfg["listener-type"] = listenerType

	var mainRouter router.Router
	for _, r := range routers {
		if mainRouter == nil {
			mainRouter = r
		} else {
			mainRouter.SetNextChain(r)
		}
	}

	return &App{
		name:           name,
		listenerConfig: cfg,
		router:         mainRouter,
	}
}

// Add a router to the app. If there's already a router, it will be chained.
func (a *App) AddRouter(r router.Router) {
	if a.router == nil {
		a.router = r
	} else {
		a.router.SetNextChain(r)
	}
}

func (a *App) numRouters() int {
	if a.router == nil {
		return 0
	}

	curRouter := a.router
	count := 0

	for curRouter != nil {
		count++
		curRouter = curRouter.GetNextChain()
	}
	return count
}

func (a *App) PrintStartInfo() {
	if a.router == nil {
		panic("No router added to the app. Use AddRouter() to add at least one router.")
	}

	fmt.Println("["+a.name+"] Starting app with", a.numRouters(), "router(s) on address",
		a.listenerConfig["addr"].(string))
	a.router.PrintRoutes()
	fmt.Println("Press CTRL+C to stop the server...")
}

func (a *App) Start() error {
	a.listener = listener.CreateListener(a.listenerConfig, a.router)
	return a.listener.ListenAndServe()
}
