package registration

import (
	"errors"
	"fmt"
	"plugin"
	"reflect"

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
	permission                *PermissionGranted
	allowNewPermissionContext bool
}

// RegisterCompiledModule implements Context.
func (c *ContextImpl) RegisterCompiledModule(moduleName string,
	pluginPath string) error {
	return c.RegisterCompiledModuleWithFuncName(moduleName, pluginPath, EntryFnRegisterModule)
}

// RegisterCompiledModuleWithFuncName implements Context.
func (c *ContextImpl) RegisterCompiledModuleWithFuncName(moduleName string,
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
	return c.RegisterModuleWithFunc(moduleName, getModuleFunc)
}

// RegisterModuleWithFunc implements Context.
func (c *ContextImpl) RegisterModuleWithFunc(moduleName string,
	getModuleFunc func(regCtx Context) error) error {
	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	if err := getModuleFunc(c); err != nil {
		return err
	}
	modules[moduleName] = true
	return nil
}

// GetHandler implements Context.
func (c *ContextImpl) GetHandler(name string) *HandlerRegister {
	return handlers[name]
}

// GetService implements Context.
func (c *ContextImpl) GetService(serviceName string) (service.Service, error) {
	if !c.permission.IsAllowedGetService(serviceName) {
		return nil, errors.New("service '" + serviceName + "' is not allowed to be accessed")
	}

	if service, exists := serviceInstances[serviceName]; exists {
		return service, nil
	}
	return nil, errors.New("service '" + serviceName + "' not found")
}

// GetServiceFactory implements Context.
func (c *ContextImpl) GetServiceFactory(factoryName string) (service.ServiceFactory, bool) {
	sf, exists := serviceFactories[factoryName]
	return sf, exists
}

// CreateService implements Context.
func (c *ContextImpl) CreateService(factoryName string, serviceName string, config ...any) (service.Service, error) {
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

	service, err := factory(cfg)
	if err != nil {
		return nil, err
	}

	if _, found := serviceInstances[serviceName]; found {
		return nil, errors.New("service with name '" + serviceName + "' already exists")
	}

	serviceInstances[serviceName] = service

	return service, nil
}

// RegisterHandler implements Context.
func (c *ContextImpl) RegisterHandler(name string, handler any) {
	if !c.permission.IsAllowedRegisterHandler() {
		panic("registering handler '" + name + "' is not allowed")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}
	if name == "" {
		panic("handler name cannot be empty")
	}

	var handlerFunc request.HandlerFunc
	switch h := handler.(type) {
	case request.HandlerFunc:
		handlerFunc = h
	default:
		// Try to match func(ctx *request.Context, params *T) error
		fnVal := reflect.ValueOf(handler)
		fnType := fnVal.Type()

		if fnType.Kind() == reflect.Func &&
			fnType.NumIn() == 2 &&
			fnType.NumOut() == 1 &&
			fnType.In(0) == reflect.TypeOf((*request.Context)(nil)) &&
			fnType.Out(0) == reflect.TypeOf((*error)(nil)).Elem() &&
			fnType.In(1).Kind() == reflect.Ptr &&
			fnType.In(1).Elem().Kind() == reflect.Struct {

			paramType := fnType.In(1)

			handlerFunc = func(ctx *request.Context) error {
				paramPtr := reflect.New(paramType.Elem()).Interface()
				if err := ctx.BindAll(paramPtr); err != nil {
					return ctx.ErrorBadRequest(err.Error())
				}
				out := fnVal.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(paramPtr)})
				if !out[0].IsNil() {
					return out[0].Interface().(error)
				}
				return nil
			}
		} else {
			fmt.Printf("Handler type: %T\n", handler)
			panic("Invalid handler type, must be a HandlerFunc, or func(ctx, params)")
		}
	}

	info := &request.HandlerRegister{
		Name:        name,
		HandlerFunc: handlerFunc,
	}

	handlers[name] = info
}

// RegisterService implements Context.
func (c *ContextImpl) RegisterService(serviceName string, service service.Service) error {
	if !c.permission.IsAllowedRegisterService() {
		return errors.New("registering service '" + serviceName + "' is not allowed")
	}

	if _, found := serviceInstances[serviceName]; found {
		return errors.New("service with name '" + serviceName + "' already exists")

	}
	serviceInstances[serviceName] = service
	return nil
}

// RegisterServiceFactory implements Context.
func (c *ContextImpl) RegisterServiceFactory(factoryName string,
	serviceFactory func(config any) (service.Service, error)) {
	if !c.permission.IsAllowedRegisterService() {
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
func (c *ContextImpl) GetMiddlewareFactory(name string) (midware.Factory, int, bool) {
	if mt, exists := mwMetas[name]; exists {
		return mt.factory, mt.priority, true
	}
	return nil, 0, false
}

// RegisterMiddlewareFunc implements Context.
func (c *ContextImpl) RegisterMiddlewareFunc(name string,
	middlewareFunc midware.Func) error {
	return c.RegisterMiddlewareFactoryWithPriority(name,
		func(_ any) midware.Func {
			return middlewareFunc
		}, 50)
}

// RegisterMiddlewareFuncWithPriority implements Context.
func (c *ContextImpl) RegisterMiddlewareFuncWithPriority(name string,
	middlewareFunc midware.Func, priority int) error {
	return c.RegisterMiddlewareFactoryWithPriority(name,
		func(_ any) midware.Func {
			return middlewareFunc
		}, priority)
}

// RegisterMiddlewareFactory implements Context.
func (c *ContextImpl) RegisterMiddlewareFactory(name string,
	middlewareFactory midware.Factory) error {
	return c.RegisterMiddlewareFactoryWithPriority(name, middlewareFactory, 50)
}

// RegisterMiddlewareFactoryWithPriority implements Context.
func (c *ContextImpl) RegisterMiddlewareFactoryWithPriority(name string,
	middlewareFactory midware.Factory, priority int) error {

	if name == "" {
		return fmt.Errorf("middleware name cannot be empty")
	}

	if !c.permission.IsAllowedRegisterMiddleware() {
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

// NewPermissionContextFromConfig implements Context.
func (c *ContextImpl) NewPermissionContextFromConfig(settings map[string]any,
	permission map[string]any) Context {

	if !c.allowNewPermissionContext {
		panic("NewPermissionContextFromConfig is not allowed in this context")
	}

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

var _ Context = (*ContextImpl)(nil)
