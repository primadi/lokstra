package service

import (
	"sync"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Lazy provides a type-safe lazy-loading service container.
// The service is only initialized on first Get() call and cached thereafter.
//
// Example usage:
//
//	type MyService struct {
//	    db *service.Lazy[Database]
//	}
//
//	func (s *MyService) DoSomething() {
//	    db := s.db.Get()  // Lazy loaded, cached, type-safe
//	    db.Query("...")
//	}
type Lazy[T any] struct {
	serviceName string
	once        sync.Once
	cache       T
}

// creates a new lazy service loader for the given service name.
func LazyLoad[T any](serviceName string) *Lazy[T] {
	return &Lazy[T]{
		serviceName: serviceName,
	}
}

// Get retrieves the service instance. The service is initialized on first call
// and cached for subsequent calls. This method is thread-safe.
func (l *Lazy[T]) Get() T {
	l.once.Do(func() {
		service := lokstra_registry.GetService[T](l.serviceName)
		l.cache = service
	})
	return l.cache
}

// MustGet retrieves the service instance or panics if the service is not found.
func (l *Lazy[T]) MustGet() T {
	svc := l.Get()
	if utils.IsNil(svc) {
		panic("service '" + l.serviceName + "' not found or not initialized")
	}
	return svc
}

// ServiceName returns the name of the service being lazily loaded.
func (l *Lazy[T]) ServiceName() string {
	return l.serviceName
}

// IsLoaded returns true if the service has been loaded (Get was called at least once).
func (l *Lazy[T]) IsLoaded() bool {
	return !utils.IsNil(l.cache)
}

// creates a lazy service loader from factory service configuration map.
func LazyLoadFromConfig[T any](cfg map[string]any, key string) *Lazy[T] {
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
func MustLazyLoadFromConfig[T any](cfg map[string]any, key string) *Lazy[T] {
	lazy := LazyLoadFromConfig[T](cfg, key)
	if lazy == nil {
		panic("missing required dependency '" + key + "'")
	}
	return lazy
}
