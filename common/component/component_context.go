package component

import (
	"errors"
	"lokstra/common/iface"
	"lokstra/core/request"
)

type HandlerRegister = request.HandlerRegister

var ErrServiceNotFound = errors.New("service not found")
var ErrServiceTypeInvalid = errors.New("service type is invalid")

type ComponentContext interface {
	RegisterService(name string, service iface.Service, allowOverride bool) error
	GetService(name string) iface.Service
	NewService(serviceType, name string, config ...any) (iface.Service, error)

	RegisterServiceFactory(serviceType string, serviceFactory func(config any) (iface.Service, error))
	GetServiceFactory(serviceType string) (iface.ServiceFactory, bool)

	GetHandler(name string) *HandlerRegister
	RegisterHandler(name string, handler request.HandlerFunc)
	RegisterHandlerWithReplace(name string, handler request.HandlerFunc)

	RegisterMiddlewareFactory(middlewareType string, middlewareFactory iface.MiddlewareFactory)
	RegisterMiddlewareFunc(middlewareType string, middlewareFunc iface.MiddlewareFunc)
	GetMiddlewareFactory(middlewareType string) (iface.MiddlewareFactory, bool)

	RegisterModule(moduleName string, fnReg func(ComponentContext) error) error
	RegisterPlugin(pluginName string, fnReg func(ComponentContext) error) error
}
