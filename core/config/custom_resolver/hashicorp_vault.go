package custom_resolver

import (
	"log"

	lokstraConfig "github.com/primadi/lokstra/core/config"
)

// VaultResolver implements VariableResolver for HashiCorp Vault
type VaultResolver struct {
	address string
	token   string
	cache   map[string]string
}

// creates a new Vault resolver
func NewVaultResolver(address, token string) *VaultResolver {
	return &VaultResolver{
		address: address,
		token:   token,
		cache:   make(map[string]string),
	}
}

// retrieves secret from Vault
func (r *VaultResolver) Resolve(source string, key string, defaultValue string) (string, bool) {
	// Check cache first
	if value, ok := r.cache[key]; ok {
		return value, true
	}

	// Implementation would use Vault HTTP API or SDK
	// For brevity, this is a simplified example

	// Example: GET https://vault.example.com/v1/secret/data/{key}
	// with X-Vault-Token header

	log.Printf("Would fetch from Vault: %s (using default for now)", key)
	return defaultValue, false
}

var _ lokstraConfig.VariableResolver = (*VaultResolver)(nil)

// Example usage
func ExampleVaultResolver() {
	// Register Vault resolver
	vaultResolver := NewVaultResolver("https://vault.example.com", "vault-token")
	lokstraConfig.AddVariableResolver("VAULT", vaultResolver)

	// Now you can use ${@VAULT:path/to/secret} in your YAML configs
	// Example: config/production.yaml
	//
	// services:
	//   - name: postgres
	//     type: dbpool_pg
	//     config:
	//       password: ${@VAULT:database/postgres/password}
	//
	//   - name: api
	//     config:
	//       apiKey: ${@VAULT:api/keys/production}
}
