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
type GlobalContext = component.ComponentContextImpl

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

const LISTENER_NETHTTP = core_service.NETHTTP_LISTENER_NAME
const LISTENER_FASTHTTP = core_service.FASTHTTP_LISTENER_NAME
const LISTENER_SECURE_NETHTTP = core_service.SECURE_NETHTTP_LISTENER_NAME

const ROUTER_ENGINE_HTTPROUTER = core_service.HTTPROUTER_ROUTER_ENGINE_NAME
const ROUTER_ENGINE_SERVEMUX = core_service.SERVEMUX_ROUTER_ENGINE_NAME

const CERT_FILE_KEY = listener.CERT_FILE
const KEY_FILE_KEY = listener.KEY_FILE

func NewApp(ctx ComponentContext, name string, addr string) *App {
	return app.NewApp(ctx, name, addr)
}

func NewAppCustom(ctx ComponentContext, name string, addr string,
	listenerType string, routerEngine string, settings map[string]any) *App {
	return app.NewAppCustom(ctx, name, addr, listenerType, routerEngine, settings)
}

func NewAppSecure(ctx ComponentContext, name string, addr string,
	certFile string, keyFile string) *App {
	settings := map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
	}
	return app.NewAppCustom(ctx, name, addr, LISTENER_SECURE_NETHTTP, ROUTER_ENGINE_HTTPROUTER, settings)
}

func NewAppFastHTTP(ctx ComponentContext, name string, addr string) *App {
	return app.NewAppCustom(ctx, name, addr, LISTENER_FASTHTTP, ROUTER_ENGINE_HTTPROUTER, nil)
}

func NamedMiddleware(middlewareType string, config ...any) *meta.MiddlewareMeta {
	return &meta.MiddlewareMeta{
		MiddlewareType: middlewareType,
		Config:         config,
	}
}
