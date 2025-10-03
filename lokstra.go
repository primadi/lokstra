package lokstra

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/common/cast"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/lokstra_registry"
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

func FetchAndCast[T any](c *request.Context, client *lokstra_registry.ClientRouter, path string) (*T, error) {
	resp, err := client.GET(path)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to fetch: %v", err))
	}
	formatter := api_formatter.GetGlobalFormatter()
	clientResp := &api_formatter.ClientResponse{}
	if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to parse response: %v", err))
	}
	if clientResp.Status != "success" {
		return nil, c.Api.BadRequest("NOT_FOUND", "Resource not found")
	}
	var result T
	if err := cast.ToStruct(clientResp.Data, &result, true); err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to cast data: %v", err))
	}
	return &result, nil
}
