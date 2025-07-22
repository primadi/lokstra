package module

import (
	"context"
	"errors"
	"fmt"
	"plugin"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

const EntryFnRegisterModule = "RegisterModule"

var handlers = make(map[string]*request.HandlerRegister)
var serviceFactories = make(map[string]service.ServiceFactory)
var serviceInstances = make(map[string]service.Service)
var modules = make(map[string]bool)

type RegistrationContextImpl struct {
	context.Context
	permission *PermissionGranted
}

var globalContext = &RegistrationContextImpl{
	permission: &PermissionGranted{
		whitelistGetService: []string{"*"},

		allowRegisterHandler:    true,
		allowRegisterMiddleware: true,
		allowRegisterService:    true,

		contextSettings: make(map[string]any),
	},
}

var globalContextCreated = false

func NewGlobalContext() *RegistrationContextImpl {
	if globalContextCreated {
		panic("GlobalContext has already been created")
	}
	globalContextCreated = true
	return globalContext
}

// RegisterModule implements ComponentContext.
func (g *RegistrationContextImpl) RegisterModule(moduleName string,
	registerFunc func(ctx RegistrationContext) error) error {
	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	if err := registerFunc(g); err != nil {
		return err
	}
	modules[moduleName] = true
	return nil
}

// RegisterPluginModule implements ComponentContext.
func (g *RegistrationContextImpl) RegisterPluginModuleWithEntry(moduleName string,
	pluginPath string, entryFn string) error {
	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("load plugin %s: %w", pluginPath, err)
	}
	sym, _ := p.Lookup(entryFn)
	entry, ok := sym.(func(ctx RegistrationContext) error)
	if !ok {
		return fmt.Errorf("plugin entry %s has wrong signature", entryFn)
	}
	if err := entry(g); err != nil {
		return fmt.Errorf("plugin %s failed: %w", moduleName, err)
	}

	modules[moduleName] = true
	return nil
}

// RegisterPluginModule implements ComponentContext.
func (g *RegistrationContextImpl) RegisterPluginModule(moduleName string,
	pluginPath string) error {
	return g.RegisterPluginModuleWithEntry(moduleName, pluginPath, EntryFnRegisterModule)
}

// GetHandler implements ComponentContext.
func (g *RegistrationContextImpl) GetHandler(name string) *HandlerRegister {
	return handlers[name]
}

// GetService implements ComponentContext.
func (g *RegistrationContextImpl) GetService(serviceUri string) service.Service {
	if !g.permission.IsAllowedGetService(serviceUri) {
		panic("service '" + serviceUri + "' is not allowed to be accessed")
	}

	if service, exists := serviceInstances[serviceUri]; exists {
		return service
	}
	return nil
}

// GetServiceFactory implements ComponentContext.
func (g *RegistrationContextImpl) GetServiceFactory(factoryName string) (service.ServiceFactory, bool) {
	sf, exists := serviceFactories[factoryName]
	return sf, exists
}

// CreateService implements ComponentContext.
func (g *RegistrationContextImpl) CreateService(factoryName string, serviceName string, config ...any) (service.Service, error) {
	factory, exists := serviceFactories[factoryName]
	if !exists {
		return nil, errors.New("service factory not found for type: " + factoryName)
	}

	var cfg any
	if len(config) == 0 {
		cfg = nil
	} else if len(config) == 1 {
		cfg = config[0]
	} else {
		cfg = config
	}

	service, err := factory(serviceName, cfg)
	if err != nil {
		return nil, err
	}

	serviceUri := service.GetServiceUri()
	if _, found := serviceInstances[serviceUri]; found {
		return nil, errors.New("service with name '" + serviceUri + "' already exists")
	}

	serviceInstances[serviceUri] = service

	return service, nil
}

// RegisterHandler implements ComponentContext.
func (g *RegistrationContextImpl) RegisterHandler(name string, handler request.HandlerFunc) {
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

// RegisterService implements ComponentContext.
func (g *RegistrationContextImpl) RegisterService(service service.Service) error {
	if !g.permission.IsAllowedRegisterService() {
		return errors.New("registering service '" + service.GetServiceUri() + "' is not allowed")
	}
	serviceInstances[service.GetServiceUri()] = service
	return nil
}

// RegisterServiceFactory implements ComponentContext.
func (g *RegistrationContextImpl) RegisterServiceFactory(factoryName string,
	serviceFactory func(serviceName string, config any) (service.Service, error)) {
	if !g.permission.IsAllowedRegisterService() {
		panic("registering service factory for '" + factoryName + "' is not allowed")
	}

	if serviceFactory == nil {
		panic("service factory cannot be nil")
	}
	if factoryName == "" {
		panic("factory name cannot be empty")
	}

	serviceFactories[factoryName] = serviceFactory
}

// RegisterServiceModule implements ComponentContext.
func (g *RegistrationContextImpl) RegisterServiceModule(module service.ServiceModule) error {
	if !g.permission.IsAllowedRegisterService() {
		return fmt.Errorf("registering service factory for '%s' is not allowed", module.FactoryName())
	}

	serviceFactories[module.FactoryName()] = module.Factory
	return nil
}

var _ RegistrationContext = (*RegistrationContextImpl)(nil)

func NewPermissionContext(permission *PermissionRequest) RegistrationContext {
	return &RegistrationContextImpl{
		permission: newPermissionGranted(permission),
	}
}

func (g *RegistrationContextImpl) NewPermissionContextFromConfig(settings map[string]any,
	permission map[string]any) RegistrationContext {

	pr := &PermissionRequest{
		WhitelistGetService:     utils.GetValueFromMap(permission, "get_service", []string{}),
		AllowRegisterHandler:    utils.GetValueFromMap(permission, "allow_register_handler", false),
		AllowRegisterMiddleware: utils.GetValueFromMap(permission, "allow_register_middleware", false),
		AllowRegisterService:    utils.GetValueFromMap(permission, "allow_register_service", false),
		ContextSettings:         settings,
	}

	return &RegistrationContextImpl{
		permission: newPermissionGranted(pr),
	}
}
