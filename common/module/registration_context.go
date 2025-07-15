package module

import (
	"errors"
	"lokstra/common/iface"
	"lokstra/core/request"
)

type HandlerRegister = request.HandlerRegister

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceTypeInvalid = errors.New("service type is invalid")

type RegistrationContext interface {
	SetService(name string, service iface.Service) error
	GetService(name string) iface.Service
	CreateService(serviceType, name string, config ...any) (iface.Service, error)

	RegisterServiceFactory(serviceType string, serviceFactory func(config any) (iface.Service, error))
	RegisterServiceModule(module iface.ServiceModule) error
	GetServiceFactory(serviceType string) (iface.ServiceFactory, bool)

	GetHandler(name string) *HandlerRegister
	RegisterHandler(name string, handler request.HandlerFunc)

	RegisterMiddlewareFactory(name string, middlewareFactory iface.MiddlewareFactory) error
	RegisterMiddlewareFunc(name string, middlewareFunc iface.MiddlewareFunc) error
	RegisterMiddlewareModule(module iface.MiddlewareModule) error
	GetMiddlewareModule(name string) (iface.MiddlewareModule, bool)

	RegisterModule(moduleName string, registerFunc func(ctx RegistrationContext) error) error
	RegisterPluginModule(moduleName string, pluginPath string) error
	RegisterPluginModuleWithEntry(moduleName string, pluginPath string, entryFn string) error
}
