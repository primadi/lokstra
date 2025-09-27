package lokstra

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

type Router = router.Router
type RequestContext = request.Context
type HandlerFunc = request.HandlerFunc
type Handler = request.Handler

func NewRouter(name string) Router { return router.New(name) }
func NewRouterWithEngine(name string, engineType string) Router {
	return router.NewWithEngine(name, engineType)
}
