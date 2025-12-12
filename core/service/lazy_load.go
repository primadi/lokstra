package service

import (
	"sync"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/internal/registry"
)

// Cached provides a type-safe lazy-loading service container.
// The service is only initialized on first Get() call and cached thereafter.
//
// Example usage:
//
//	type MyService struct {
//	    db *service.Cached[Database]
//	}
//
//	func (s *MyService) DoSomething() {
//	    db := s.db.Get()  // Cached loaded, cached, type-safe
//	    db.Query("...")
//	}
type Cached[T any] struct {
	serviceName string
	loader      func() T
	once        sync.Once
	cache       T
}

// LazyLoad creates a new lazy service loader for the given service name.
// The service will be automatically loaded from the global registry on first Get() call.
//
// Example usage:
//
//	type MyService struct {
//	    db *service.Cached[*DBPool]
//	}
//
//	func NewMyService() *MyService {
//	    return &MyService{
//	        db: service.LazyLoad[*DBPool]("db-pool"),
//	    }
//	}
//
//	func (s *MyService) Query() {
//	    db := s.db.Get() // Loaded from registry on first call
//	    db.Query("...")
//	}
func LazyLoad[T any](serviceName string) *Cached[T] {
	return &Cached[T]{
		serviceName: serviceName,
		loader: func() T {
			// Load from global registry
			if reg := registry.Global(); reg != nil {
				if svc, ok := reg.GetServiceAny(serviceName); ok {
					if typed, ok := svc.(T); ok {
						return typed
					}
				}
			}
			// Return zero value if not found
			var zero T
			return zero
		},
	}
}

// Get retrieves the service instance. The service is initialized on first call
// and cached for subsequent calls. This method is thread-safe.
func (l *Cached[T]) Get() T {
	l.once.Do(func() {
		if l.loader != nil {
			// Custom loader
			l.cache = l.loader()

			// Log when service is loaded
			if l.serviceName != "" && !utils.IsNil(l.cache) {
				logger.LogDebug("🔧 Lazy loaded service: '%s'", l.serviceName)
			}
		} else {
			// No loader provided - return zero value
			var zero T
			l.cache = zero
		}
	})
	return l.cache
}

// MustGet retrieves the service instance or panics if the service is not found.
func (l *Cached[T]) MustGet() T {
	svc := l.Get()
	if utils.IsNil(svc) {
		panic("service '" + l.serviceName + "' not found or not initialized")
	}
	return svc
}

// ServiceName returns the name of the service being lazily loaded.
func (l *Cached[T]) ServiceName() string {
	return l.serviceName
}

// IsLoaded returns true if the service has been loaded (Get was called at least once).
func (l *Cached[T]) IsLoaded() bool {
	return !utils.IsNil(l.cache)
}

// creates a lazy service loader from factory service configuration map.
func LazyLoadFromConfig[T any](cfg map[string]any, key string) *Cached[T] {
	if cfg == nil {
		return nil
	}

	val, ok := cfg[key]
	if !ok {
		return nil
	}

	if svcName, ok := val.(string); ok {
		return LazyLoad[T](svcName)
	}

	// Not a valid service reference
	return nil
}

// creates a lazy service loader from factory service configuration map.
// It panics if the key is missing or invalid.
func MustLazyLoadFromConfig[T any](cfg map[string]any, key string) *Cached[T] {
	lazy := LazyLoadFromConfig[T](cfg, key)
	if lazy == nil {
		panic("missing required dependency '" + key + "'")
	}
	return lazy
}

// LazyLoadWith creates a lazy service loader with a custom loader function.
// The loader function is called on first Get() and the result is cached.
// This is useful for dependency injection frameworks that manage their own service resolution.
//
// Example usage:
//
//	deps["db"] = service.LazyLoadWith(func() any {
//	    return app.GetService("db-pool")
//	})
func LazyLoadWith[T any](loader func() T) *Cached[T] {
	return &Cached[T]{
		loader: loader,
	}
}

// Value creates a Cached instance with a pre-loaded value (no lazy loading).
// Useful for testing or when the value is already available.
func Value[T any](value T) *Cached[T] {
	c := &Cached[T]{
		cache: value,
	}
	// Mark as loaded by setting a no-op loader
	c.loader = func() T {
		return value
	}
	// Execute once.Do to mark as loaded
	c.once.Do(func() {})
	return c
}

// Cast converts a dependency value from map[string]any to a typed Cached[T].
// This is a helper for factory functions that receive deps as map[string]any.
//
// Example usage in factory:
//
//	func userServiceFactory(deps map[string]any, config map[string]any) any {
//	    return &UserService{
//	        DB:     service.Cast[*DBPool](deps["db"]),
//	        Logger: service.Cast[*Logger](deps["logger"]),
//	    }
//	}
func Cast[T any](value any) *Cached[T] {
	if cached, ok := value.(*Cached[any]); ok {
		// Wrap the any Cached with a typed loader, preserving serviceName
		return &Cached[T]{
			serviceName: cached.serviceName, // Preserve service name for logging
			loader: func() T {
				return cached.Get().(T)
			},
		}
	}
	// If it's already the right type, return as-is
	if cached, ok := value.(*Cached[T]); ok {
		return cached
	}
	panic("Cast: value is not a *Cached type")
}

// LazyLoadFrom creates a Cached that lazy-loads a service from a ServiceGetter.
// This is useful for loading services from deployment apps.
//
// Example usage:
//
// app := dep.GetServerApp("api", "crud-api")
// userService := service.LazyLoadFrom[*UserService](app, "user-service")
// // Service is loaded on first Get() call
// users := userService.MustGet().GetAll()
type ServiceGetter interface {
	GetService(serviceName string) (any, error)
}

func LazyLoadFrom[T any](getter ServiceGetter, serviceName string) *Cached[T] {
	return LazyLoadWith(func() T {
		svc, err := getter.GetService(serviceName)
		if err != nil {
			panic("failed to load service '" + serviceName + "': " + err.Error())
		}
		return svc.(T)
	})
}

// CastProxyService casts a dependency value to *proxy.Service
// This is used in remote service factories where the framework pre-instantiates proxy.Service
//
// Example usage:
//
//	func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
//	    return &UserServiceRemote{
//	        service: service.CastProxyService(deps["remote"]),
//	    }
//	}
func CastProxyService(value any) *proxy.Service {
	// Allow nil for metadata reading scenarios
	if value == nil {
		return nil
	}
	if svc, ok := value.(*proxy.Service); ok {
		return svc
	}
	panic("CastProxyService: value is not a *proxy.Service")
}
