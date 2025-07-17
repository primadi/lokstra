package lokstra

import (
	"github.com/primadi/lokstra/common/config"
	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/common/meta"
	"github.com/primadi/lokstra/common/module"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/modules/coreservice"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
)

type Context = request.Context
type RegistrationContext = module.RegistrationContext
type GlobalContext = module.RegistrationContextImpl

type HandlerFunc = iface.HandlerFunc

type MiddlewareFunc = iface.MiddlewareFunc
type MiddlewareFactory = iface.MiddlewareFactory
type MiddlewareMeta = iface.MiddlewareMeta
type MiddlewareModule = iface.MiddlewareModule

type Server = server.Server
type App = app.App

type LogFields = serviceapi.LogFields

type Service = iface.Service
type ServiceModule = iface.ServiceModule
type ServiceFactory = iface.ServiceFactory

var Logger = logger.NewService(serviceapi.LogLevelInfo)

func NewGlobalContext() *GlobalContext {
	ctx := module.NewGlobalContext()

	_ = ctx.RegisterServiceModule(logger.GetModule())
	_ = ctx.RegisterModule("coreservice_module", coreservice.RegisterModule)

	return ctx
}

func NewServer(ctx *GlobalContext, name string) *Server {
	return server.NewServer(ctx, name)
}

func NewServerFromConfig(ctx *GlobalContext, cfg *config.LokstraConfig) (*Server, error) {
	return config.NewServerFromConfig(ctx, cfg)
}

const LISTENER_NETHTTP = serviceapi.NETHTTP_LISTENER_NAME
const LISTENER_FASTHTTP = serviceapi.FASTHTTP_LISTENER_NAME
const LISTENER_SECURE_NETHTTP = serviceapi.SECURE_NETHTTP_LISTENER_NAME
const LISTENER_HTTP3 = serviceapi.HTTP3_LISTENER_NAME

const ROUTER_ENGINE_HTTPROUTER = serviceapi.HTTPROUTER_ROUTER_ENGINE_NAME
const ROUTER_ENGINE_SERVEMUX = serviceapi.SERVEMUX_ROUTER_ENGINE_NAME

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

func NamedMiddleware(middlewareType string, config ...any) *meta.MiddlewareExecution {
	return meta.NamedMiddleware(middlewareType, config...)
}

// LoadConfigDir loads the configuration from the specified directory.
// It returns a pointer to the LokstraConfig and an error if any.
func LoadConfigDir(dir string) (*config.LokstraConfig, error) {
	return config.LoadConfigDir(dir)
}
