package lokstra_registry

import (
	"github.com/primadi/lokstra/core/router"
)

var routerRegistry = make(map[string]router.Router)

// Register a router with a name.
// If a router with the same name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterRouter(name string, r router.Router, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	if !options.allowOverride {
		if _, exists := routerRegistry[name]; exists {
			panic("router " + name + " already registered")
		}
	}
	routerRegistry[name] = r
}

// Retrieve a router by name.
// If the router does not exist, it returns nil.
func GetRouter(name string) router.Router {
	if r, ok := routerRegistry[name]; ok {
		return r
	}
	return nil
}

// GetRouterRegistry returns the router registry for Router Integration
func GetRouterRegistry() map[string]router.Router {
	return routerRegistry
}
