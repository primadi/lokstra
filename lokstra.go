package lokstra

import (
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

type Router = router.Router
type RequestContext = request.Context
type HandlerFunc = request.HandlerFunc
type Handler = request.Handler

// Create a new Router instance
func NewRouter(name string) Router { return router.New(name) }

// Create a new Router instance with specific engine type (e.g., "default", "servemux")
func NewRouterWithEngine(name string, engineType string) Router {
	return router.NewWithEngine(name, engineType)
}

// Create a new App instance with given routers
func NewApp(name string, addr string, routers ...Router) *app.App {
	app := app.New(name, addr)
	for _, r := range routers {
		app.AddRouter(r)
	}
	return app
}

// Create a new App instance with given routers and custom listener configuration
func NewAppWithConfig(name string, addr string, listenerType string,
	config map[string]any, routers ...Router) *app.App {
	app := app.NewWithConfig(name, addr, listenerType, config)
	for _, r := range routers {
		app.AddRouter(r)
	}
	return app
}
