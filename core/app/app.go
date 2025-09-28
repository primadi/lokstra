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
	routers        []router.Router
	listenerConfig map[string]any

	listener listener.AppListener
}

// Create a new App instance with default listener configuration
func New(name string, addr string) *App {
	return NewWithConfig(name, addr, "default", nil)
}

// Create a new App instance with custom listener configuration
func NewWithConfig(name string, addr string, listenerType string, cfg map[string]any) *App {
	if cfg == nil {
		cfg = make(map[string]any)
	}
	cfg["addr"] = addr
	cfg["listener-type"] = listenerType
	return &App{
		name:           name,
		routers:        make([]router.Router, 0),
		listenerConfig: cfg,
	}
}

func (a *App) AddRouter(r router.Router) {
	a.routers = append(a.routers, r)
}

func (a *App) PrintStartInfo() {
	fmt.Println("["+a.name+"] Starting app with", len(a.routers), "router(s) on address",
		a.listenerConfig["addr"].(string))
	for _, r := range a.routers {
		r.PrintRoutes()
	}
}

func (a *App) Start() error {
	// chain routers if more than 1
	for i := 0; i < len(a.routers)-1; i++ {
		a.routers[i].SetNextChain(a.routers[i+1])
	}
	a.listener = listener.CreateListener(a.listenerConfig, a.routers[0])
	return a.listener.ListenAndServe()
}
