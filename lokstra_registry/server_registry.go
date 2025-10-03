package lokstra_registry

import (
	"github.com/primadi/lokstra/core/server"
)

// Server registry using interface to avoid circular dependency
type ServerInterface interface {
	GetName() string
	Start() error
	Shutdown(timeout interface{}) error
}

var serverRegistry = make(map[string]*server.Server)

// Register a server with a name.
// If a server with the same name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterServer(name string, srv *server.Server, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	if !options.allowOverride {
		if _, exists := serverRegistry[name]; exists {
			panic("server " + name + " already registered")
		}
	}
	serverRegistry[name] = srv
}

// Retrieve a server by name.
// If the server does not exist, it returns nil.
func GetServer(name string) *server.Server {
	if srv, ok := serverRegistry[name]; ok {
		return srv
	}
	return nil
}

// List all registered server names
func ListServerNames() []string {
	names := make([]string, 0, len(serverRegistry))
	for name := range serverRegistry {
		names = append(names, name)
	}
	return names
}
