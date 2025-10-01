package lokstra_registry

import (
	"github.com/primadi/lokstra/core/server"
)

// Initialize callback to avoid circular dependency
func init() {
	// Set callback from registry to server package
	server.SetShutdownServicesCallback(ShutdownServices)
}
