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
	"fmt"

	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/service"
)

// ===== MIDDLEWARE FACTORY (compatible with old_registry) =====

// MiddlewareFactory is a function that creates a middleware handler from config
// For compatibility with old_registry, we use the specific signature
// But it's actually an alias to deploy.MiddlewareFactory which returns any
type MiddlewareFactory = deploy.MiddlewareFactory

// registerOptions holds options for registration functions
type registerOptions struct {
	allowOverride bool
}

// RegisterOption is an interface for registration options
type RegisterOption interface {
	apply(opt *registerOptions)
}

type allowOverrideOption struct {
	allowOverride bool
}

func (o *allowOverrideOption) apply(opt *registerOptions) {
	opt.allowOverride = o.allowOverride
}

// AllowOverride returns a RegisterOption that allows overriding existing registrations
func AllowOverride(enable bool) RegisterOption {
	return &allowOverrideOption{allowOverride: enable}
}

// RegisterMiddlewareFactory registers a middleware factory function for a given middleware type.
// This is a helper function that wraps deploy.Global().RegisterMiddlewareType().
//
// For compatibility with old_registry pattern where factories return request.HandlerFunc,
// this function accepts both:
//   - func(config map[string]any) request.HandlerFunc (old pattern)
//   - func(config map[string]any) any (new pattern)
//
// Example:
//
//	lokstra_registry.RegisterMiddlewareFactory("logger", loggerFactory,
//	    lokstra_registry.AllowOverride(true))
func RegisterMiddlewareFactory(mwType string, factory any, opts ...RegisterOption) {
	var deployOpts []deploy.MiddlewareTypeOption
	for _, opt := range opts {
		if allowOpt, ok := opt.(*allowOverrideOption); ok {
			deployOpts = append(deployOpts, deploy.WithAllowOverride(allowOpt.allowOverride))
		}
	}

	// Convert factory to deploy.MiddlewareFactory (returns any) if needed
	var deployFactory deploy.MiddlewareFactory

	// Check if it's already the right signature
	if f, ok := factory.(func(map[string]any) any); ok {
		deployFactory = f
	} else if f, ok := factory.(func(map[string]any) request.HandlerFunc); ok {
		// Wrap old-style factory that returns request.HandlerFunc
		deployFactory = func(config map[string]any) any {
			return f(config)
		}
	} else {
		panic(fmt.Sprintf("invalid middleware factory signature for type %s: must be func(map[string]any) any or func(map[string]any) request.HandlerFunc", mwType))
	}

	deploy.Global().RegisterMiddlewareType(mwType, deployFactory, deployOpts...)
}

