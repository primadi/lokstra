package lokstra

import (
	"errors"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/modules/coreservice"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/modules/rpc_service"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/logger"
	"github.com/primadi/lokstra/standardservices"
)

type Context = request.Context
type RegistrationContext = iface.RegistrationContext

type HandlerFunc = request.HandlerFunc

type MiddlewareFunc = midware.Func
type MiddlewareFactory = midware.Factory

type Module = registration.Module

type Server = server.Server
type App = app.App

type LogFields = serviceapi.LogFields

type Service = service.Service
type ServiceFactory = service.ServiceFactory

var Logger serviceapi.Logger

func NewGlobalRegistrationContext() RegistrationContext {
	ctx := registration.NewGlobalContext()

	standardservices.RegisterAll(ctx)

	// register logger module
	_ = logger.GetModule().Register(ctx)

	// create default logger service
	l, _ := ctx.CreateService(logger.FACTORY_NAME, "logger.default", "info")
	Logger = l.(serviceapi.Logger)

	// register core service module
	_ = coreservice.GetModule().Register(ctx)

	// register rpc service module
	_ = rpc_service.GetModule().Register(ctx)

	return ctx
}

func NewServer(regCtx RegistrationContext, name string) *Server {
	return server.NewServer(regCtx, name)
}

func NewServerFromConfig(regCtx RegistrationContext, cfg *config.LokstraConfig) (*Server, error) {
	svr, err := config.NewServerFromConfig(regCtx, cfg)
	if err != nil {
		return nil, err
	}

	// change log_level is exists on server settings
	if l, exists := cfg.Server.Settings[serviceapi.ConfigKeyLogLevel]; exists {
		if LvlStr, ok := l.(string); ok {
			if logLvl, ok := serviceapi.ParseLogLevelSafe(LvlStr); ok {
				Logger.SetLogLevel(logLvl)
			}
		}
	}

	return svr, nil
}

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
	return app.NewAppCustom(ctx, name, addr, standardservices.HTTP_LISTENER_SECURE_NETHTTP, "", settings)
}

func NewAppHttp3(ctx RegistrationContext, name string, addr string,
	certFile string, keyFile string, caFile string) *App {
	settings := map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
		CA_FILE_KEY:   caFile,
	}
	return app.NewAppCustom(ctx, name, addr, standardservices.HTTP_LISTENER_HTTP3, "", settings)
}

func NewAppFastHTTP(ctx RegistrationContext, name string, addr string) *App {
	return app.NewAppCustom(ctx, name, addr, standardservices.HTTP_LISTENER_FASTHTTP, "", nil)
}

func NamedMiddleware(middlewareType string, config ...any) *midware.Execution {
	return midware.Named(middlewareType, config...)
}

// LoadConfigDir loads the configuration from the specified directory.
// It returns a pointer to the LokstraConfig and an error if any.
func LoadConfigDir(dir string) (*config.LokstraConfig, error) {
	return config.LoadConfigDir(dir)
}

func GetService[T service.Service](ctx RegistrationContext, serviceName string) (T, error) {
	svc, err := ctx.GetService(serviceName)
	if err != nil {
		var zero T
		return zero, errors.New("service not found: " + serviceName)
	}
	if typedSvc, ok := svc.(T); ok {
		return typedSvc, nil
	}
	var zero T
	return zero, errors.New("service type mismatch: " + serviceName)
}

func GetOrCreateService[T any](ctx RegistrationContext,
	serviceName string, factoryName string, config ...any) (T, error) {
	svc, err := ctx.GetService(serviceName)
	if err != nil {
		svc, err = ctx.CreateService(factoryName, serviceName, config...)
		if err != nil {
			var zero T
			return zero, errors.New("failed to create service: " + err.Error())
		}
	}
	if typedSvc, ok := svc.(T); ok {
		return typedSvc, nil
	}
	var zero T
	return zero, errors.New("service type mismatch: " + serviceName)
}
