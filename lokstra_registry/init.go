package lokstra_registry

import (
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/server"
)

// Initialize callback to avoid circular dependency
func init() {
	// Set callback from registry to server package
	server.SetShutdownServicesCallback(ShutdownServices)

	// Register CFG resolver for two-pass variable expansion
	// This allows ${@CFG:config-name} syntax in YAML configs
	cfgResolver := config.NewConfigResolver(&ConfigRegistryGetter{})
	config.AddVariableResolver("CFG", cfgResolver)
}
