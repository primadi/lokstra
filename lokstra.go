package lokstra

import (
	"lokstra/common/component"
	"lokstra/common/iface"
	"lokstra/common/meta"
	"lokstra/core/app"
	"lokstra/core/request"
	"lokstra/core/server"
	"lokstra/serviceapi/core_service"
)

type Context = request.Context
type ComponentContext = component.ComponentContext

type HandlerFunc = iface.HandlerFunc
type App = app.App

func NewGlobalContext() *component.GlobalContext {
	return component.NewGlobalContext()
}

func NewServer(ctx component.ComponentContext, name string) *server.Server {
	return server.NewServer(ctx, name)
}

func NewApp(ctx component.ComponentContext, name string, port int) *app.App {
	return app.NewApp(ctx, name, port)
}

const LISTENER_NETHTTP = core_service.NETHTTP_LISTENER_NAME
const LISTENER_FASTHTTP = core_service.FASTHTTP_LISTENER_NAME
const LISTENER_SECURE_NETHTTP = core_service.SECURE_NETHTTP_LISTENER_NAME

const ROUTER_ENGINE_HTTPROUTER = core_service.HTTPROUTER_ROUTER_ENGINE_NAME
const ROUTER_ENGINE_SERVEMUX = core_service.SERVEMUX_ROUTER_ENGINE_NAME

func NewAppCustom(ctx component.ComponentContext, name string, port int,
	listenerType string, routerEngine string) *app.App {
	appMeta := meta.NewApp(name, port).WithListenerType(listenerType).WithRouterEngineType(routerEngine)
	return app.NewAppFromMeta(ctx, appMeta)
}
