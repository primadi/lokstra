package lokstra

import (
	"errors"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/modules/coreservice"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/modules/coreservice/router_engine"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
)

type Context = request.Context
type RegistrationContext = registration.Context
type GlobalRegistrationContext = registration.ContextImpl

type HandlerFunc = request.HandlerFunc

type MiddlewareFunc = midware.Func
type MiddlewareFactory = midware.Factory

type Module = registration.Module

type Server = server.Server
type App = app.App

type LogFields = serviceapi.LogFields

type Service = service.Service
type ServiceFactory = service.ServiceFactory

var Logger, _ = logger.NewService("default", serviceapi.LogLevelInfo)

func NewGlobalRegistrationContext() *GlobalRegistrationContext {
	ctx := registration.NewGlobalContext()

	// automatically register logger and core services module
	_ = logger.GetModule().Register(ctx)
	_ = coreservice.GetModule().Register(ctx)

	return ctx
}

func NewServer(ctx *GlobalRegistrationContext, name string) *Server {
	return server.NewServer(ctx, name)
}

func NewServerFromConfig(ctx *GlobalRegistrationContext, cfg *config.LokstraConfig) (*Server, error) {
	return config.NewServerFromConfig(ctx, cfg)
}

const LISTENER_NETHTTP = listener.NETHTTP_LISTENER_NAME
const LISTENER_FASTHTTP = listener.FASTHTTP_LISTENER_NAME
const LISTENER_SECURE_NETHTTP = listener.SECURE_NETHTTP_LISTENER_NAME
const LISTENER_HTTP3 = listener.HTTP3_LISTENER_NAME

const ROUTER_ENGINE_HTTPROUTER = router_engine.HTTPROUTER_ROUTER_ENGINE_NAME
const ROUTER_ENGINE_SERVEMUX = router_engine.SERVEMUX_ROUTER_ENGINE_NAME

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

func NamedMiddleware(middlewareType string, config ...any) *midware.Execution {
	return midware.Named(middlewareType, config...)
}

// LoadConfigDir loads the configuration from the specified directory.
// It returns a pointer to the LokstraConfig and an error if any.
func LoadConfigDir(dir string) (*config.LokstraConfig, error) {
	return config.LoadConfigDir(dir)
}

func GetService[T service.Service](ctx RegistrationContext, serviceUri string) (T, error) {
	svc := ctx.GetService(serviceUri)
	if svc == nil {
		var zero T
		return zero, errors.New("service not found: " + serviceUri)
	}
	if typedSvc, ok := svc.(T); ok {
		return typedSvc, nil
	}
	var zero T
	return zero, errors.New("service type mismatch: " + serviceUri)
}
