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
	"reflect"

	"github.com/primadi/lokstra/common/cast"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader/resolver"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/service"
)

// Register path resolver for router package
func init() {
	router.RegisterPathResolver(SimpleResolver)
}

// ===== TYPE ALIASES FOR CLEANER API =====

// ServiceTypeConfig is a structured configuration for service type registration
type ServiceTypeConfig = deploy.ServiceTypeConfig

// RouteConfig defines custom configuration for a specific route
type RouteConfig = deploy.RouteConfig

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
	switch f := factory.(type) {
	case MiddlewareFactory:
		deployFactory = f
	case func(map[string]any) request.HandlerFunc:
		deployFactory = func(cfg map[string]any) any {
			return f(cfg)
		}
	case func(map[string]any) any:
		deployFactory = f
	case func() request.HandlerFunc:
		deployFactory = func(cfg map[string]any) any {
			return f()
		}
	case func() any:
		deployFactory = func(cfg map[string]any) any {
			return f()
		}
	default:
		panic(fmt.Sprintf("invalid middleware factory signature: %T", factory))
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
//
// RegisterRouterServiceType registers a service type with HTTP routing configuration.
// Use this for services that expose HTTP endpoints (annotated with @RouterService).
// For simple infrastructure services (DB, Redis, etc), use RegisterServiceType instead.
//
// Parameters:
//   - serviceType: Unique identifier for this service type
//   - local: Factory for local deployment (same process)
//   - remote: Factory for remote deployment (HTTP client)
//   - config: Optional routing configuration (path prefix, middlewares, route overrides)
//
// Example:
//
//	lokstra_registry.RegisterRouterServiceType("user-service-factory",
//	    application.UserServiceFactory, nil,
//	    &deploy.ServiceTypeConfig{
//	        PathPrefix: "/api/users",
//	        Middlewares: []string{"auth"},
//	    })
func RegisterRouterServiceType(serviceType string, local, remote any, config *deploy.ServiceTypeConfig) {
	deploy.Global().RegisterRouterServiceType(serviceType, local, remote, config)
}

// RegisterServiceType registers a simple service type without HTTP routing.
// Use this for infrastructure services like database pools, Redis clients, metrics, etc.
// For services that expose HTTP endpoints, use RegisterRouterServiceType instead.
//
// Parameters:
//   - serviceType: Unique identifier for this service type
//   - factory: Factory function that creates service instances
//
// Supported factory signatures:
//   - func() any
//   - func(deps map[string]any) any
//   - func(cfg map[string]any) any
//   - func(deps, cfg map[string]any) any
//
// Example:
//
//	lokstra_registry.RegisterServiceType("db-pool-factory",
//	    func(cfg map[string]any) any {
//	        return db.NewPool(cfg["dsn"].(string))
//	    })
func RegisterServiceType(serviceType string, factory any) {
	deploy.Global().RegisterServiceType(serviceType, factory)
}

// GetServiceFactory returns the service factory for a service type
// isLocal: true for local factory, false for remote factory
func GetServiceFactory(serviceType string, isLocal bool) deploy.ServiceFactory {
	return deploy.Global().GetServiceFactory(serviceType, isLocal)
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

// SetConfig sets a runtime configuration value.
// Useful for:
//   - Runtime detection results (mode, environment)
//   - Computed values from expensive calculations
//   - Dynamic service discovery
//   - Feature flags
//
// Example:
//
//	// Store runtime mode
//	lokstra_registry.SetConfig("runtime.mode", "dev")
//
//	// Store computed license key
//	licenseKey := generateComplexLicenseKey() // expensive
//	lokstra_registry.SetConfig("computed.license-key", licenseKey)
//
//	// Later retrieve
//	mode := lokstra_registry.GetConfig("runtime.mode", "prod")
//	key := lokstra_registry.GetConfig("computed.license-key", "")
func SetConfig(key string, value any) {
	deploy.Global().SetConfig(key, value)
}

// GetConfig retrieves a configuration value with type assertion and default value.
// Supports automatic conversion from map[string]any to struct T.
//
// Usage examples:
//
//	// Simple types (direct type assertion)
//	dsn := GetConfig("global-db.dsn", "")
//	port := GetConfig("server.port", 8080)
//
//	// Map access
//	dbConfig := GetConfig[map[string]any]("global-db", nil)
//
//	// Struct binding (automatic conversion from map)
//	type DBConfig struct {
//	    DSN    string `json:"dsn"`
//	    Schema string `json:"schema"`
//	}
//	dbConfig := GetConfig[DBConfig]("global-db", DBConfig{})
func GetConfig[T any](name string, defaultValue T) T {
	value, ok := deploy.Global().GetConfig(name)
	if !ok {
		return defaultValue
	}

	// Direct type match - fastest path
	if typed, ok := value.(T); ok {
		return typed
	}

	// Try converting map[string]any to struct T
	if mapValue, ok := value.(map[string]any); ok {
		var zero T
		targetType := reflect.TypeOf(zero)

		// Check if T is a struct (not pointer)
		if targetType.Kind() == reflect.Struct {
			// Create a pointer to T for ToStruct
			ptr := reflect.New(targetType).Interface()
			if err := cast.ToStruct(mapValue, ptr, false); err == nil {
				// Return dereferenced value
				return reflect.ValueOf(ptr).Elem().Interface().(T)
			}
		}

		// Check if T is a pointer to struct
		if targetType.Kind() == reflect.Pointer && targetType.Elem().Kind() == reflect.Struct {
			// Create a new instance of T
			ptr := reflect.New(targetType.Elem()).Interface()
			if err := cast.ToStruct(mapValue, ptr, false); err == nil {
				return ptr.(T)
			}
		}
	}

	return defaultValue
}

// SimpleResolver resolves variables in the format ${key} or ${key:default}
// by looking up values from config registry via GetConfig().
//
// This is a wrapper function that delegates to deploy.Global().SimpleResolver().
//
// This is designed for annotations where config values are managed centrally:
//   - Annotation prefix: prefix="${auth-prefix}"
//   - Route paths: @Route "GET ${api-version}/users/{id}"
//
// The config registry values can use any provider (ENV, AWS, VAULT, etc.)
// as defined in config.yaml, providing full flexibility while keeping
// annotation syntax simple.
//
// Examples:
//
//	SimpleResolver("${auth-prefix}")              -> GetConfig("auth-prefix", "")
//	SimpleResolver("${auth-prefix:/api/auth}")    -> GetConfig("auth-prefix", "/api/auth")
//	SimpleResolver("/api/${version:v1}/users")    -> "/api/" + GetConfig("version", "v1") + "/users"
//	SimpleResolver("GET ${api-version}/users/{id}") -> "GET /api/users/{id}" (if api-version=/api)
//
// YAML Config Example:
//
//	configs:
//	  - name: auth-prefix
//	    value: ${AUTH_PREFIX:/api/auth}  # ENV variable with default
//	  - name: api-version
//	    value: /api                      # Static value
//	  - name: db-host
//	    value: ${@VAULT:database/host:localhost}  # Vault with default
//
// Note: This function requires lokstra_registry to be initialized with loaded config.
func SimpleResolver(input string) string {
	return deploy.Global().SimpleResolver(input)
}

// ===== PROVIDER REGISTRY (for custom config resolvers) =====

// Provider is an alias to loader.Provider for easier access
// This allows registering custom config value providers (AWS, Vault, K8s, etc.)
type Provider = resolver.Provider

// RegisterProvider registers a custom provider for config value resolution
// Providers can resolve values from various sources (AWS Secrets, Vault, K8s, etc.)
//
// Examples:
//   - RegisterProvider(&AWSSecretProvider{}) -> resolve ${@aws-secret:key}
//   - RegisterProvider(&VaultProvider{}) -> resolve ${@vault:path}
//   - RegisterProvider(&K8sConfigMapProvider{}) -> resolve ${@k8s:configmap/key}
//
// See core/deploy/loader/PROVIDER-REGISTRY.md for complete examples
func RegisterProvider(p Provider) {
	resolver.RegisterProvider(p)
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

// ===== DEPLOYMENT TOPOLOGY REGISTRATION =====

// RegisterDeployment registers a deployment topology from code
// This is the code-equivalent of YAML deployment definition
//
// Example:
//
//	lokstra_registry.RegisterDeployment("microservice", &lokstra_registry.DeploymentConfig{
//	    Servers: map[string]*lokstra_registry.ServerConfig{
//	        "user-server": {
//	            BaseURL: "http://localhost:3001",
//	            Addr: ":3001",
//	            PublishedServices: []string{"user-service"},
//	        },
//	        "order-server": {
//	            BaseURL: "http://localhost:3002",
//	            Addr: ":3002",
//	            PublishedServices: []string{"order-service"},
//	        },
//	    },
//	})
func RegisterDeployment(name string, config *DeploymentConfig) error {
	return deploy.Global().RegisterDeployment(name, config)
}
