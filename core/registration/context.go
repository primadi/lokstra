package registration

import (
	"errors"

	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

type HandlerRegister = request.HandlerRegister

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceTypeInvalid = errors.New("service type is invalid")

type Context interface {
	// Service creation and retrieval
	RegisterService(service service.Service) error
	GetService(serviceUri string) service.Service
	CreateService(factoryName, serviceName string, config ...any) (service.Service, error)

	// Service factory registration and retrieval
	RegisterServiceFactory(factoryName string,
		serviceFactory func(serviceName string, config any) (service.Service, error))
	GetServiceFactory(factoryName string) (service.ServiceFactory, bool)

	// Handler registration and retrieval
	GetHandler(name string) *HandlerRegister
	RegisterHandler(name string, handler request.HandlerFunc)

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

	// Module registration
	RegisterCompiledModule(moduleName string, pluginPath string) error // funcName is "GetModule"
	RegisterCompiledModuleWithFuncName(moduleName string, pluginPath string, getModuleFuncName string) error
	RegisterModuleWithFunc(moduleName string, getModuleFunc func(ctx Context) error) error
}
