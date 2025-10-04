package lokstra

import (
	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

type Server = server.Server
type App = app.App
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
	return app.New(name, addr, routers...)
}

// Create a new App instance with given routers and custom listener configuration
func NewAppWithConfig(name string, addr string, listenerType string,
	config map[string]any, routers ...Router) *app.App {
	return app.NewWithConfig(name, addr, listenerType, config, routers...)
}

func NewServer(name string, apps ...*app.App) *server.Server {
	return server.New(name, apps...)
}

func FetchAndCast[T any](c *request.Context, client *api_client.ClientRouter, path string,
	opts ...api_client.FetchOption) (*T, error) {
	return api_client.FetchAndCast[T](c, client, path, opts...)
}
