package app

import (
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra/core/app/listener"
	"github.com/primadi/lokstra/core/router"
)

type App struct {
	http.Handler

	name           string
	mainRouter     router.Router
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
		mainRouter:     mainRouter,
	}
}

// Get the app name
func (a *App) GetName() string {
	return a.name
}

// Get the app listening address
func (a *App) GetAddress() string {
	if addr, ok := a.listenerConfig["addr"].(string); ok {
		return addr
	}
	return ""
}

// Get the main router of the app
func (a *App) GetRouter() router.Router {
	return a.mainRouter
}

// Add a router to the app. If there's already a router, it will be chained.
func (a *App) AddRouter(r router.Router) {
	if a.mainRouter == nil {
		a.mainRouter = r
	} else {
		a.mainRouter.SetNextChain(r)
	}
}

func (a *App) numRouters() int {
	if a.mainRouter == nil {
		return 0
	}

	curRouter := a.mainRouter
	count := 0

	for curRouter != nil {
		count++
		curRouter = curRouter.GetNextChain()
	}
	return count
}

// Print app start information, including the number of routers and their routes
func (a *App) PrintStartInfo() {
	if a.mainRouter == nil {
		panic("No router added to the app. Use AddRouter() to add at least one router.")
	}

	fmt.Println("Starting ["+a.name+"] app with", a.numRouters(), "router(s) on address",
		a.listenerConfig["addr"], "###")
	a.mainRouter.PrintRoutes()
}

// Start runs the HTTP server. It blocks until the app stops or returns an error.
// Shutdown must be called separately.
func (a *App) Start() error {
	a.listener = listener.CreateListener(a.listenerConfig, a.mainRouter)
	return a.listener.ListenAndServe()
}

// Shutdown gracefully shuts down the app with a timeout.
func (a *App) Shutdown(timeout time.Duration) error {
	if a.listener != nil {
		return a.listener.Shutdown(timeout)
	}
	return nil
}
