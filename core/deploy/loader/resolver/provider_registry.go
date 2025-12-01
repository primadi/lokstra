package resolver

import "sync"

// Provider interface for custom value providers
// Providers resolve keys to values from various sources (env, aws-secret, vault, k8s, etc.)
type Provider interface {
	// Name returns the provider name (e.g., "env", "aws-secret", "vault", "k8s")
	Name() string

	// Resolve resolves a key to its value
	// Returns the resolved value and whether it was found
	Resolve(key string) (string, bool)
}

// Provider registry
var (
	providers   = make(map[string]Provider)
	providersMu sync.RWMutex
)

// RegisterProvider registers a custom provider for config resolution
// Examples:
//   - RegisterProvider(&AWSSecretProvider{}) -> resolve ${@aws-secret:key}
//   - RegisterProvider(&VaultProvider{}) -> resolve ${@vault:path}
//   - RegisterProvider(&K8sConfigMapProvider{}) -> resolve ${@k8s:configmap/key}
func RegisterProvider(p Provider) {
	providersMu.Lock()
	defer providersMu.Unlock()
	providers[p.Name()] = p
}

// GetProvider retrieves a provider by name
func GetProvider(name string) Provider {
	providersMu.RLock()
	defer providersMu.RUnlock()
	return providers[name]
}

// init registers default providers
func init() {
	// Register default @env provider
	RegisterProvider(&envProvider{})
}
