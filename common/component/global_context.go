package component

import (
	"errors"
	"lokstra/common/iface"
	"lokstra/core/request"
	"strings"
)

var handlers = make(map[string]*request.HandlerRegister)
var middlewareFactories = make(map[string]iface.MiddlewareFactory)
var serviceFactories = make(map[string]iface.ServiceFactory)
var serviceInstances = make(map[string]iface.Service)
var modules = make(map[string]bool)

type ComponentContextImpl struct {
	permission *PermissionGranted
}

var globalContext = &ComponentContextImpl{
	permission: &PermissionGranted{
		whitelistGetService: []string{"*"},

		allowRegisterHandler:    true,
		allowRegisterMiddleware: true,
		allowRegisterService:    true,

		contextSettings: make(map[string]any),
	},
}

var globalContextCreated = false

func NewGlobalContext() *ComponentContextImpl {
	if globalContextCreated {
		panic("GlobalContext has already been created")
	}
	globalContextCreated = true
	return globalContext
}

// RegisterModule implements ComponentContext.
func (g *ComponentContextImpl) RegisterModule(moduleName string,
	registerFunc func(ctx ComponentContext) error) error {
	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	if err := registerFunc(g); err != nil {
		return err
	}
	modules[moduleName] = true
	return nil
}

// GetHandler implements ComponentContext.
func (g *ComponentContextImpl) GetHandler(name string) *HandlerRegister {
	return handlers[name]
}

// GetMiddlewareFactory implements ComponentContext.
func (g *ComponentContextImpl) GetMiddlewareFactory(middlewareType string) (iface.MiddlewareFactory, bool) {
	middlewareFactory, exists := middlewareFactories[middlewareType]
	return middlewareFactory, exists
}

// GetService implements ComponentContext.
func (g *ComponentContextImpl) GetService(name string) iface.Service {
	if !g.permission.IsAllowedGetService(name) {
		panic("service '" + name + "' is not allowed to be accessed")
	}

	if service, exists := serviceInstances[name]; exists {
		return service
	}
	return nil
}

// GetServiceFactory implements ComponentContext.
func (g *ComponentContextImpl) GetServiceFactory(serviceType string) (iface.ServiceFactory, bool) {
	sf, exists := serviceFactories[serviceType]
	return sf, exists
}

// NewService implements ComponentContext.
func (g *ComponentContextImpl) NewService(serviceType string, name string, config ...any) (iface.Service, error) {
	factory, exists := serviceFactories[serviceType]
	if !exists {
		return nil, errors.New("service factory not found for type: " + serviceType)
	}

	if _, found := serviceInstances[name]; found {
		return nil, errors.New("service with name '" + name + "' already exists")
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

	serviceInstances[name] = service

	return service, nil
}

// RegisterHandler implements ComponentContext.
func (g *ComponentContextImpl) RegisterHandler(name string, handler request.HandlerFunc) {
	if !g.permission.IsAllowedRegisterHandler() {
		panic("registering handler '" + name + "' is not allowed")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}
	if name == "" {
		panic("handler name cannot be empty")
	}

	info := &request.HandlerRegister{
		Name:        name,
		HandlerFunc: handler,
	}

	handlers[name] = info
}

// RegisterMiddlewareFactory implements ComponentContext.
func (g *ComponentContextImpl) RegisterMiddlewareFactory(middlewareType string,
	middlewareFactory iface.MiddlewareFactory) {
	if !g.permission.IsAllowedRegisterMiddleware() {
		panic("registering middleware '" + middlewareType + "' is not allowed")
	}

	if middlewareFactory == nil {
		panic("middleware factory cannot be nil")
	}
	if middlewareType == "" {
		panic("middlewareType cannot be empty")
	}

	if _, exists := middlewareFactories[middlewareType]; exists {
		panic("middleware with middlewareType '" + middlewareType + "' already exists")
	}

	middlewareFactories[middlewareType] = middlewareFactory
}

// RegisterMiddlewareFunc implements ComponentContext.
func (g *ComponentContextImpl) RegisterMiddlewareFunc(middlewareType string, middlewareFunc iface.MiddlewareFunc) {
	g.RegisterMiddlewareFactory(middlewareType, func(_ any) iface.MiddlewareFunc {
		return middlewareFunc
	})
}

// RegisterService implements ComponentContext.
func (g *ComponentContextImpl) RegisterService(name string, service iface.Service) error {
	if !g.permission.IsAllowedRegisterService() {
		return errors.New("registering service '" + name + "' is not allowed")
	}

	serviceInstances[name] = service
	return nil
}

// RegisterServiceFactory implements ComponentContext.
func (g *ComponentContextImpl) RegisterServiceFactory(serviceType string,
	serviceFactory func(config any) (iface.Service, error)) {
	if !g.permission.IsAllowedRegisterService() {
		panic("registering service factory for '" + serviceType + "' is not allowed")
	}

	if serviceFactory == nil {
		panic("service factory cannot be nil")
	}
	if serviceType == "" {
		panic("serviceType cannot be empty")
	}

	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	serviceFactories[serviceType] = serviceFactory
}

var _ ComponentContext = (*ComponentContextImpl)(nil)

func (g *ComponentContextImpl) CreatePermissionContext(permission *PermissionRequest) ComponentContext {
	return &ComponentContextImpl{
		permission: newPermissionGranted(permission),
	}
}
