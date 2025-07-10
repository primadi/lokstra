package lokstra

import (
	"lokstra/common/config"
	"lokstra/common/iface"
	"lokstra/common/meta"
	"lokstra/common/module"
	"lokstra/core/app"
	"lokstra/core/request"
	"lokstra/core/server"
	"lokstra/modules/coreservice_module/listener"
	"lokstra/serviceapi/core_service"
	"lokstra/serviceapi/logger_api"
	"lokstra/services/logger"
)

type Context = request.Context
type RegistrationContext = module.RegistrationContext
type GlobalContext = module.RegistrationContextImpl

type HandlerFunc = iface.HandlerFunc
type Server = server.Server
type App = app.App

var Logger = logger.NewService(logger_api.LogLevelInfo)

func NewGlobalContext() *GlobalContext {
	return module.NewGlobalContext()
}

func NewServer(ctx *GlobalContext, name string) *Server {
	return server.NewServer(ctx, name)
}

func NewServerFromConfig(ctx *GlobalContext, cfg *config.LokstraConfig) (*Server, error) {
	return config.NewServerFromConfig(ctx, cfg)
}

const LISTENER_NETHTTP = core_service.NETHTTP_LISTENER_NAME
const LISTENER_FASTHTTP = core_service.FASTHTTP_LISTENER_NAME
const LISTENER_SECURE_NETHTTP = core_service.SECURE_NETHTTP_LISTENER_NAME
const LISTENER_HTTP3 = core_service.HTTP3_LISTENER_NAME

const ROUTER_ENGINE_HTTPROUTER = core_service.HTTPROUTER_ROUTER_ENGINE_NAME
const ROUTER_ENGINE_SERVEMUX = core_service.SERVEMUX_ROUTER_ENGINE_NAME

const CERT_FILE_KEY = listener.CERT_FILE_KEY
const KEY_FILE_KEY = listener.KEY_FILE_KEY
const CA_FILE_KEY = listener.CA_FILE_KEY

func NewApp(ctx RegistrationContext, name string, addr string) *App {
	return app.NewApp(ctx, name, addr)
}

func NewAppCustom(ctx RegistrationContext, name string, addr string,
	listenerType string, routerEngine string, settings map[string]any) *App {
	return app.NewAppCustom(ctx, name, addr, listenerType, routerEngine, settings)
}

func NewAppSecure(ctx RegistrationContext, name string, addr string,
	certFile string, keyFile string, caFile string) *App {
	settings := map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
		CA_FILE_KEY:   caFile,
	}
	return app.NewAppCustom(ctx, name, addr, LISTENER_SECURE_NETHTTP, ROUTER_ENGINE_HTTPROUTER, settings)
}

func NewAppHttp3(ctx RegistrationContext, name string, addr string,
	certFile string, keyFile string, caFile string) *App {
	settings := map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
		CA_FILE_KEY:   caFile,
	}
	return app.NewAppCustom(ctx, name, addr, LISTENER_HTTP3, ROUTER_ENGINE_HTTPROUTER, settings)
}

func NewAppFastHTTP(ctx RegistrationContext, name string, addr string) *App {
	return app.NewAppCustom(ctx, name, addr, LISTENER_FASTHTTP, ROUTER_ENGINE_HTTPROUTER, nil)
}

func NamedMiddleware(middlewareType string, config ...any) *meta.MiddlewareMeta {
	return &meta.MiddlewareMeta{
		MiddlewareType: middlewareType,
		Config:         config,
	}
}

// LoadConfigDir loads the configuration from the specified directory.
// It returns a pointer to the LokstraConfig and an error if any.
func LoadConfigDir(dir string) (*config.LokstraConfig, error) {
	return config.LoadConfigDir(dir)
}