// RegisterMiddlewareName registers a middleware entry by name, associating it with a type and config.
// This is a helper function that wraps deploy.Global().RegisterMiddlewareName().
//
// Example:
//
//	lokstra_registry.RegisterMiddlewareName("logger-debug", "logger", map[string]any{"level": "debug"})
//	lokstra_registry.RegisterMiddlewareName("logger-info", "logger", map[string]any{"level": "info"})
func RegisterMiddlewareName(mwName string, mwType string, config map[string]any, opts ...RegisterOption) {
	var deployOpts []deploy.MiddlewareNameOption
	for _, opt := range opts {
		if allowOpt, ok := opt.(*allowOverrideOption); ok {
			deployOpts = append(deployOpts, deploy.WithAllowOverrideForName(allowOpt.allowOverride))
		}
	}
	deploy.Global().RegisterMiddlewareName(mwName, mwType, config, deployOpts...)
}

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
// Supports three factory signatures (auto-wrapped by framework):
//   - func(deps, cfg map[string]any) any - full control (canonical)
//   - func(cfg map[string]any) any       - only config
//   - func() any                          - no params
//
// Both local and remote factories support all three signatures.
//
// Example:
//
//	// Simple factory (no deps, no config)
//	lokstra_registry.RegisterServiceType("user-service",
//	    func() any { return service.NewUserService() },
//	    nil,
//	    deploy.WithResource("user", "users"),
//	)
//
//	// With config
//	lokstra_registry.RegisterServiceType("db-service",
//	    func(cfg map[string]any) any {
//	        return db.NewConnection(cfg["dsn"].(string))
//	    },
//	    nil,
//	)
//
//	// Full signature with deps
//	lokstra_registry.RegisterServiceType("order-service",
//	    func(deps, cfg map[string]any) any {
//	        userSvc := deps["userService"].(*service.Cached[*UserService])
//	        return service.NewOrderService(userSvc.Get())
//	    },
//	    nil,
//	)
func RegisterServiceType(serviceType string, local, remote any, options ...deploy.RegisterServiceTypeOption) {
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

// RegisterLazyService registers a lazy service factory that will be instantiated on first access.
// The factory will be called only once, and the result is cached.
// This allows services to be registered in any order, regardless of dependencies.
//
// Supports two factory signatures (auto-wrapped by framework):
//   - func(cfg map[string]any) any - with config
//   - func() any                    - no params (simplest!)
//
// Dependencies are resolved manually via lokstra_registry.MustGetService() inside factory.
//
// Benefits:
//   - No need to worry about service creation order
//   - Dependencies are auto-resolved on first access
//   - Services only created when needed
//   - Thread-safe singleton pattern
//   - Config per instance (e.g., multiple DB connections)
//
// Example with config:
//
//	lokstra_registry.RegisterLazyService("db-main", func(cfg map[string]any) any {
//	    return db.NewConnection(cfg["dsn"].(string))
//	}, map[string]any{
//	    "dsn": "postgresql://localhost/main",
//	})
//
// Example without params:
//
//	lokstra_registry.RegisterLazyService("user-repo", func() any {
//	    db := lokstra_registry.MustGetService[*DB]("db-main")
//	    return repository.NewUserRepository(db)
//	}, nil)
//
// For explicit dependency injection, use RegisterLazyServiceWithDeps instead.
func RegisterLazyService(name string, factory any, config map[string]any) {
	deploy.Global().RegisterLazyService(name, factory, config)
}

// RegisterLazyServiceWithDeps registers a lazy service with explicit dependency injection.
// The factory will be called only once, and the result is cached.
//
// The deps parameter maps dependency keys to service names for auto-injection:
//
//	deps := map[string]string{
//	    "userService": "user-service",  // key in factory -> service name in registry
//	    "orderRepo": "order-repo",
//	}
//
// Factory signature: func(deps, cfg map[string]any) any
//
// Benefits:
//   - Explicit dependency declaration
//   - Auto-injection by framework
//   - No manual lokstra_registry.MustGetService() calls
//   - Clear dependency graph
//
// Example:
//
//	lokstra_registry.RegisterLazyServiceWithDeps("order-service",
//	    func(deps, cfg map[string]any) any {
//	        // deps already contains resolved services!
//	        userSvc := deps["userService"].(*UserService)
//	        orderRepo := deps["orderRepo"].(*OrderRepository)
//	        maxItems := cfg["max_items"].(int)
//	        return &OrderService{
//	            userService: userSvc,
//	            orderRepo: orderRepo,
//	            maxItems: maxItems,
//	        }
//	    },
//	    map[string]string{
//	        "userService": "user-service",
//	        "orderRepo": "order-repo",
//	    },
//	    map[string]any{"max_items": 5},
//	)
//
// By default, panics if service already registered. Use options to change behavior:
//
//	// Skip if already registered (idempotent)
//	lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
//	    deploy.WithRegistrationMode(deploy.LazyServiceSkip))
//
//	// Override existing registration
//	lokstra_registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
//	    deploy.WithRegistrationMode(deploy.LazyServiceOverride))
func RegisterLazyServiceWithDeps(name string, factory any, deps map[string]string, config map[string]any, opts ...deploy.LazyServiceOption) {
	deploy.Global().RegisterLazyServiceWithDeps(name, factory, deps, config, opts...)
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

// GetLazyService creates a lazy-loading service wrapper.
// The service will be loaded from the global registry only on first access (Get() call).
// This is perfect for dependency injection in handlers and other components.
//
// Example usage:
//
//	// In main.go or handler setup
//	userService := lokstra_registry.GetLazyService[*UserService]("user-service")
//	handler := NewUserHandler(userService)
//
//	// In handler - service loaded only when first accessed
//	func (h *UserHandler) GetUser(ctx *request.Context) error {
//	    users := h.userService.Get().GetAll()  // Lazy loaded here!
//	    return ctx.Api.Ok(users)
//	}
func GetLazyService[T any](serviceName string) *service.Cached[T] {
	return service.LazyLoad[T](serviceName)
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
// Supports both YAML-defined middlewares and factory-based middlewares
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

// ===== SHUTDOWN =====

// Shutdownable is an interface for services that need cleanup on shutdown
type Shutdownable interface {
	Shutdown() error
}

// ShutdownServices gracefully shuts down all services that implement the Shutdownable interface.
// This function iterates through all registered service instances and calls Shutdown() on those
// that implement the Shutdownable interface.
//
// Example service with shutdown:
//
//	type DatabaseService struct {
//	    conn *sql.DB
//	}
//
//	func (s *DatabaseService) Shutdown() error {
//	    return s.conn.Close()
//	}
//
// Usage in main.go:
//
//	defer lokstra_registry.ShutdownServices()
//
// Or with signal handling:
//
//	c := make(chan os.Signal, 1)
//	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
//	go func() {
//	    <-c
//	    lokstra_registry.ShutdownServices()
//	    os.Exit(0)
//	}()
func ShutdownServices() {
	deploy.Global().ShutdownServices()
}
