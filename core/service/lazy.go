package service

import (
	"sync"

	"github.com/primadi/lokstra/core/config"
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
	cache       *T
}

// NewLazy creates a new lazy service loader for the given service name.
func NewLazy[T any](serviceName string) *Lazy[T] {
	return &Lazy[T]{
		serviceName: serviceName,
	}
}

// Get retrieves the service instance. The service is initialized on first call
// and cached for subsequent calls. This method is thread-safe.
func (l *Lazy[T]) Get() *T {
	l.once.Do(func() {
		service := lokstra_registry.GetService[*T](l.serviceName, nil)
		l.cache = service
	})
	return l.cache
}

// MustGet retrieves the service instance or panics if the service is not found.
func (l *Lazy[T]) MustGet() *T {
	svc := l.Get()
	if svc == nil {
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
	return l.cache != nil
}

// GetLazyService retrieves a lazy service from config.
// It handles both old-style string references and new GenericLazyService instances.
//
// Example usage in a factory:
//
//	func NewUserRepo(cfg map[string]interface{}) (*UserRepo, error) {
//	    dbLazy := service.GetLazyService[Database](cfg, "db_service")
//	    return &UserRepo{
//	        db: dbLazy,  // type: *service.Lazy[Database]
//	    }, nil
//	}
//
// In the service method:
//
//	func (r *UserRepo) FindUser(id string) (*User, error) {
//	    db := r.db.Get()  // Lazy loaded, cached, type-safe
//	    return db.QueryUser(id)
//	}
func GetLazyService[T any](cfg map[string]interface{}, key string) *Lazy[T] {
	if cfg == nil {
		return nil
	}

	val, ok := cfg[key]
	if !ok {
		return nil
	}

	// Check if it's already a GenericLazyService (layered mode with depends-on)
	if generic, ok := val.(*config.GenericLazyService); ok {
		return NewLazy[T](generic.ServiceName())
	}

	// Check if it's a string reference (simple mode or old config)
	if str, ok := val.(string); ok {
		return NewLazy[T](str)
	}

	// Not a valid service reference
	return nil
}
