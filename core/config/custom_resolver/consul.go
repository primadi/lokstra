package custom_resolver

import (
	"log"

	lokstraConfig "github.com/primadi/lokstra/core/config"
)

// ConsulResolver implements VariableResolver for Consul KV Store
type ConsulResolver struct {
	address string
	prefix  string
	cache   map[string]string
}

// creates a new Consul KV resolver
func NewConsulResolver(address, prefix string) *ConsulResolver {
	return &ConsulResolver{
		address: address,
		prefix:  prefix,
		cache:   make(map[string]string),
	}
}

// retrieves value from Consul KV store
func (r *ConsulResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	// Check cache first
	if value, ok := r.cache[key]; ok {
		return value, true
	}

	// Implementation would use Consul HTTP API or SDK
	// For brevity, this is a simplified example

	// Example: GET http://consul.example.com/v1/kv/{prefix}/{key}

	log.Printf("Would fetch from Consul: %s/%s (using default for now)", r.prefix, key)
	return defaultValue, false
}

var _ lokstraConfig.VariableResolver = (*ConsulResolver)(nil)

// Example usage
func ExampleConsulResolver() {
	// Register Consul resolver
	consulResolver := NewConsulResolver("http://consul.example.com:8500", "lokstra")
	lokstraConfig.AddVariableResolver("CONSUL", consulResolver)

	// Now you can use ${@CONSUL:key} in your YAML configs
	// Example: config/production.yaml
	//
	// services:
	//   - name: api
	//     config:
	//       featureFlags: ${@CONSUL:feature-flags}
	//       rateLimits: ${@CONSUL:rate-limits}
}
