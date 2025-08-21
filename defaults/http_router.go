package defaults

import (
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/modules/coreservice/router_engine"
	"github.com/primadi/lokstra/serviceapi"
)

const (
	HTTP_ROUTER_HTTPROUTER = serviceapi.HTTP_ROUTER_PREFIX + "httprouter"
	HTTP_ROUTER_SERVEMUX   = serviceapi.HTTP_ROUTER_PREFIX + "servemux"
)

func RegisterAllHTTPRouters(regCtx iface.RegistrationContext) {
	regCtx.RegisterServiceFactory(HTTP_ROUTER_HTTPROUTER,
		router_engine.NewHttpRouterEngine)
	regCtx.RegisterServiceFactory(HTTP_ROUTER_SERVEMUX,
		router_engine.NewServeMuxEngine)

	regCtx.RegisterServiceFactory(serviceapi.HTTP_ROUTER_PREFIX+"default",
		router_engine.NewHttpRouterEngine)
}
