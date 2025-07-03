package lokstra

import (
	"lokstra/common/iface"
	"lokstra/common/meta"
	"lokstra/common/registry"
	"lokstra/core/request"
	"lokstra/core/router/listener"
)

type ServerMeta = meta.ServerMeta
type AppMeta = meta.AppMeta
type RouterMeta = meta.RouterMeta
type HandlerMeta = meta.HandlerMeta
type MiddlewareMeta = meta.MiddlewareMeta

type Context = request.Context

type MiddlewareFunc = iface.MiddlewareFunc
type HandlerFunc = request.HandlerFunc

type HttpListener = listener.HttpListener

func NewServer(serverName string) *ServerMeta {
	return meta.NewServer(serverName)
}

func NewRouter() *meta.RouterMeta {
	return meta.NewRouterInfo()
}

func NewApp(name string, port int) *meta.AppMeta {
	return meta.NewApp(name, port)
}

func RegisterHandler(name string, handler func(ctx *Context) error) {
	registry.RegisterHandler(name, handler)
}

func RegisterMiddlewareFactory(name string, middlewareFactory func(config any) MiddlewareFunc) {
	registry.RegisterMiddlewareFactory(name, middlewareFactory)
}

func RegisterMiddlewareFunc(name string,
	middlewareFunc func(next HandlerFunc) HandlerFunc) {
	registry.RegisterMiddlewareFunc(name, middlewareFunc)
}

func RegisterServiceFactory(name string, serviceFactory func(config any) (iface.Service, error)) {
	registry.RegisterServiceFactory(name, serviceFactory)
}

// NamedMiddleware creates a middleware info with the given name and config.
// This is useful for registering middlewares with the registry.
func NamedMiddleware(name string, config ...any) *meta.MiddlewareMeta {
	return meta.NamedMiddleware(name, config...)
}

// NamedService creates a service info with the given name and instance.
func NewService[T iface.Service](serviceType, name string, config ...any) (T, error) {
	svc, err := registry.NewService(serviceType, name, config...)
	if err != nil {
		var zero T
		return zero, err
	}
	if service, ok := svc.(T); ok {
		return service, nil
	}
	var zero T
	return zero, iface.ErrServiceTypeMismatch
}

func GetService[T iface.Service](name string) T {
	if s := registry.GetService(name); s != nil {
		if service, ok := s.(T); ok {
			return service
		}
	}

	var zero T
	return zero
}

func Version() string {
	return "1.0.0"
}
