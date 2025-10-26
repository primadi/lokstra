package config

import (
	"fmt"
)

// ConfigResolver resolves variables from the config registry
// This requires lokstra_registry to be imported and SetConfig to be called first
//
// Example usage in YAML:
//   services:
//     - name: api
//       config:
//         baseUrl: ${@CFG:api.baseUrl:http://localhost:8080}
//         timeout: ${@CFG:api.timeout:30}
//
// This is useful for:
// - Referencing common configuration values
// - Avoiding duplication in config files
// - Dynamic configuration based on environment
//
// Note: CFG resolver uses two-pass expansion:
//   Pass 1: All other resolvers (ENV, AWS, K8S, etc.) are expanded
//   Pass 2: Config registry is populated via SetConfig
//   Pass 3: CFG resolver is expanded using ExpandCFGVariablesInConfig()

// ConfigGetter is an interface for getting config values
// This is implemented by lokstra_registry
type ConfigGetter interface {
	GetConfig(key string) (any, bool)
}

// ConfigResolver resolves from config registry
type ConfigResolver struct {
	getter ConfigGetter
}

// NewConfigResolver creates a new config resolver with a config getter
func NewConfigResolver(getter ConfigGetter) *ConfigResolver {
	return &ConfigResolver{
		getter: getter,
	}
}

// Resolve gets value from config registry
func (r *ConfigResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	if r.getter == nil {
		// Config registry not set yet, return default
		return defaultValue, false
	}

	value, found := r.getter.GetConfig(key)
	if !found {
		return defaultValue, false
	}

	// Convert value to string
	str := fmt.Sprint(value)
	return str, true
}
