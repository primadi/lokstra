package component

import (
	"errors"
	"lokstra/common/iface"
	"lokstra/core/request"
	"strings"
)

type GlobalContext struct {
	handlers            map[string]*request.HandlerRegister
	middlewareFactories map[string]iface.MiddlewareFactory
	serviceFactories    map[string]iface.ServiceFactory
	serviceInstances    map[string]iface.Service
	modules             map[string]bool
}

var globalContext = &GlobalContext{
	handlers:            make(map[string]*request.HandlerRegister),
	middlewareFactories: make(map[string]iface.MiddlewareFactory),
	serviceFactories:    make(map[string]iface.ServiceFactory),
	serviceInstances:    make(map[string]iface.Service),
	modules:             make(map[string]bool),
}

var globalContextCreated = false

func NewGlobalContext() *GlobalContext {
	if globalContextCreated {
		panic("GlobalContext has already been created")
	}
	globalContextCreated = true
	return globalContext
}

// RegisterModuleFactory implements ComponentContext.
func (g *GlobalContext) RegisterModuleFactory(moduleName string, moduleFactory func(ComponentContext) error) error {
	if _, exists := g.modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	if err := moduleFactory(g); err != nil {
		return err
	}
	g.modules[moduleName] = true
	return nil
}

// RegisterPluginFactory implements ComponentContext.
func (g *GlobalContext) RegisterPluginFactory(pluginName string, pluginFactory func(ComponentContext) error) error {
	panic("Cannot register plugin factory in GlobalContext, use RegisterPluginFactory in PluginContext instead")
}

// GetHandler implements ComponentContext.
func (g *GlobalContext) GetHandler(name string) *HandlerRegister {
	if !strings.Contains(name, ".") {
		name = "main." + name
	}

	return g.handlers[name]
}

// GetMiddlewareFactory implements ComponentContext.
func (g *GlobalContext) GetMiddlewareFactory(middlewareType string) (iface.MiddlewareFactory, bool) {
	if !strings.Contains(middlewareType, ".") {
		middlewareType = "main." + middlewareType
	}

	middlewareFactory, exists := g.middlewareFactories[middlewareType]
	return middlewareFactory, exists
}

// GetService implements ComponentContext.
func (g *GlobalContext) GetService(name string) iface.Service {
	if !strings.Contains(name, ".") {
		name = "main." + name
	}
	if service, exists := g.serviceInstances[name]; exists {
		return service
	}
	return nil
}

// GetServiceFactory implements ComponentContext.
func (g *GlobalContext) GetServiceFactory(serviceType string) (iface.ServiceFactory, bool) {
	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	sf, exists := g.serviceFactories[serviceType]
	return sf, exists
}

// NewService implements ComponentContext.
func (g *GlobalContext) NewService(serviceType string, name string, config ...any) (iface.Service, error) {
	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	factory, exists := g.serviceFactories[serviceType]
	if !exists {
		return nil, errors.New("service factory not found for type: " + serviceType)
	}

	serviceName := serviceType + ":" + name
	if _, found := g.serviceInstances[serviceName]; found {
		return nil, errors.New("service with name '" + serviceName + "' already exists")
	}

	var cfg any
	if len(config) == 0 {
		cfg = nil
	} else if len(config) == 1 {
		cfg = config[0]
	} else {
		cfg = config
	}

	service, err := factory(cfg)
	if err != nil {
		return nil, err
	}

	g.serviceInstances[serviceName] = service

	return service, nil
}

// RegisterHandler implements ComponentContext.
func (g *GlobalContext) RegisterHandler(name string, handler request.HandlerFunc) {
	g.registerHandler(name, handler, false)
}

// RegisterHandlerWithReplace implements ComponentContext.
func (g *GlobalContext) RegisterHandlerWithReplace(name string, handler request.HandlerFunc) {
	g.registerHandler(name, handler, true)
}

func (g *GlobalContext) registerHandler(name string, handler request.HandlerFunc, allowReplace bool) {
	if handler == nil {
		panic("handler cannot be nil")
	}
	if name == "" {
		panic("handler name cannot be empty")
	}

	if !strings.Contains(name, ".") {
		name = "main." + name
	}

	if _, exists := g.handlers[name]; exists && !allowReplace {
		panic("handler with name '" + name + "' already exists")
	}

	info := &request.HandlerRegister{
		Name:        name,
		HandlerFunc: handler,
	}

	g.handlers[name] = info
}

// RegisterMiddlewareFactory implements ComponentContext.
func (g *GlobalContext) RegisterMiddlewareFactory(middlewareType string, middlewareFactory iface.MiddlewareFactory) {
	if middlewareFactory == nil {
		panic("middleware factory cannot be nil")
	}
	if middlewareType == "" {
		panic("middlewareType cannot be empty")
	}

	if !strings.Contains(middlewareType, ".") {
		middlewareType = "main." + middlewareType
	}

	if _, exists := g.middlewareFactories[middlewareType]; exists {
		panic("middleware with middlewareType '" + middlewareType + "' already exists")
	}

	g.middlewareFactories[middlewareType] = middlewareFactory
}

// RegisterMiddlewareFunc implements ComponentContext.
func (g *GlobalContext) RegisterMiddlewareFunc(middlewareType string, middlewareFunc iface.MiddlewareFunc) {
	g.RegisterMiddlewareFactory(middlewareType, func(_ any) iface.MiddlewareFunc {
		return middlewareFunc
	})
}

// RegisterService implements ComponentContext.
func (g *GlobalContext) RegisterService(name string, service iface.Service, allowOverride bool) error {
	if !strings.Contains(name, ".") {
		name = "main." + name
	}

	if _, exists := g.serviceInstances[name]; exists && !allowOverride {
		return errors.New("service with name '" + name + "' already exists")
	}

	g.serviceInstances[name] = service
	return nil
}

// RegisterServiceFactory implements ComponentContext.
func (g *GlobalContext) RegisterServiceFactory(serviceType string, serviceFactory func(config any) (iface.Service, error)) {
	if serviceFactory == nil {
		panic("service factory cannot be nil")
	}
	if serviceType == "" {
		panic("serviceType cannot be empty")
	}

	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	if _, exists := g.serviceFactories[serviceType]; exists {
		panic("service factory with serviceType '" + serviceType + "' already exists")
	}

	g.serviceFactories[serviceType] = serviceFactory
}

var _ ComponentContext = (*GlobalContext)(nil)
