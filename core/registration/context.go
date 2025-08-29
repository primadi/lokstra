package registration

import (
	"errors"
	"fmt"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

type HandlerRegister = request.HandlerRegister
type RawHandlerRegister = request.RawHandlerRegister

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceTypeInvalid = errors.New("service type is invalid")

// Module defines the interface for a module in Lokstra.
type Module interface {
	Name() string
	Description() string
	Register(regCtx Context) error
}

// registration.Context is used only during startup phase
// to register services, handlers, middleware, and modules.
// It must not be used after server.Start().
type Context interface {
	// Service creation and retrieval
	RegisterService(serviceName string, service service.Service) error
	GetService(serviceName string) (service.Service, error)
	CreateService(factoryName, serviceName string, config ...any) (service.Service, error)
	GetOrCreateService(factoryName, serviceName string, config ...any) (service.Service, error)

	// Service factory registration and retrieval
	RegisterServiceFactory(factoryName string,
		serviceFactory func(config any) (service.Service, error))
	GetServiceFactory(factoryName string) (service.ServiceFactory, bool)
	GetServiceFactories(pattern string) []service.ServiceFactory

	// Handler registration and retrieval
	GetHandler(name string) *HandlerRegister
	RegisterHandler(name string, handler any)

	GetRawHandler(name string) *RawHandlerRegister
	RegisterRawHandler(name string, handler request.RawHandlerFunc)

	// Middleware registration and retrieval

	// Register middleware factory by name, with default priority 50
	RegisterMiddlewareFactory(name string, middlewareFactory midware.Factory) error

	// priority scale is 1-100, where 1 is the highest priority
	RegisterMiddlewareFactoryWithPriority(name string, middlewareFactory midware.Factory, priority int) error

	// Register middleware function by name, with default priority 50
	RegisterMiddlewareFunc(name string, middlewareFunc midware.Func) error

	// priority scale is 1-100, where 1 is the highest priority
	RegisterMiddlewareFuncWithPriority(name string, middlewareFunc midware.Func, priority int) error

	// GetMiddlewareFactory retrieves a middleware factory by name
	// return the factory, priority, and whether it exists
	GetMiddlewareFactory(name string) (midware.Factory, int, bool)

	GetValue(key string) (any, bool)
	SetValue(key string, value any)

	// Module registration
	RegisterCompiledModule(pluginPath string) error // funcName is "GetModule"
	RegisterCompiledModuleWithFuncName(pluginPath string, getModuleFuncName string) error
	RegisterModule(getModuleFunc func() Module) error

	NewPermissionContextFromConfig(settings map[string]any, permission map[string]any) Context
}

func GetServiceFromConfig[T service.Service](regCtx Context,
	config any, paramServiceName string) (T, error) {
	var zero T
	svcName := ""
	switch cfg := config.(type) {
	case string:
		svcName = cfg
	case map[string]string:
		svcName = cfg[paramServiceName]
	default:
		return zero, service.ErrUnsupportedConfig(config)
	}

	if svcName == "" {
		return zero, fmt.Errorf(
			"failed to get service for %s: service name must be provided in config",
			paramServiceName)
	}
	svc, err := regCtx.GetService(svcName)
	if err != nil {
		return zero, fmt.Errorf("failed to get service for %s: %s", paramServiceName, err.Error())
	}
	if typedSvc, ok := svc.(T); ok {
		return typedSvc, nil
	}

	return zero, service.ErrInvalidServiceType(paramServiceName,
		fmt.Sprintf("%T", (*T)(nil)))
}

func GetValueFromConfig[T any](regCtx Context,
	config any, paramName string) (T, error) {
	var zero T
	switch cfg := config.(type) {
	case map[string]any:
		if val, ok := cfg[paramName]; ok {
			if typedVal, ok := val.(T); ok {
				return typedVal, nil
			}
			return zero, fmt.Errorf(
				"failed to get value for %s: expected type %T, got %T", paramName, zero, val)
		}
		return zero, fmt.Errorf("failed to get value for %s: key not found", paramName)
	default:
		if typedVal, ok := cfg.(T); ok {
			return typedVal, nil
		}
		return zero, fmt.Errorf("failed to get value for %s: unsupported config type %T",
			paramName, cfg)
	}
}
