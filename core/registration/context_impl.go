package registration

import (
	"errors"
	"fmt"
	"path"
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
var rawHandlers = make(map[string]*request.RawHandlerRegister)

var serviceFactories = make(map[string]service.ServiceFactory)
var serviceInstances = make(map[string]service.Service)
var mwMetas = make(map[string]*mwMeta)

var modules = make(map[string]bool)

type ContextImpl struct {
	permission                *PermissionGranted
	allowNewPermissionContext bool
}

// RegisterCompiledModule implements Context.
func (c *ContextImpl) RegisterCompiledModule(pluginPath string) error {
	return c.RegisterCompiledModuleWithFuncName(pluginPath, EntryFnRegisterModule)
}

// RegisterCompiledModuleWithFuncName implements Context.
func (c *ContextImpl) RegisterCompiledModuleWithFuncName(pluginPath string,
	getModuleFuncName string) error {
	if pluginPath == "" {
		return fmt.Errorf("plugin path cannot be empty")
	}

	if getModuleFuncName == "" {
		getModuleFuncName = EntryFnRegisterModule
	}

	p, err := plugin.Open(pluginPath)
	if err != nil {
		return fmt.Errorf("load plugin %s: %w", pluginPath, err)
	}
	sym, _ := p.Lookup(getModuleFuncName)
	getModuleFunc, ok := sym.(func() Module)
	if !ok {
		return fmt.Errorf("plugin entry %s has wrong signature", getModuleFuncName)
	}
	moduleName := getModuleFunc().Name()

	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	return c.RegisterModule(getModuleFunc)
}

// RegisterModule implements Context.
func (c *ContextImpl) RegisterModule(getModuleFunc func() Module) error {
	mdl := getModuleFunc()
	moduleName := mdl.Name()

	if _, exists := modules[moduleName]; exists {
		return errors.New("module with name '" + moduleName + "' already registered")
	}
	modules[moduleName] = true

	return mdl.Register(c)
}

// GetHandler implements Context.
func (c *ContextImpl) GetHandler(name string) *request.HandlerRegister {
	return handlers[name]
}

// GetRawHandler implements Context.
func (c *ContextImpl) GetRawHandler(name string) *request.RawHandlerRegister {
	return rawHandlers[name]
}

// GetService implements Context.
func (c *ContextImpl) GetService(serviceName string) (service.Service, error) {
	if !c.permission.IsAllowedGetService(serviceName) {
		return nil, ErrServiceIsNotAllowed(serviceName)
	}

	if service, exists := serviceInstances[serviceName]; exists {
		return service, nil
	}
	return nil, ErrServiceNotFound(serviceName)
}

// GetServiceFactory implements Context.
func (c *ContextImpl) GetServiceFactory(factoryName string) (service.ServiceFactory, bool) {
	sf, exists := serviceFactories[factoryName]
	return sf, exists
}

// GetServiceFactories implements Context.
func (c *ContextImpl) GetServiceFactories(pattern string) []service.ServiceFactory {
	result := make([]service.ServiceFactory, 0)
	for name, factory := range serviceFactories {
		if matched, _ := path.Match(pattern, name); matched {
			result = append(result, factory)
		}
	}
	return result
}

// CreateService implements Context.
func (c *ContextImpl) CreateService(factoryName string, serviceName string, allowReplace bool, config ...any) (service.Service, error) {
	if !c.permission.IsAllowedGetService(serviceName) {
		return nil, ErrServiceIsNotAllowed(serviceName)
	}

	factory, exists := c.GetServiceFactory(factoryName)
	if !exists {
		return nil, ErrServiceFactoryNotFound(factoryName)
	}

	if _, found := serviceInstances[serviceName]; found {
		if !allowReplace {
			return nil, ErrServiceAlreadyExists(serviceName)
		}
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

	serviceInstances[serviceName] = service

	return service, nil
}

// GetOrCreateService implements Context.
func (c *ContextImpl) GetOrCreateService(factoryName string, serviceName string, config ...any) (service.Service, error) {
	if svc, err := c.GetService(serviceName); err == nil {
		return svc, nil // Return existing service if found
	}
	return c.CreateService(factoryName, serviceName, true, config...)
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
				if err := ctx.BindAllSmart(paramPtr); err != nil {
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

// RegisterRawHandler implements Context.
func (c *ContextImpl) RegisterRawHandler(name string, handlerFunc request.RawHandlerFunc) {
	if !c.permission.IsAllowedRegisterHandler() {
		panic("registering handler '" + name + "' is not allowed")
	}

	if handlerFunc == nil {
		panic("handler cannot be nil")
	}
	if name == "" {
		panic("handler name cannot be empty")
	}

	info := &request.RawHandlerRegister{
		Name:        name,
		HandlerFunc: handlerFunc,
	}

	rawHandlers[name] = info
}

// RegisterService implements Context.
func (c *ContextImpl) RegisterService(serviceName string, service service.Service, allowReplace bool) error {
	if !c.permission.IsAllowedRegisterService() {
		return ErrServiceIsNotAllowed(serviceName)
	}

	if _, found := serviceInstances[serviceName]; found {
		if !allowReplace {
			return ErrServiceAlreadyExists(serviceName)
		}
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

// GetValue implements Context.
func (c *ContextImpl) GetValue(key string) (any, bool) {
	if value, exists := c.permission.contextSettings[key]; exists {
		return value, true
	}

	return nil, false
}

// SetValue implements Context.
func (c *ContextImpl) SetValue(key string, value any) {
	c.permission.contextSettings[key] = value
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
		ContextSettings:         utils.CloneMap(settings),
	}

	return &ContextImpl{
		permission: newPermissionGranted(pr),
	}
}

var _ Context = (*ContextImpl)(nil)
