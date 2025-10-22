// Package lokstra_registry provides a simplified API for the Lokstra registry system.
//
// This package wraps deploy.GlobalRegistry() to provide a cleaner developer experience:
// - Shorter import path
// - Package-level functions instead of singleton access
// - Generic helper functions like GetService[T]
//
// Example usage:
//
//	import "github.com/primadi/lokstra/lokstra_registry"
//
//	lokstra_registry.RegisterService("user-service", userSvc)
//	userSvc := lokstra_registry.GetService[*UserService]("user-service")
package lokstra_registry

import (
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// Global returns the global registry instance
func Global() *deploy.GlobalRegistry {
	return deploy.Global()
}

// ===== ROUTER =====

// RegisterRouter registers a router instance in the runtime registry
func RegisterRouter(name string, r router.Router) {
	deploy.Global().RegisterRouter(name, r)
}

// GetRouter retrieves a router instance from the runtime registry
func GetRouter(name string) router.Router {
	return deploy.Global().GetRouter(name)
}

// GetAllRouters returns all registered router instances
func GetAllRouters() map[string]router.Router {
	return deploy.Global().GetAllRouters()
}

// ===== SERVICE =====

// RegisterServiceType registers a service factory in the global registry
// local: factory for local service instances
// remote: factory for remote service proxies (can be nil)
// options: optional metadata for auto-router generation
func RegisterServiceType(serviceType string, local, remote deploy.ServiceFactory, options ...deploy.RegisterServiceTypeOption) {
	deploy.Global().RegisterServiceType(serviceType, local, remote, options...)
}

// GetServiceFactory returns the service factory for a service type
// isLocal: true for local factory, false for remote factory
func GetServiceFactory(serviceType string, isLocal bool) deploy.ServiceFactory {
	return deploy.Global().GetServiceFactory(serviceType, isLocal)
}

// DefineService defines a service in the global registry (for code-based config)
func DefineService(def *schema.ServiceDef) {
	deploy.Global().DefineService(def)
}

// RegisterService registers a service instance in the runtime registry
func RegisterService(name string, instance any) {
	deploy.Global().RegisterService(name, instance)
}

// GetServiceAny retrieves a service instance (non-generic version)
func GetServiceAny(name string) (any, bool) {
	return deploy.Global().GetServiceAny(name)
}

// GetService retrieves a service instance with type assertion
// Returns zero value if not found or type mismatch
func GetService[T any](name string) T {
	instance, ok := deploy.Global().GetServiceAny(name)
	if !ok {
		var zero T
		return zero
	}

	if typed, ok := instance.(T); ok {
		return typed
	}

	var zero T
	return zero
}

// MustGetService retrieves a service instance with type assertion
// Panics if not found or type mismatch
func MustGetService[T any](name string) T {
	svc, ok := TryGetService[T](name)
	if !ok {
		panic("service " + name + " not found or type mismatch")
	}
	return svc
}

// TryGetService retrieves a service instance with type assertion
// Returns (value, true) if found and type matches, (zero, false) otherwise
func TryGetService[T any](name string) (T, bool) {
	instance, ok := deploy.Global().GetServiceAny(name)
	if !ok {
		var zero T
		return zero, false
	}

	if typed, ok := instance.(T); ok {
		return typed, true
	}

	var zero T
	return zero, false
}

// ===== MIDDLEWARE =====

// RegisterMiddleware registers a middleware instance in the runtime registry
func RegisterMiddleware(name string, handler request.HandlerFunc) {
	deploy.Global().RegisterMiddleware(name, handler)
}

// GetMiddleware retrieves a middleware instance from the runtime registry
func GetMiddleware(name string) (request.HandlerFunc, bool) {
	return deploy.Global().GetMiddleware(name)
}

// CreateMiddleware creates a middleware from its definition and caches it
func CreateMiddleware(name string) request.HandlerFunc {
	return deploy.Global().CreateMiddleware(name)
}

// ===== CONFIGURATION =====

// DefineConfig defines a configuration value in the global registry (for YAML config)
func DefineConfig(name string, value any) {
	deploy.Global().DefineConfig(&schema.ConfigDef{
		Name:  name,
		Value: value,
	})
}

// GetResolvedConfig gets a resolved configuration value from the global registry
func GetResolvedConfig(key string) (any, bool) {
	return deploy.Global().GetResolvedConfig(key)
}

// GetConfig retrieves a configuration value with type assertion and default value
func GetConfig[T any](name string, defaultValue T) T {
	value, ok := deploy.Global().GetResolvedConfig(name)
	if !ok {
		return defaultValue
	}

	if typed, ok := value.(T); ok {
		return typed
	}

	return defaultValue
}
