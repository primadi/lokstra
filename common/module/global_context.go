package module

import (
	"context"
	"errors"
	"fmt"
	"plugin"

	"github.com/primadi/lokstra/common/iface"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
)

const EntryFnRegisterModule = "RegisterModule"

var handlers = make(map[string]*request.HandlerRegister)
var serviceFactories = make(map[string]iface.ServiceFactory)
var serviceInstances = make(map[string]iface.Service)
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
func (g *RegistrationContextImpl) GetService(name string) iface.Service {
	if !g.permission.IsAllowedGetService(name) {
		panic("service '" + name + "' is not allowed to be accessed")
	}

	if service, exists := serviceInstances[name]; exists {
		return service
	}
	return nil
}

// GetServiceFactory implements ComponentContext.
func (g *RegistrationContextImpl) GetServiceFactory(serviceType string) (iface.ServiceFactory, bool) {
	sf, exists := serviceFactories[serviceType]
	return sf, exists
}

// CreateService implements ComponentContext.
func (g *RegistrationContextImpl) CreateService(serviceType string, name string, config ...any) (iface.Service, error) {
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

// SetService implements ComponentContext.
func (g *RegistrationContextImpl) SetService(name string, service iface.Service) error {
	if !g.permission.IsAllowedRegisterService() {
		return errors.New("registering service '" + name + "' is not allowed")
	}

	serviceInstances[name] = service
	return nil
}

// RegisterServiceFactory implements ComponentContext.
func (g *RegistrationContextImpl) RegisterServiceFactory(serviceType string,
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

	serviceFactories[serviceType] = serviceFactory
}

// RegisterServiceModule implements ComponentContext.
func (g *RegistrationContextImpl) RegisterServiceModule(module iface.ServiceModule) error {
	if !g.permission.IsAllowedRegisterService() {
		return fmt.Errorf("registering service factory for '%s' is not allowed", module.Name())
	}

	serviceFactories[module.Name()] = module.Factory
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
