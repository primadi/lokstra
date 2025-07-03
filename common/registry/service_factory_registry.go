package registry

import (
	"lokstra/common/iface"
	"lokstra/common/permission"
	"strings"
)

type ServiceFactory = func(config any) (iface.Service, error)

var serviceFactories = make(map[string]ServiceFactory) // map of serviceType to ServiceFactory

// GetServiceFactory retrieves a registered service factory by its serviceType.
func GetServiceFactory(serviceType string) (ServiceFactory, bool) {
	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	sf, exists := serviceFactories[serviceType]
	return sf, exists
}

// RegisterServiceFactory registers a new service factory with the given serviceType.
func RegisterServiceFactory(serviceType string, serviceFactory func(config any) (iface.Service, error),
	lic ...*permission.PermissionLicense) {
	if permission.GlobalAccessLocked() {
		if lic == nil || lic[0] == nil || !strings.HasPrefix(serviceType, lic[0].GetModuleName()+":") {
			panic("cannot register service after global access is locked or service is not created in the same module")
		}
	}

	if serviceFactory == nil {
		panic("service factory cannot be nil")
	}
	if serviceType == "" {
		panic("serviceType cannot be empty")
	}

	if !strings.Contains(serviceType, ".") {
		serviceType = "main." + serviceType
	}

	if _, exists := serviceFactories[serviceType]; exists {
		panic("service factory with serviceType '" + serviceType + "' already exists")
	}

	serviceFactories[serviceType] = serviceFactory
}

// ResetServiceFactories clears all registered services factories.
// This is useful for testing or reinitializing the registry.
func ResetServiceFactories() {
	if permission.GlobalAccessLocked() {
		panic("cannot reset service factories after global access is locked")
	}

	// Clear the service factories map
	serviceFactories = make(map[string]ServiceFactory)
}
