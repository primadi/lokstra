package lokstra

import (
	"errors"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/flow"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/server"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/defaults"
	"github.com/primadi/lokstra/modules/coreservice/listener"
	"github.com/primadi/lokstra/serviceapi"
)

type Context = request.Context
type RegistrationContext = registration.Context

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

	defaults.RegisterAll(ctx)

	// get default logger service
	Logger, _ = serviceapi.GetService[serviceapi.Logger](ctx, "logger")

	return ctx
}

// NewServer creates a new Server instance with the given context and name.
func NewServer(regCtx RegistrationContext, name string) *Server {
	return server.NewServer(regCtx, name)
}

// NewServerFromConfig start all modules, start all services, new all apps, and
// creates a new server instance from the provided configuration.
func NewServerFromConfig(regCtx RegistrationContext, cfg *config.LokstraConfig) (*Server, error) {
	svr, err := cfg.NewServerFromConfig(regCtx)
	if err != nil {
		return nil, err
	}
	loadSettingFromConfig(regCtx, svr, cfg)
	return svr, nil
}

// LoadConfigToServer loads the configuration into the provided server instance.
func LoadConfigToServer(regCtx RegistrationContext, cfg *config.LokstraConfig, svr *server.Server) (*Server, error) {
	svr, err := cfg.LoadConfigToServer(regCtx, svr)
	if err != nil {
		return nil, err
	}
	loadSettingFromConfig(regCtx, svr, cfg)
	return svr, nil
}

func loadSettingFromConfig(regCtx RegistrationContext, svr *server.Server, cfg *config.LokstraConfig) (*Server, error) {
	// change log_level if exists on server settings
	if l, exists := cfg.Server.Settings[serviceapi.ConfigKeyLogLevel]; exists {
		if LvlStr, ok := l.(string); ok {
			if logLvl, ok := serviceapi.ParseLogLevelSafe(LvlStr); ok {
				Logger.SetLogLevel(logLvl)
			}
		}
	}

	// change log_format if exists on server settings
	if l, exists := cfg.Server.Settings[serviceapi.ConfigKeyLogFormat]; exists {
		if formatStr, ok := l.(string); ok {
			Logger.SetFormat(formatStr)
		}
	}

	// change log_output if exists on server settings
	if l, exists := cfg.Server.Settings[serviceapi.ConfigKeyLogOutput]; exists {
		if output, ok := l.(string); ok {
			Logger.SetOutput(output)
		}
	}

	// change Logger if exists on server settings
	svc, err := serviceapi.GetService[serviceapi.Logger](regCtx, "logger")
	if err == nil {
		Logger = svc
	}

	// Initialize default flow services from global settings
	if dbPoolName, exists := cfg.Server.Settings["flow_dbPool"]; exists {
		if dbPoolStr, ok := dbPoolName.(string); ok {
			flow.SetDefaultDbPool(regCtx, dbPoolStr)
		}
	}

	// Initialize default flow services from global settings
	if loggerName, exists := cfg.Server.Settings["flow_logger"]; exists {
		if loggerStr, ok := loggerName.(string); ok {
			flow.SetDefaultLogger(regCtx, loggerStr)
		}
	}

	// Initialize default flow dbschema from global settings
	if dbSchemaName, exists := cfg.Server.Settings["flow_dbschema"]; exists {
		if dbSchemaStr, ok := dbSchemaName.(string); ok {
			flow.SetDefaultDbSchemaName(dbSchemaStr)
		}
	}

	return svr, nil
}

const CERT_FILE_KEY = listener.CERT_FILE_KEY
const KEY_FILE_KEY = listener.KEY_FILE_KEY
const CA_FILE_KEY = listener.CA_FILE_KEY

func NewApp(regCtx RegistrationContext, name string, addr string) *App {
	return app.NewApp(regCtx, name, addr)
}

func NewAppCustom(regCtx RegistrationContext, name string, addr string,
	listenerType string, routerEngine string, settings map[string]any) *App {
	return app.NewAppCustom(regCtx, name, addr, listenerType, routerEngine, settings)
}

func NewAppSecure(regCtx RegistrationContext, name string, addr string,
	certFile string, keyFile string, caFile string) *App {
	settings := map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
		CA_FILE_KEY:   caFile,
	}
	return app.NewAppCustom(regCtx, name, addr, defaults.HTTP_LISTENER_SECURE_NETHTTP, "", settings)
}

func NewAppHttp3(regCtx RegistrationContext, name string, addr string,
	certFile string, keyFile string, caFile string) *App {
	settings := map[string]any{
		CERT_FILE_KEY: certFile,
		KEY_FILE_KEY:  keyFile,
		CA_FILE_KEY:   caFile,
	}
	return app.NewAppCustom(regCtx, name, addr, defaults.HTTP_LISTENER_HTTP3, "", settings)
}

func NewAppFastHTTP(regCtx RegistrationContext, name string, addr string) *App {
	return app.NewAppCustom(regCtx, name, addr, defaults.HTTP_LISTENER_FASTHTTP, "", nil)
}

func NamedMiddleware(middlewareType string, config ...any) *midware.Execution {
	return midware.Named(middlewareType, config...)
}

// LoadConfigDir loads the configuration from the specified directory.
// It returns a pointer to the LokstraConfig and an error if any.
func LoadConfigDir(dir string) (*config.LokstraConfig, error) {
	return config.LoadConfigDir(dir)
}

// LoadConfigFile loads the configuration from the specified file.
// It returns a pointer to the LokstraConfig and an error if any.
func LoadConfigFile(filePath string) (*config.LokstraConfig, error) {
	return config.LoadConfigFile(filePath)
}

func GetService[T service.Service](regCtx RegistrationContext, serviceName string) (T, error) {
	svc, err := regCtx.GetService(serviceName)
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

func GetOrCreateService[T any](regCtx RegistrationContext,
	serviceName string, factoryName string, config ...any) (T, error) {
	svc, err := regCtx.GetService(serviceName)
	if err != nil {
		svc, err = regCtx.CreateService(factoryName, serviceName, config...)
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
