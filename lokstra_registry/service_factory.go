package lokstra_registry

import (
	"fmt"
	"sync"
)

type ServiceFactory = func(config map[string]any) any

// Service factory registry - separate for local and remote
type serviceFactoryEntry struct {
	localFactory  ServiceFactory
	remoteFactory ServiceFactory
}

var serviceFactoryRegistry sync.Map

func RegisterServiceFactoryLocalAndRemote(serviceType string, localFactory, remoteFactory ServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	entryAny, exists := serviceFactoryRegistry.Load(serviceType)
	var entry *serviceFactoryEntry

	if !exists {
		entry = &serviceFactoryEntry{}
	} else {
		entry = entryAny.(*serviceFactoryEntry)
		if !options.allowOverride && (entry.localFactory != nil || entry.remoteFactory != nil) {
			panic("service factory for type " + serviceType + " already registered")
		}
	}

	entry.localFactory = localFactory
	entry.remoteFactory = remoteFactory

	serviceFactoryRegistry.Store(serviceType, entry)
}

// RegisterServiceFactory registers BOTH local and remote factory (backward compatibility)
// Deprecated: Use RegisterServiceFactoryLocal and RegisterServiceFactoryRemote instead
func RegisterServiceFactory(serviceType string, factory ServiceFactory,
	opts ...RegisterOption) {
	RegisterServiceFactoryLocal(serviceType, factory, opts...)
}

// RegisterServiceFactoryLocal registers a factory for creating LOCAL service instances
// This factory will be used when the service needs to run in the same process
func RegisterServiceFactoryLocal(serviceType string, factory ServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	entryAny, exists := serviceFactoryRegistry.Load(serviceType)
	var entry *serviceFactoryEntry

	if !exists {
		entry = &serviceFactoryEntry{}
	} else {
		entry = entryAny.(*serviceFactoryEntry)
		if !options.allowOverride && entry.localFactory != nil {
			panic("local service factory for type " + serviceType + " already registered")
		}
	}

	entry.localFactory = factory
	serviceFactoryRegistry.Store(serviceType, entry)
}

// RegisterServiceFactoryRemote registers a factory for creating REMOTE service client instances
// This factory will be used when the service needs to call a remote instance via HTTP
func RegisterServiceFactoryRemote(serviceType string, factory ServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}

	entryAny, exists := serviceFactoryRegistry.Load(serviceType)
	var entry *serviceFactoryEntry

	if !exists {
		entry = &serviceFactoryEntry{}
	} else {
		entry = entryAny.(*serviceFactoryEntry)
		if !options.allowOverride && entry.remoteFactory != nil {
			panic("remote service factory for type " + serviceType + " already registered")
		}
	}

	entry.remoteFactory = factory
	serviceFactoryRegistry.Store(serviceType, entry)
}

// GetServiceFactory retrieves the appropriate factory (local or remote) based on service location
// Returns local factory if service should run locally, remote factory otherwise
// Framework automatically decides based on:
// - Router registration (same server = local, different server = remote)
// - ClientRouter metadata (IsLocal flag)
// serviceName is the instance name (e.g., "user-service"), serviceType is the factory type (e.g., "user_service")
func GetServiceFactory(serviceType string, serviceName string) ServiceFactory {
	entryAny, ok := serviceFactoryRegistry.Load(serviceType)
	if !ok {
		return nil
	}

	entry := entryAny.(*serviceFactoryEntry)

	// Determine if we need local or remote factory
	// Strategy: Check if there's a ClientRouter for this service name (NOT serviceType)
	// If ClientRouter exists and IsLocal=true -> use local
	// If ClientRouter exists and IsLocal=false -> use remote
	// If no ClientRouter -> default to local

	client := GetClientRouter(serviceName) // Use serviceName instead of serviceType
	// fmt.Printf("[DEBUG GetServiceFactory] serviceType=%s, serviceName=%s, client=%v, IsLocal=%v",
	// 	serviceType, serviceName, client != nil, client != nil && client.IsLocal)
	// if client != nil {
	// 	fmt.Printf(", ServerName=%s, currentServer=%s\n", client.ServerName, GetCurrentServerName())
	// } else {
	// 	fmt.Printf("\n")
	// }
	if client != nil {
		if client.IsLocal {
			// Service is on same server - use local factory
			if entry.localFactory == nil {
				panic(fmt.Sprintf("service %s requires local factory but only remote is registered", serviceType))
			}
			return entry.localFactory
		}
		// Service is on different server - use remote factory
		if entry.remoteFactory == nil {
			panic(fmt.Sprintf("service %s requires remote factory but only local is registered", serviceType))
		}
		return entry.remoteFactory
	}

	// No ClientRouter found - default to local
	if entry.localFactory != nil {
		return entry.localFactory
	}

	// Fallback to remote if local not available
	return entry.remoteFactory
}

// GetServiceFactoryLocal explicitly gets the local factory (for testing/debugging)
func GetServiceFactoryLocal(serviceType string) ServiceFactory {
	if entryAny, ok := serviceFactoryRegistry.Load(serviceType); ok {
		entry := entryAny.(*serviceFactoryEntry)
		return entry.localFactory
	}
	return nil
}

// GetServiceFactoryRemote explicitly gets the remote factory (for testing/debugging)
func GetServiceFactoryRemote(serviceType string) ServiceFactory {
	if entryAny, ok := serviceFactoryRegistry.Load(serviceType); ok {
		entry := entryAny.(*serviceFactoryEntry)
		return entry.remoteFactory
	}
	return nil
}
