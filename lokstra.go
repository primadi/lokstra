package lokstra

import (
	"lokstra/common/component"
	"lokstra/common/iface"
	"lokstra/common/meta"
	"lokstra/core/app"
	"lokstra/core/request"
	"lokstra/core/server"
	"lokstra/modules/coreservice_module/listener"
	"lokstra/serviceapi/core_service"
	"lokstra/serviceapi/logger_api"
	"lokstra/services/logger"
)

type Context = request.Context
type ComponentContext = component.ComponentContext
type GlobalContext = component.GlobalContext

type HandlerFunc = iface.HandlerFunc
type Server = server.Server
type App = app.App

var Logger = logger.NewService(logger_api.LogLevelInfo)

func NewGlobalContext() *GlobalContext {
	return component.NewGlobalContext()
}

func NewServer(ctx ComponentContext, name string) *Server {
	return server.NewServer(ctx, name)
}

func NewApp(ctx ComponentContext, name string, port int) *App {
	return app.NewApp(ctx, name, port)
}

const LISTENER_NETHTTP = core_service.NETHTTP_LISTENER_NAME
const LISTENER_FASTHTTP = core_service.FASTHTTP_LISTENER_NAME
const LISTENER_SECURE_NETHTTP = core_service.SECURE_NETHTTP_LISTENER_NAME

const ROUTER_ENGINE_HTTPROUTER = core_service.HTTPROUTER_ROUTER_ENGINE_NAME
const ROUTER_ENGINE_SERVEMUX = core_service.SERVEMUX_ROUTER_ENGINE_NAME

const CERT_FILE_KEY = listener.CERT_FILE
const KEY_FILE_KEY = listener.KEY_FILE

func NewAppCustom(ctx ComponentContext, name string, port int,
	listenerType string, routerEngine string) *App {
	appMeta := meta.NewApp(name, port).WithListenerType(listenerType).WithRouterEngineType(routerEngine)
	return app.NewAppFromMeta(ctx, appMeta)
}

func NewAppSecure(ctx ComponentContext, name string, port int,
	certFile string, keyFile string) *App {
	appMeta := meta.NewApp(name, port).
		WithListenerType(LISTENER_SECURE_NETHTTP)
	appMeta.SetSetting(CERT_FILE_KEY, certFile)
	appMeta.SetSetting(KEY_FILE_KEY, keyFile)
	return app.NewAppFromMeta(ctx, appMeta)
}

func NewAppFastHTTP(ctx ComponentContext, name string, port int) *App {
	appMeta := meta.NewApp(name, port).
		WithListenerType(LISTENER_FASTHTTP)
	return app.NewAppFromMeta(ctx, appMeta)
}

func NamedMiddleware(middlewareType string, config ...any) *meta.MiddlewareMeta {
	return &meta.MiddlewareMeta{
		MiddlewareType: middlewareType,
		Config:         config,
	}
}
