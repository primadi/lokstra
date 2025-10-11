package lokstra_registry

import (
	"sync"

	"github.com/primadi/lokstra/core/router"
)

// routerRegistry uses sync.Map for better concurrent read performance (14.6x faster than RWMutex)
// Write once during startup, read many times per request - perfect use case for sync.Map
var routerRegistry sync.Map

// Register a router with a name.
// If a router with the same name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterRouter(name string, r router.Router, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	if !options.allowOverride {
		if _, exists := routerRegistry.Load(name); exists {
			panic("router " + name + " already registered")
		}
	}
	routerRegistry.Store(name, r)
}

// Retrieve a router by name.
// If the router does not exist, it returns nil.
func GetRouter(name string) router.Router {
	if v, ok := routerRegistry.Load(name); ok {
		return v.(router.Router)
	}
	return nil
}

// GetRouterRegistry returns the router registry for Router Integration
// Returns a copy to prevent concurrent map access issues
func GetRouterRegistry() map[string]router.Router {
	copy := make(map[string]router.Router)
	routerRegistry.Range(func(key, value any) bool {
		copy[key.(string)] = value.(router.Router)
		return true
	})
	return copy
}
