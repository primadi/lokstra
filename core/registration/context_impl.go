package registration

import (
	"errors"
	"fmt"
	"plugin"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

const EntryFnRegisterModule = "GetModule"

type mwMeta struct {
	factory  midware.Factory
	priority int
}

var handlers = make(map[string]*request.HandlerRegister)
var serviceFactories = make(map[string]service.ServiceFactory)
var serviceInstances = make(map[string]service.Service)
var mwMetas = make(map[string]*mwMeta)

var modules = make(map[string]bool)

type ContextImpl struct {
	permission *PermissionGranted
}

// RegisterCompiledModule implements Context.
func (g *ContextImpl) RegisterCompiledModule(moduleName string,
	pluginPath string) error {
	return g.RegisterCompiledModuleWithFuncName(moduleName, pluginPath, EntryFnRegisterModule)
}

// RegisterCompiledModuleWithFuncName implements Context.
func (g *ContextImpl) RegisterCompiledModuleWithFuncName(moduleName string,
	pluginPath string, getModuleFuncName string) error {

	if pluginPath == "" {
		return fmt.Errorf("plugin path cannot be empty")
	}

	if getModuleFuncName == "" {
		getModuleFuncName = EntryFnRegisterModule
	}

	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("load plugin %s: %w", pluginPath, err)
	}
	sym, _ := p.Lookup(getModuleFuncName)
	getModuleFunc, ok := sym.(func(Context) error)
	if !ok {
		return fmt.Errorf("plugin entry %s has wrong signature", getModuleFuncName)
	}
	return g.RegisterModuleWithFunc(moduleName, getModuleFunc)
}

// RegisterModuleWithFunc implements Context.
func (g *ContextImpl) RegisterModuleWithFunc(moduleName string,
	getModuleFunc func(regCtx Context) error) error {
	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	if err := getModuleFunc(g); err != nil {
		return err
	}
	modules[moduleName] = true
	return nil
}

// GetHandler implements Context.
func (g *ContextImpl) GetHandler(name string) *HandlerRegister {
	return handlers[name]
}

// GetService implements Context.
func (g *ContextImpl) GetService(serviceUri string) service.Service {
	if !g.permission.IsAllowedGetService(serviceUri) {
		panic("service '" + serviceUri + "' is not allowed to be accessed")
	}

	if service, exists := serviceInstances[serviceUri]; exists {
		return service
	}
	return nil
}

// GetServiceFactory implements Context.
func (g *ContextImpl) GetServiceFactory(factoryName string) (service.ServiceFactory, bool) {
	sf, exists := serviceFactories[factoryName]
	return sf, exists
}

// CreateService implements Context.
func (g *ContextImpl) CreateService(factoryName string, serviceName string, config ...any) (service.Service, error) {
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

// RegisterHandler implements Context.
func (g *ContextImpl) RegisterHandler(name string, handler request.HandlerFunc) {
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

// RegisterService implements Context.
func (g *ContextImpl) RegisterService(service service.Service) error {
	if !g.permission.IsAllowedRegisterService() {
		return errors.New("registering service '" + service.GetServiceUri() + "' is not allowed")
	}
	serviceInstances[service.GetServiceUri()] = service
	return nil
}

// RegisterServiceFactory implements Context.
func (g *ContextImpl) RegisterServiceFactory(factoryName string,
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

// GetMiddlewareFactory implements Context.
func (g *ContextImpl) GetMiddlewareFactory(name string) (midware.Factory, int, bool) {
	if mt, exists := mwMetas[name]; exists {
		return mt.factory, mt.priority, true
	}
	return nil, 0, false
}

// RegisterMiddlewareFunc implements Context.
func (g *ContextImpl) RegisterMiddlewareFunc(name string,
	middlewareFunc midware.Func) error {
	return g.RegisterMiddlewareFactoryWithPriority(name,
		func(_ any) midware.Func {
			return middlewareFunc
		}, 50)
}

// RegisterMiddlewareFuncWithPriority implements Context.
func (g *ContextImpl) RegisterMiddlewareFuncWithPriority(name string,
	middlewareFunc midware.Func, priority int) error {
	return g.RegisterMiddlewareFactoryWithPriority(name,
		func(_ any) midware.Func {
			return middlewareFunc
		}, priority)
}

// RegisterMiddlewareFactory implements Context.
func (g *ContextImpl) RegisterMiddlewareFactory(name string,
	middlewareFactory midware.Factory) error {
	return g.RegisterMiddlewareFactoryWithPriority(name, middlewareFactory, 50)
}

// RegisterMiddlewareFactoryWithPriority implements Context.
func (g *ContextImpl) RegisterMiddlewareFactoryWithPriority(name string,
	middlewareFactory midware.Factory, priority int) error {

	if name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	if !g.permission.IsAllowedRegisterMiddleware() {
		return fmt.Errorf("registering middleware '%s' is not allowed", name)
	}

	if _, exists := mwMetas[name]; exists {
		return fmt.Errorf("middleware with name '%s' already exists", name)
	}

	mwMetas[name] = &mwMeta{
		factory:  middlewareFactory,
		priority: priority,
	}

	return nil
}

var _ Context = (*ContextImpl)(nil)

func NewPermissionContext(permission *PermissionRequest) Context {
	return &ContextImpl{
		permission: newPermissionGranted(permission),
	}
}

func (g *ContextImpl) NewPermissionContextFromConfig(settings map[string]any,
	permission map[string]any) Context {

	pr := &PermissionRequest{
		WhitelistGetService:     utils.GetValueFromMap(permission, "get_service", []string{}),
		AllowRegisterHandler:    utils.GetValueFromMap(permission, "allow_register_handler", false),
		AllowRegisterMiddleware: utils.GetValueFromMap(permission, "allow_register_middleware", false),
		AllowRegisterService:    utils.GetValueFromMap(permission, "allow_register_service", false),
		ContextSettings:         settings,
	}

	return &ContextImpl{
		permission: newPermissionGranted(pr),
	}
}
