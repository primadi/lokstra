package old_registry

import (
	"sync"

	"github.com/primadi/lokstra/core/server"
)

// Server registry using interface to avoid circular dependency
type ServerInterface interface {
	GetName() string
	Start() error
	Shutdown(timeout any) error
}

var serverRegistry sync.Map

// Register a server with a name.
// If a server with the same name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterServer(name string, srv *server.Server, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	if !options.allowOverride {
		if _, exists := serverRegistry.Load(name); exists {
			panic("server " + name + " already registered")
		}
	}
	serverRegistry.Store(name, srv)
}

// Retrieve a server by name.
// If the server does not exist, it returns nil.
func GetServer(name string) *server.Server {
	if srvAny, ok := serverRegistry.Load(name); ok {
		return srvAny.(*server.Server)
	}
	return nil
}

// List all registered server names
func ListServerNames() []string {
	names := make([]string, 0)
	serverRegistry.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}
