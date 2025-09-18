package registration

import (
	"fmt"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

type HandlerRegister = request.HandlerRegister
type RawHandlerRegister = request.RawHandlerRegister

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

	// Registers a service with the given name into service registry.
	//
	// The service must be initialized before calling this function.
	//
	// Returns:
	//   - nil on success
	//   - ErrServiceAlreadyExists if a service with the same name already exists, and allowReplace is false
	//   - ErrServiceIsNotAllowed if registering the service is not allowed
	RegisterService(serviceName string, service service.Service, allowReplace bool) error

	// Retrieves a service by name from service registry.
	//
	// Returns error:
	//  - nil on success
	//  - ErrServiceNotAllowed if accessing the service is not allowed.
	//  - ErrServiceNotFound if the service does not exist.
	GetService(serviceName string) (service.Service, error)

	// Creates a service using the specified factory and configuration,
	// and insert into service registry.
	//
	// Returns error:
	//   - nil on success
	//   - ErrServiceAlreadyExists if a service with the same name already exists, and allowReplace is false
	//   - ErrServiceIsNotAllowed if registering the service is not allowed
	//   - ErrServiceFactoryNotFound if the specified factory does not exist
	CreateService(factoryName, serviceName string, allowReplace bool, config ...any) (service.Service, error)

	// Retrieves a service by name if it exists, otherwise creates it using the specified factory
	// and configuration, and insert into service registry.
	//
	// Returns error:
	//  - nil on success
	//  - ErrServiceNotAllowed if accessing the service is not allowed.
	//  - ErrServiceFactoryNotFound if the specified factory does not exist
	GetOrCreateService(factoryName, serviceName string, config ...any) (service.Service, error)

	// Gracefully shutdown all services that implement service.Shutdownable interface.
	// This function is called during server shutdown.
	// Returns error:
	//   - nil on success
	//   - all errors from services that failed to shutdown, aggregated into one error using errors.Join (Go 1.20+)
	ShutdownAllServices() error

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
		return zero, ErrUnsupportedConfig(config)
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

	return zero, ErrInvalidServiceType(paramServiceName,
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
