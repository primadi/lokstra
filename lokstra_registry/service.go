package lokstra_registry

import (
	"fmt"
)

var serviceRegistry = make(map[string]any)

type lazyServiceConfig struct {
	serviceType string
	config      map[string]any
}

var lazyServiceConfigRegistry = make(map[string]lazyServiceConfig)

// Registers a service instance with a given name.
// If the same service name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterService(svcName string, svcInstance any, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	if !options.allowOverride {
		if _, exists := serviceRegistry[svcName]; exists {
			panic("service " + svcName + " already registered")
		}
	}

	serviceRegistry[svcName] = svcInstance
}

// Registers a lazy service configuration with a given name.
// The actual service instance will be created when first requested via LazyGetService.
// If the same lazy service name already exists,
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterLazyService(svcName string, svcType string,
	config map[string]any, opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	if !options.allowOverride {
		if _, exists := lazyServiceConfigRegistry[svcName]; exists {
			panic("lazy service " + svcName + " already registered")
		}
		if _, exists := serviceRegistry[svcName]; exists {
			panic("service " + svcName + " already registered")
		}
	}
	lazyServiceConfigRegistry[svcName] = lazyServiceConfig{
		serviceType: svcType,
		config:      config,
	}
}

// Tries to resolve a service from the registry and assign it to current.
//
// If current is already set (non-nil), it will be returned as is.
// Otherwise, it will attempt to get from registry and set it.
// If current is nil and not found in registry, it tries to create from lazy config if exists.
// If fail to create, it will panic.
// It will panic if the type in registry does not match T.
func GetService[T comparable](name string, current T) T {
	s, ok := TryGetService(name, current)
	if !ok {
		panic(fmt.Sprintf("service %s not found or type mismatch", name))
	}
	return s
}

// Tries to resolve a service from the registry.
//
// If current != nil, it will be returned immediately with ok=true.
// If current is nil and found in registry with correct type, it will be returned with ok=true.
// If not found or type mismatch, it tries to create from lazy config if exists.
// If still not found, it returns zero value of T with ok=false.
func TryGetService[T comparable](svcName string, current T) (T, bool) {
	// if current is already set (non-nil), return as is
	var zero T
	if current != zero {
		return current, true
	}

	// lookup in registry
	if svc, ok := serviceRegistry[svcName]; ok {
		if typed, ok := svc.(T); ok {
			return typed, true
		}
		// type mismatch
		return zero, false
	}

	// not found, check if lazy config exists
	if lazyCfg, ok := lazyServiceConfigRegistry[svcName]; ok {
		if factory := GetServiceFactory(lazyCfg.serviceType); factory != nil {
			if svc := factory(lazyCfg.config); svc != nil {
				return svc.(T), true
			}
		}
	}

	// not found
	return zero, false
}

// Create a new service using registered factory, register it, and return it.
// If factory not found or creation failed, return zero value of T.
// It will panic if the created service type does not match T or if
// a service with the same name already exists, unless
// the RegisterOption allowOverride is set to true.
func NewService[T any](svcName, svcType string, config map[string]any,
	opts ...RegisterOption) T {
	if factory := GetServiceFactory(svcType); factory != nil {
		if svc := factory(config); svc != nil {
			RegisterService(svcName, svc, opts...)
			return svc.(T)
		}
	}
	var zero T
	return zero
}
