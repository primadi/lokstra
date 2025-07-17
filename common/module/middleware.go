package module

import (
	"fmt"

	"github.com/primadi/lokstra/common/iface"
)

var middlewareFactories = make(map[string]iface.MiddlewareModule)

// GetMiddlewareModule implements ComponentContext.
func (g *RegistrationContextImpl) GetMiddlewareModule(name string) (iface.MiddlewareModule, bool) {
	module, exists := middlewareFactories[name]
	return module, exists
}

// RegisterMiddlewareFactory implements ComponentContext.
func (g *RegistrationContextImpl) RegisterMiddlewareFactory(name string,
	middlewareFactory iface.MiddlewareFactory) error {
	return g.RegisterMiddlewareModule(NewMiddlewareModule(name, middlewareFactory, nil))
}

// RegisterMiddlewareFunc implements ComponentContext.
func (g *RegistrationContextImpl) RegisterMiddlewareFunc(name string,
	middlewareFunc iface.MiddlewareFunc) error {
	return g.RegisterMiddlewareModule(NewMiddlewareModule(name,
		func(_ any) iface.MiddlewareFunc {
			return middlewareFunc
		}, nil))
}

// RegisterMiddlewareModule implements ComponentContext.
func (g *RegistrationContextImpl) RegisterMiddlewareModule(module iface.MiddlewareModule) error {
	if module == nil {
		return fmt.Errorf("middleware module cannot be nil")
	}

	if module.Name() == "" {
		return fmt.Errorf("middlewareType cannot be empty")
	}

	if !g.permission.IsAllowedRegisterMiddleware() {
		return fmt.Errorf("registering middleware '%s' is not allowed", module.Name())
	}

	if _, exists := middlewareFactories[module.Name()]; exists {
		return fmt.Errorf("middleware with name '%s' already exists", module.Name())
	}

	middlewareFactories[module.Name()] = module
	return nil
}
