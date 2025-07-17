package module

import (
	"errors"

	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/core/request"
)

type HandlerRegister = request.HandlerRegister

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceTypeInvalid = errors.New("service type is invalid")

type RegistrationContext interface {
	// Service creation and retrieval
	SetService(name string, service iface.Service) error
	GetService(name string) iface.Service
	CreateService(serviceType, name string, config ...any) (iface.Service, error)

	// Service factory registration and retrieval
	RegisterServiceFactory(serviceType string, serviceFactory func(config any) (iface.Service, error))
	RegisterServiceModule(module iface.ServiceModule) error
	GetServiceFactory(serviceType string) (iface.ServiceFactory, bool)

	// Handler registration and retrieval
	GetHandler(name string) *HandlerRegister
	RegisterHandler(name string, handler request.HandlerFunc)

	// Middleware registration and retrieval
	RegisterMiddlewareFactory(name string, middlewareFactory iface.MiddlewareFactory) error
	RegisterMiddlewareFunc(name string, middlewareFunc iface.MiddlewareFunc) error
	RegisterMiddlewareModule(module iface.MiddlewareModule) error
	GetMiddlewareModule(name string) (iface.MiddlewareModule, bool)

	// Module registration
	RegisterModule(moduleName string, registerFunc func(ctx RegistrationContext) error) error

	// Plugin module registration
	RegisterPluginModule(moduleName string, pluginPath string) error
	RegisterPluginModuleWithEntry(moduleName string, pluginPath string, entryFn string) error
}
