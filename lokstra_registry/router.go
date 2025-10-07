package lokstra_registry

import (
	"maps"
	"sync"

	"github.com/primadi/lokstra/core/router"
)

var routerRegistry = make(map[string]router.Router)
var routerMutex sync.RWMutex

// Register a router with a name.
// If a router with the same name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterRouter(name string, r router.Router, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	routerMutex.Lock()
	defer routerMutex.Unlock()

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
	routerMutex.RLock()
	defer routerMutex.RUnlock()

	if r, ok := routerRegistry[name]; ok {
		return r
	}
	return nil
}

// GetRouterRegistry returns the router registry for Router Integration
// Returns a copy to prevent concurrent map access issues
func GetRouterRegistry() map[string]router.Router {
	routerMutex.RLock()
	defer routerMutex.RUnlock()

	// Return a copy to avoid concurrent map iteration issues
	copy := make(map[string]router.Router, len(routerRegistry))
	maps.Copy(copy, routerRegistry)
	return copy
}
