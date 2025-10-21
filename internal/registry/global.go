package registry

import "sync"

// GlobalRegistryInstance is the interface for accessing the global registry
// This allows core packages to access the registry without circular dependencies
type GlobalRegistryInstance interface {
	GetServiceAny(name string) (any, bool)
	RegisterService(name string, service any)
}

var (
	instance     GlobalRegistryInstance
	instanceOnce sync.Once
)

// SetGlobal sets the global registry instance (called once by deploy.Global())
func SetGlobal(reg GlobalRegistryInstance) {
	instanceOnce.Do(func() {
		instance = reg
	})
}

// Global returns the global registry instance
// No mutex needed: sync.Once in SetGlobal guarantees instance is set before any reads
func Global() GlobalRegistryInstance {
	return instance
}

// HasGlobal returns true if the global registry has been initialized
func HasGlobal() bool {
	return instance != nil
}
