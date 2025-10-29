package deploy

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/primadi/lokstra/core/deploy/resolver"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/internal/registry"
)

// GlobalRegistry stores all global definitions (configs, middlewares, services, etc.)
// These are shared across all deployments
type GlobalRegistry struct {
	mu sync.RWMutex

	// Resolver for config values
	resolver *resolver.Registry

	// Factories (code-defined)
	serviceFactories    map[string]*ServiceFactoryEntry
	middlewareFactories map[string]MiddlewareFactory

	// Middleware factory entries (for old_registry pattern compatibility)
	middlewareEntries sync.Map // map[string]*MiddlewareEntry

	// Runtime instances (registered routers, services, middlewares)
	routerInstances     sync.Map // map[string]router.Router
	serviceInstances    sync.Map // map[string]any
	middlewareInstances sync.Map // map[string]request.HandlerFunc

	// Lazy service factories (for on-demand creation)
	lazyServiceFactories sync.Map // map[string]*LazyServiceEntry
	lazyServiceOnce      sync.Map // map[string]*sync.Once

	// Deferred service definitions (string factory type - instantiated on access)
	serviceDefs sync.Map // map[string]*deferredServiceDef

	// Definitions (YAML or code-defined)
	configs map[string]*schema.ConfigDef
	routers map[string]*schema.RouterDef
	// Note: routerOverrides removed - overrides are now inline in RouterDef
	// Note: middlewares map removed - use middlewareEntries sync.Map (unified API)
	// Note: services map removed - use serviceDefs sync.Map (unified API - Opsi 2)

	// Resolved config values (after resolver processing)
	resolvedConfigs map[string]any

	// Topology storage (2-Layer Architecture)
	// Single source of truth for runtime topology
	deploymentTopologies sync.Map // map[deploymentName]*DeploymentTopology
	serverTopologies     sync.Map // map[compositeKey]*ServerTopology (key: "deployment.server")

	// Current server context (for runtime service resolution)
	currentCompositeKey string // "deployment.server" - set by SetCurrentServer
}

// deferredServiceDef holds service definition for deferred instantiation
// Used when RegisterLazyService is called with string factory type
// Instantiation is deferred until first access, allowing auto-detect LOCAL/REMOTE
type deferredServiceDef struct {
	Name        string
	FactoryType string
	DependsOn   []string
	Config      map[string]any
}

// ServiceFactoryEntry holds local and remote factory functions plus metadata
type ServiceFactoryEntry struct {
	Local    ServiceFactory
	Remote   ServiceFactory
	Metadata *ServiceMetadata // Optional metadata for auto-router generation
}

// LazyServiceEntry holds a lazy service factory and its config
type LazyServiceEntry struct {
	Factory func(deps, config map[string]any) any
	Config  map[string]any
	Deps    map[string]string // Dependency mapping: key in factory -> service name in registry
}

// ServiceMetadata holds metadata for service auto-generation
// ServiceMetadata holds metadata for a service type registration
// Can be populated from ServiceTypeConfig or legacy functional options
type ServiceMetadata struct {
	Resource        string                   // Singular resource name (e.g., "user")
	ResourcePlural  string                   // Plural resource name (e.g., "users")
	Convention      string                   // Convention type (e.g., "rest", "rpc")
	PathPrefix      string                   // Path prefix for all routes
	MiddlewareNames []string                 // Router-level middleware names
	HiddenMethods   []string                 // Methods to hide from router
	RouteOverrides  map[string]RouteMetadata // Method name -> full route metadata (NEW: supports route-level middlewares)
}

// RouteMetadata holds full metadata for a custom route
// This supports both path override AND route-level middlewares
type RouteMetadata struct {
	Method      string   // HTTP method (e.g., "POST", "GET") - empty means auto-detect
	Path        string   // Custom path (e.g., "/auth/login")
	Middlewares []string // Route-specific middleware names
}

// MiddlewareEntry holds middleware type and config for factory pattern
type MiddlewareEntry struct {
	Type   string         // Middleware type (factory name)
	Config map[string]any // Configuration for the middleware
}

// ===== TOPOLOGY STRUCTS (2-Layer Architecture) =====
// These replace the complex Deployment/Server/App structs with simple data holders
// All topology data is stored in GlobalRegistry (single source of truth)

// DeploymentTopology holds deployment-level configuration
type DeploymentTopology struct {
	Name            string
	ConfigOverrides map[string]any
	Servers         map[string]*ServerTopology
}

// ServerTopology holds server-level topology
// Services and RemoteServices are at SERVER level (shared across all apps)
type ServerTopology struct {
	Name           string
	DeploymentName string
	BaseURL        string
	Services       []string          // Service names (server-level, shared)
	RemoteServices map[string]string // serviceName -> remoteBaseURL (empty string if local)
	Apps           []*AppTopology
}

// AppTopology holds app-level topology
// Apps only have addr and routers (services are at server level)
type AppTopology struct {
	Addr    string
	Routers []string // Router names
}

// ServiceFactory creates a service instance
// deps: dependencies resolved as map[paramName]*service.Cached[any]
// config: configuration for this service instance
// Dependencies are lazy-loaded - call .Get() to resolve
type ServiceFactory func(deps map[string]any, config map[string]any) any

// MiddlewareFactory creates a middleware instance
type MiddlewareFactory func(config map[string]any) any

var globalRegistry = NewGlobalRegistry()

func init() {
	// Register to internal registry for access by core packages
	registry.SetGlobal(globalRegistry)

	// Set middleware resolver for router package (to avoid import cycle)
	router.MiddlewareResolver = func(name string) request.HandlerFunc {
		return globalRegistry.CreateMiddleware(name)
	}
}

// Global returns the singleton global registry
func Global() *GlobalRegistry {
	return globalRegistry
}

// NewGlobalRegistry creates a new global registry
func NewGlobalRegistry() *GlobalRegistry {
	return &GlobalRegistry{
		resolver:            resolver.NewRegistry(),
		serviceFactories:    make(map[string]*ServiceFactoryEntry),
		middlewareFactories: make(map[string]MiddlewareFactory),
		configs:             make(map[string]*schema.ConfigDef),
		routers:             make(map[string]*schema.RouterDef),
		resolvedConfigs:     make(map[string]any),
		// Topology maps, serviceDefs, and middlewareEntries use sync.Map, no initialization needed
	}
}

// ===== FACTORY REGISTRATION (CODE) =====

// RegisterServiceType registers a service factory with configuration
// Supports two patterns:
//
//  1. Struct-based (recommended):
//     RegisterServiceType(type, local, remote, &ServiceTypeConfig{...})
//
//  2. Functional options (legacy):
//     RegisterServiceType(type, local, remote, WithResource(...), WithConvention(...))
//
// Factory signatures (auto-wrapped by framework):
//   - func(deps, cfg map[string]any) any - full control (canonical)
//   - func(cfg map[string]any) any       - only config
//   - func() any                          - no params
//
// Both local and remote factories support all three signatures.
func (g *GlobalRegistry) RegisterServiceType(serviceType string, local, remote any,
	configOrOptions ...any) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.serviceFactories[serviceType]; exists {
		panic(fmt.Sprintf("service type %s already registered", serviceType))
	}

	// Build metadata from config or options
	metadata := &ServiceMetadata{
		Convention: "rest", // Default convention
	}

	// Check if first argument is ServiceTypeConfig
	if len(configOrOptions) > 0 {
		if config, ok := configOrOptions[0].(*ServiceTypeConfig); ok {
			// Struct-based configuration (new pattern)
			if config != nil {
				metadata.Resource = config.Resource
				metadata.ResourcePlural = config.ResourcePlural
				metadata.Convention = config.Convention
				metadata.PathPrefix = config.PathPrefix
				metadata.MiddlewareNames = config.Middlewares
				metadata.HiddenMethods = config.Hidden

				// Convert RouteConfig to RouteMetadata
				if len(config.RouteOverrides) > 0 {
					metadata.RouteOverrides = make(map[string]RouteMetadata)
					for methodName, routeConfig := range config.RouteOverrides {
						metadata.RouteOverrides[methodName] = RouteMetadata(routeConfig)
					}
				}
			}
		} else {
			// Functional options (legacy pattern)
			for _, opt := range configOrOptions {
				if optFunc, ok := opt.(RegisterServiceTypeOption); ok {
					optFunc(metadata)
				}
			}
		}
	}

	// Only store metadata if resource name is provided
	var metadataPtr *ServiceMetadata
	if metadata.Resource != "" {
		metadataPtr = metadata
	}

	// Normalize local and remote factories
	var localFactory ServiceFactory
	if local != nil {
		localFactory = normalizeServiceFactory(local, serviceType, "local")
	}

	var remoteFactory ServiceFactory
	if remote != nil {
		remoteFactory = normalizeServiceFactory(remote, serviceType, "remote")
	}

	g.serviceFactories[serviceType] = &ServiceFactoryEntry{
		Local:    localFactory,
		Remote:   remoteFactory,
		Metadata: metadataPtr,
	}
}

// converts any supported factory signature to canonical ServiceFactory
func normalizeServiceFactory(factoryInput any, serviceType, factoryKind string) ServiceFactory {
	factoryType := reflect.TypeOf(factoryInput)

	// Must be a function
	if factoryType.Kind() != reflect.Func {
		panic(fmt.Sprintf("invalid %s factory for service type %s: must be a function", factoryKind, serviceType))
	}

	// Check number of parameters and return values
	numIn := factoryType.NumIn()
	numOut := factoryType.NumOut()

	// Must return exactly 1 value
	if numOut != 1 {
		panic(fmt.Sprintf("invalid %s factory signature for service type %s: must return exactly 1 value", factoryKind, serviceType))
	}

	factoryValue := reflect.ValueOf(factoryInput)

	// Match based on number of input parameters
	switch numIn {
	case 0:
		// func() T where T is assignable to any
		// Wrap to: func(deps, cfg map[string]any) any
		return func(_ map[string]any, _ map[string]any) any {
			results := factoryValue.Call([]reflect.Value{})
			return results[0].Interface()
		}

	case 1:
		// func(cfg map[string]any) T where T is assignable to any
		// Verify first param is map[string]any
		param0 := factoryType.In(0)
		mapStringAnyType := reflect.TypeOf(map[string]any{})
		if param0 != mapStringAnyType {
			panic(fmt.Sprintf("invalid %s factory signature for service type %s: single parameter must be map[string]any, got %s", factoryKind, serviceType, param0))
		}

		// Wrap to: func(deps, cfg map[string]any) any
		return func(_ map[string]any, cfg map[string]any) any {
			results := factoryValue.Call([]reflect.Value{reflect.ValueOf(cfg)})
			return results[0].Interface()
		}

	case 2:
		// func(deps, cfg map[string]any) T where T is assignable to any
		// Verify both params are map[string]any
		param0 := factoryType.In(0)
		param1 := factoryType.In(1)
		mapStringAnyType := reflect.TypeOf(map[string]any{})

		if param0 != mapStringAnyType || param1 != mapStringAnyType {
			panic(fmt.Sprintf("invalid %s factory signature for service type %s: both parameters must be map[string]any, got (%s, %s)", factoryKind, serviceType, param0, param1))
		}

		// Already canonical form - wrap to ensure return type is any
		return func(deps, cfg map[string]any) any {
			results := factoryValue.Call([]reflect.Value{reflect.ValueOf(deps), reflect.ValueOf(cfg)})
			return results[0].Interface()
		}

	default:
		panic(fmt.Sprintf("invalid %s factory signature for service type %s: must have 0, 1, or 2 parameters, got %d", factoryKind, serviceType, numIn))
	}
}

// RegisterMiddlewareType registers a middleware factory
// Supports optional AllowOverride option
func (g *GlobalRegistry) RegisterMiddlewareType(middlewareType string, factory MiddlewareFactory, opts ...MiddlewareTypeOption) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var options middlewareTypeOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.allowOverride {
		if _, exists := g.middlewareFactories[middlewareType]; exists {
			panic(fmt.Sprintf("middleware type %s already registered", middlewareType))
		}
	}

	g.middlewareFactories[middlewareType] = factory
}

// RegisterMiddlewareName registers a middleware entry by name, associating it with a type and config.
// This allows creating multiple middleware instances from the same factory with different configurations.
//
// Example:
//
//	g.RegisterMiddlewareType("logger", loggerFactory)
//	g.RegisterMiddlewareName("logger-debug", "logger", map[string]any{"level": "debug"})
//	g.RegisterMiddlewareName("logger-info", "logger", map[string]any{"level": "info"})
func (g *GlobalRegistry) RegisterMiddlewareName(name, middlewareType string, config map[string]any, opts ...MiddlewareNameOption) {
	var options middlewareNameOptions
	for _, opt := range opts {
		opt(&options)
	}

	if !options.allowOverride {
		if _, exists := g.middlewareEntries.Load(name); exists {
			panic(fmt.Sprintf("middleware name %s already registered", name))
		}
	}

	g.middlewareEntries.Store(name, &MiddlewareEntry{
		Type:   middlewareType,
		Config: config,
	})
}

// RegisterResolver registers a custom config resolver
func (g *GlobalRegistry) RegisterResolver(r resolver.Resolver) {
	g.resolver.Register(r)
}

// ===== DEFINITION REGISTRATION (YAML OR CODE) =====

// DefineConfig defines a configuration value
func (g *GlobalRegistry) DefineConfig(def *schema.ConfigDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.configs[def.Name]; exists {
		panic(fmt.Sprintf("config %s already defined", def.Name))
	}

	g.configs[def.Name] = def
}

// DefineRouter defines a router
func (g *GlobalRegistry) DefineRouter(name string, def *schema.RouterDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.routers[name]; exists {
		panic(fmt.Sprintf("router %s already defined", name))
	}

	g.routers[name] = def
}

// ===== GETTERS =====

// GetServiceFactory returns the service factory for a service type
// isLocal: true for local factory, false for remote factory
func (g *GlobalRegistry) GetServiceFactory(serviceType string, isLocal bool) ServiceFactory {
	g.mu.RLock()
	defer g.mu.RUnlock()

	entry, ok := g.serviceFactories[serviceType]
	if !ok {
		return nil
	}

	if isLocal {
		return entry.Local
	}
	return entry.Remote
}

// GetServiceMetadata returns the service metadata for a service type
func (g *GlobalRegistry) GetServiceMetadata(serviceType string) *ServiceMetadata {
	g.mu.RLock()
	defer g.mu.RUnlock()

	entry, ok := g.serviceFactories[serviceType]
	if !ok {
		return nil
	}

	return entry.Metadata
}

// GetMiddlewareFactory returns the middleware factory
func (g *GlobalRegistry) GetMiddlewareFactory(middlewareType string) MiddlewareFactory {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.middlewareFactories[middlewareType]
}

// GetConfig returns a config definition
func (g *GlobalRegistry) GetConfig(name string) *schema.ConfigDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.configs[name]
}

// GetRouterDef returns a router definition
func (g *GlobalRegistry) GetRouterDef(name string) *schema.RouterDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.routers[name]
}

// ===== CONFIG RESOLUTION =====

// ResolveConfigs resolves all config values using the resolver
// This performs 2-step resolution:
//  1. Resolve all ${...} except ${@cfg:...}
//  2. Resolve ${@cfg:...} using step 1 results
func (g *GlobalRegistry) ResolveConfigs() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Build initial resolved configs map (before resolution)
	tempConfigs := make(map[string]any)
	for name, def := range g.configs {
		tempConfigs[name] = def.Value
	}

	// Resolve each config value
	for name, def := range g.configs {
		// Convert value to string if needed
		var valueStr string
		switch v := def.Value.(type) {
		case string:
			valueStr = v
		default:
			// Non-string values are used as-is
			g.resolvedConfigs[name] = v
			continue
		}

		// Resolve the value
		resolved, err := g.resolver.ResolveValue(valueStr, tempConfigs)
		if err != nil {
			return fmt.Errorf("failed to resolve config %s: %w", name, err)
		}

		g.resolvedConfigs[name] = resolved

		// Update temp map for subsequent @cfg references
		tempConfigs[name] = resolved
	}

	return nil
}

// GetResolvedConfig returns a resolved config value
func (g *GlobalRegistry) GetResolvedConfig(name string) (any, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	value, ok := g.resolvedConfigs[name]
	return value, ok
}

// ResolveConfigValue resolves a single config value (helper for service configs)
func (g *GlobalRegistry) ResolveConfigValue(value any) (any, error) {
	// Convert to string if needed
	valueStr, ok := value.(string)
	if !ok {
		// Non-string values are used as-is
		return value, nil
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.resolver.ResolveValue(valueStr, g.resolvedConfigs)
}

// ===== RUNTIME INSTANCE REGISTRATION =====

// RegisterRouter registers a router instance
func (g *GlobalRegistry) RegisterRouter(name string, r router.Router) {
	if _, exists := g.routerInstances.Load(name); exists {
		panic(fmt.Sprintf("router %s already registered", name))
	}
	g.routerInstances.Store(name, r)
}

// GetRouter retrieves a router instance by name
func (g *GlobalRegistry) GetRouter(name string) router.Router {
	if v, ok := g.routerInstances.Load(name); ok {
		return v.(router.Router)
	}
	return nil
}

// GetAllRouters returns all registered routers
func (g *GlobalRegistry) GetAllRouters() map[string]router.Router {
	result := make(map[string]router.Router)
	g.routerInstances.Range(func(key, value any) bool {
		result[key.(string)] = value.(router.Router)
		return true
	})
	return result
}

// RegisterService registers a service instance
func (g *GlobalRegistry) RegisterService(name string, service any) {
	if _, exists := g.serviceInstances.Load(name); exists {
		panic(fmt.Sprintf("service %s already registered", name))
	}
	g.serviceInstances.Store(name, service)
}

// RegisterLazyService registers a lazy service factory that will be instantiated on first access.
// The factory will be called only once, and the result is cached.
// This allows services to be registered in any order, regardless of dependencies.
//
// Supports two factory signatures:
//   - func(config map[string]any) any - for services that need config
//   - func() any                       - for services without config (simpler!)
//
// The framework auto-wraps the simpler signature for you.
//
// Example with config:
//
//	// Multiple DB instances with different DSN
//	lokstra_registry.RegisterLazyService("db-main", func(cfg map[string]any) any {
//	    return NewDB(cfg["dsn"].(string))
//	}, map[string]any{"dsn": "main-dsn"})
//
// RegisterLazyService registers a lazy service factory that will be instantiated on first access.
// The factory will be called only once, and the result is cached.
// This allows services to be registered in any order, regardless of dependencies.
//
// Supports three factory signatures (auto-wrapped by framework):
//   - func(cfg map[string]any) any - with config
//
// RegisterLazyService registers a lazy service with optional dependencies.
//
// The factory parameter can be:
//   - A string: References a pre-registered factory type (via RegisterServiceType)
//   - A function: Inline factory function
//
// When using string factory type:
//   - Framework auto-wraps dependencies as service.Cached
//   - Supports auto-router generation (if factory has metadata)
//   - Equivalent to YAML service-definitions
//   - Instantiation is deferred until first access
//   - Auto-detects LOCAL vs REMOTE based on server topology
//
// When using inline function:
//   - No metadata, no auto-router
//   - Manual dependency handling required
//   - Suitable for simple services or prototyping
//   - Instantiated immediately (no auto-detect)
//
// Example with string factory type (YAML equivalent):
//
//	lokstra_registry.RegisterLazyService("user-service",
//	    "user-service-factory",  // String factory type
//	    map[string]any{
//	        "depends-on": []string{"user-repository"},
//	        "max-users": 1000,  // Additional config
//	    })
//
// Example with inline function:
//
//	lokstra_registry.RegisterLazyService("cache",
//	    func(deps, cfg map[string]any) any {
//	        return redis.NewClient(&redis.Options{
//	            Addr: cfg["addr"].(string),
//	        })
//	    },
//	    map[string]any{"addr": "localhost:6379"})
func (g *GlobalRegistry) RegisterLazyService(name string, factory any, config map[string]any) {
	// Type detection: string factory type name vs inline function
	if factoryTypeName, ok := factory.(string); ok {
		// String factory type - store definition for deferred instantiation
		g.registerDeferredService(name, factoryTypeName, config)
		return
	}

	// Inline function - immediate registration (no auto-detect)
	g.RegisterLazyServiceWithDeps(name, factory, nil, config)
}

// registerDeferredService stores a service definition using a factory type name.
// The service will be instantiated on first access with auto-detect LOCAL/REMOTE
// based on deployment topology.
func (g *GlobalRegistry) registerDeferredService(name, factoryType string, config map[string]any) {
	// Extract depends-on from config
	var dependsOn []string
	if depsRaw, ok := config["depends-on"]; ok {
		switch deps := depsRaw.(type) {
		case []string:
			dependsOn = deps
		case []interface{}:
			// Handle YAML unmarshaling []interface{}
			dependsOn = make([]string, len(deps))
			for i, d := range deps {
				dependsOn[i] = d.(string)
			}
		}
	}

	// Store deferred definition
	def := &deferredServiceDef{
		Name:        name,
		FactoryType: factoryType,
		DependsOn:   dependsOn,
		Config:      config,
	}

	g.serviceDefs.Store(name, def)
}

// LazyServiceRegistrationMode defines how to handle duplicate registrations
type LazyServiceRegistrationMode int

const (
	// LazyServiceError panics if service already registered (default, strict)
	LazyServiceError LazyServiceRegistrationMode = iota
	// LazyServiceSkip silently skips if service already registered (idempotent)
	LazyServiceSkip
	// LazyServiceOverride replaces existing registration (force update)
	LazyServiceOverride
)

// LazyServiceOption configures lazy service registration behavior
type LazyServiceOption func(*lazyServiceOptions)

type lazyServiceOptions struct {
	mode LazyServiceRegistrationMode
}

// WithRegistrationMode sets the duplicate registration handling mode
func WithRegistrationMode(mode LazyServiceRegistrationMode) LazyServiceOption {
	return func(opts *lazyServiceOptions) {
		opts.mode = mode
	}
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
// By default, panics if service already registered. Use options to change behavior:
//
//	// Skip if already registered (idempotent)
//	registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
//	    deploy.WithRegistrationMode(deploy.LazyServiceSkip))
//
//	// Override existing registration
//	registry.RegisterLazyServiceWithDeps(name, factory, deps, cfg,
//	    deploy.WithRegistrationMode(deploy.LazyServiceOverride))
//
// Example:
//
//	lokstra_registry.RegisterLazyServiceWithDeps("order-service",
//	    func(deps, cfg map[string]any) any {
//	        // deps already contains resolved services!
//	        userSvc := deps["userService"].(*UserService)
//	        orderRepo := deps["orderRepo"].(*OrderRepository)
//	        maxItems := cfg["max_items"].(int)
//	        return &OrderService{userService: userSvc, orderRepo: orderRepo, maxItems: maxItems}
//	    },
//	    map[string]string{
//	        "userService": "user-service",
//	        "orderRepo": "order-repo",
//	    },
//	    map[string]any{"max_items": 5},
//	)
func (g *GlobalRegistry) RegisterLazyServiceWithDeps(name string, factory any, deps map[string]string, config map[string]any, opts ...LazyServiceOption) {
	// Parse options
	options := &lazyServiceOptions{
		mode: LazyServiceError, // Default: strict error on duplicate
	}
	for _, opt := range opts {
		opt(options)
	}

	// Check eager registry
	if _, exists := g.serviceInstances.Load(name); exists {
		switch options.mode {
		case LazyServiceSkip:
			return // Silently skip
		case LazyServiceOverride:
			// Remove from eager registry to allow lazy override
			g.serviceInstances.Delete(name)
		case LazyServiceError:
			panic(fmt.Sprintf("service %s already registered as eager service", name))
		}
	}

	// Check lazy registry
	if _, exists := g.lazyServiceFactories.Load(name); exists {
		switch options.mode {
		case LazyServiceSkip:
			return // Silently skip
		case LazyServiceOverride:
			// Will be overridden below
		case LazyServiceError:
			panic(fmt.Sprintf("lazy service %s already registered", name))
		}
	}

	// Normalize factory using reflection-based normalizer
	normFactory := normalizeServiceFactory(factory, name, "lazy service")

	entry := &LazyServiceEntry{
		Factory: normFactory,
		Config:  config,
		Deps:    deps, // Store dependency mapping
	}

	g.lazyServiceFactories.Store(name, entry)
	g.lazyServiceOnce.Store(name, &sync.Once{})
}

// GetServiceAny retrieves a service instance by name as any
// If not found in eager registry, checks lazy registry and instantiates
// If still not found, checks service-definitions and auto-creates lazy service
func (g *GlobalRegistry) GetServiceAny(name string) (any, bool) {
	// Check eager registry first
	if svc, ok := g.serviceInstances.Load(name); ok {
		return svc, true
	}

	// Check lazy registry and create if needed
	onceAny, hasOnce := g.lazyServiceOnce.Load(name)
	if !hasOnce {
		// Not in lazy registry - check if in deferred service definitions
		if defAny, exists := g.serviceDefs.Load(name); exists {
			deferredDef := defAny.(*deferredServiceDef)

			// Convert deferred definition to schema.ServiceDef for auto-registration
			serviceDef := &schema.ServiceDef{
				Name:      deferredDef.Name,
				Type:      deferredDef.FactoryType,
				DependsOn: deferredDef.DependsOn,
				Config:    deferredDef.Config,
			}

			// Auto-create lazy service from definition
			g.autoRegisterLazyService(name, serviceDef)

			// Now try again
			onceAny, hasOnce = g.lazyServiceOnce.Load(name)
			if !hasOnce {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	once := onceAny.(*sync.Once)

	// Create instance once and cache it
	// IMPORTANT: Load factory inside once.Do to avoid race condition!
	once.Do(func() {
		entryAny, ok := g.lazyServiceFactories.Load(name)
		if !ok {
			// Should not happen, but handle gracefully
			return
		}

		entry := entryAny.(*LazyServiceEntry)

		// Resolve dependencies if specified
		var resolvedDeps map[string]any
		if len(entry.Deps) > 0 {
			resolvedDeps = make(map[string]any, len(entry.Deps))
			for key, serviceName := range entry.Deps {
				// Recursively resolve dependency
				depSvc, ok := g.GetServiceAny(serviceName)
				if !ok {
					panic(fmt.Sprintf("lazy service %s: dependency %s (service %s) not found", name, key, serviceName))
				}
				resolvedDeps[key] = depSvc
			}
		}

		// Call factory with resolved deps or nil
		instance := entry.Factory(resolvedDeps, entry.Config)
		g.serviceInstances.Store(name, instance)
	})

	// Return cached instance
	svc, ok := g.serviceInstances.Load(name)
	return svc, ok
}

// HasLazyService checks if a service is registered in the lazy service registry
// or defined in the deferred service definitions (from YAML or code).
func (g *GlobalRegistry) HasLazyService(name string) bool {
	// Check if already instantiated in lazy registry
	if _, ok := g.lazyServiceFactories.Load(name); ok {
		return true
	}

	// Check if defined but not yet instantiated
	if _, ok := g.serviceDefs.Load(name); ok {
		return true
	}

	return false
}

// GetDeferredServiceDef retrieves a deferred service definition by name.
// Returns the definition if found in serviceDefs, or nil if not found.
// This is primarily used by wrapper functions that need access to service metadata.
func (g *GlobalRegistry) GetDeferredServiceDef(name string) *schema.ServiceDef {
	if defAny, ok := g.serviceDefs.Load(name); ok {
		deferredDef := defAny.(*deferredServiceDef)
		return &schema.ServiceDef{
			Name:      deferredDef.Name,
			Type:      deferredDef.FactoryType,
			DependsOn: deferredDef.DependsOn,
			Config:    deferredDef.Config,
		}
	}
	return nil
}

// autoRegisterLazyService auto-registers a service from service-definitions as a lazy service
// This enables zero-config pattern - services are created on-demand from YAML definitions
// Logic: Check if published on another server → REMOTE, else → LOCAL from service-definitions
func (g *GlobalRegistry) autoRegisterLazyService(name string, def *schema.ServiceDef) {
	// Get current deployment context
	currentKey := g.GetCurrentCompositeKey()
	if currentKey == "" {
		// No current context - default to LOCAL
		g.autoRegisterLocalService(name, def)
		return
	}

	// Get current server topology
	currentServerTopo, ok := g.GetServerTopology(currentKey)
	if !ok {
		// No topology found - default to LOCAL
		g.autoRegisterLocalService(name, def)
		return
	}

	// Check if service is published on another server (REMOTE)
	remoteBaseURL, isRemote := currentServerTopo.RemoteServices[name]
	if isRemote {
		// Register as REMOTE service (HTTP proxy)
		g.AutoRegisterRemoteService(name, def, remoteBaseURL)
		return
	}

	// Not remote - register as LOCAL
	g.autoRegisterLocalService(name, def)
}

// autoRegisterLocalService registers a service as LOCAL (from factory)
func (g *GlobalRegistry) autoRegisterLocalService(name string, def *schema.ServiceDef) {
	// Get factory
	factory := g.GetServiceFactory(def.Type, true) // true = local factory
	if factory == nil {
		panic(fmt.Sprintf("service factory '%s' not registered for service '%s'", def.Type, name))
	}

	// Parse dependencies from DependsOn field
	deps := make(map[string]string)
	if len(def.DependsOn) > 0 {
		for _, depStr := range def.DependsOn {
			// Format: "paramName:serviceName" or just "serviceName"
			parts := strings.Split(depStr, ":")
			if len(parts) == 2 {
				paramName := parts[0]
				serviceName := parts[1]
				deps[paramName] = serviceName
			} else {
				// No explicit param name - use service name as key
				deps[depStr] = depStr
			}
		}
	}

	// Register as lazy service with wrapper factory
	// Factory expects service.Cached for dependencies, so we wrap resolved deps
	g.RegisterLazyServiceWithDeps(name, func(resolvedDeps, cfg map[string]any) any {
		// Wrap resolved dependencies as service.Cached
		// This allows factories to use service.Cast[T](deps["key"])
		lazyDeps := make(map[string]any)
		for key, depSvc := range resolvedDeps {
			depSvcCopy := depSvc // Capture for closure
			lazyDeps[key] = service.LazyLoadWith(func() any { return depSvcCopy })
		}

		// Call original factory
		return factory(lazyDeps, cfg)
	}, deps, def.Config)
}

// AutoRegisterRemoteService registers a service as REMOTE (HTTP proxy)
func (g *GlobalRegistry) AutoRegisterRemoteService(name string, def *schema.ServiceDef, remoteBaseURL string) {
	// Get remote factory
	factory := g.GetServiceFactory(def.Type, false) // false = remote factory
	if factory == nil {
		panic(fmt.Sprintf("remote service factory '%s' not registered for service '%s'", def.Type, name))
	}

	// Get service metadata for proxy.Service creation
	metadata := g.GetServiceMetadata(def.Type)

	// Create proxy.Service for HTTP calls
	var proxyService *proxy.Service
	if metadata != nil && metadata.Resource != "" {
		// Use metadata from RegisterServiceType
		override := autogen.RouteOverride{
			PathPrefix: metadata.PathPrefix,
			Hidden:     metadata.HiddenMethods,
		}

		// Convert RouteOverrides map to Custom routes
		if len(metadata.RouteOverrides) > 0 {
			override.Custom = make(map[string]autogen.Route)
			for methodName, routeMeta := range metadata.RouteOverrides {
				// RouteMetadata now has Method and Path directly
				override.Custom[methodName] = autogen.Route{
					Method:      routeMeta.Method,
					Path:        routeMeta.Path,
					Middlewares: convertMiddlewareNames(g, routeMeta.Middlewares), // Convert middleware names to instances
				}
			}
		}

		// Convert router-level middlewares
		if len(metadata.MiddlewareNames) > 0 {
			override.Middlewares = convertMiddlewareNames(g, metadata.MiddlewareNames)
		}

		proxyService = proxy.NewService(
			remoteBaseURL,
			autogen.ConversionRule{
				Convention:     convention.ConventionType(metadata.Convention),
				Resource:       metadata.Resource,
				ResourcePlural: metadata.ResourcePlural,
			},
			override,
		)
	} else {
		// Fallback: auto-generate from service name
		resourceName := strings.TrimSuffix(name, "-service")
		resourcePlural := resourceName + "s" // Simple pluralization
		proxyService = proxy.NewService(
			remoteBaseURL,
			autogen.ConversionRule{
				Convention:     convention.REST,
				Resource:       resourceName,
				ResourcePlural: resourcePlural,
			},
			autogen.RouteOverride{},
		)
	}

	// Build config with proxy.Service
	remoteConfig := make(map[string]any)
	// Copy service-level config if exists
	for k, v := range def.Config {
		remoteConfig[k] = v
	}
	// Add proxy.Service for remote calls
	remoteConfig["remote"] = proxyService

	// Register as lazy service (remote services have no dependencies)
	g.RegisterLazyServiceWithDeps(name, func(_, cfg map[string]any) any {
		return factory(nil, cfg)
	}, nil, remoteConfig, WithRegistrationMode(LazyServiceSkip))
}

// RegisterMiddleware registers a middleware instance by name
func (g *GlobalRegistry) RegisterMiddleware(name string, mw request.HandlerFunc) {
	if _, exists := g.middlewareInstances.Load(name); exists {
		panic(fmt.Sprintf("middleware %s already registered", name))
	}
	g.middlewareInstances.Store(name, mw)
}

// GetMiddleware retrieves a middleware instance by name
func (g *GlobalRegistry) GetMiddleware(name string) (request.HandlerFunc, bool) {
	if v, ok := g.middlewareInstances.Load(name); ok {
		return v.(request.HandlerFunc), true
	}
	return nil, false
}

// CreateMiddleware creates a middleware instance from definition
func (g *GlobalRegistry) CreateMiddleware(name string) request.HandlerFunc {
	// First check if already instantiated
	if mw, ok := g.middlewareInstances.Load(name); ok {
		return mw.(request.HandlerFunc)
	}

	// Check if it's registered via RegisterMiddlewareName (factory pattern)
	if entryAny, ok := g.middlewareEntries.Load(name); ok {
		entry := entryAny.(*MiddlewareEntry)
		factory := g.GetMiddlewareFactory(entry.Type)
		if factory != nil {
			mw := factory(entry.Config)
			if handlerFunc, ok := mw.(request.HandlerFunc); ok {
				// Cache it
				g.middlewareInstances.Store(name, handlerFunc)
				return handlerFunc
			}
		}
		return nil
	}

	// Try to create from middleware entry (RegisterMiddlewareName)
	if entryAny, exists := g.middlewareEntries.Load(name); exists {
		entry := entryAny.(*MiddlewareEntry)

		// Get factory
		factory := g.GetMiddlewareFactory(entry.Type)
		if factory == nil {
			return nil
		}

		// Create instance
		mw := factory(entry.Config)
		if handlerFunc, ok := mw.(request.HandlerFunc); ok {
			// Cache it
			g.RegisterMiddleware(name, handlerFunc)
			return handlerFunc
		}
	}

	return nil
}

// ===== TOPOLOGY MANAGEMENT (2-Layer Architecture) =====

// DeploymentConfig is used by RegisterDeployment for code-based topology registration
// This is defined in lokstra_registry package to avoid import cycles
type DeploymentConfig interface {
	GetConfigOverrides() map[string]any
	GetServers() map[string]ServerConfig
}

// ServerConfig interface for deployment registration
type ServerConfig interface {
	GetBaseURL() string
	GetApps() []AppConfig
	GetAddr() string
	GetRouters() []string
	GetPublishedServices() []string
}

// AppConfig interface for deployment registration
type AppConfig interface {
	GetAddr() string
	GetRouters() []string
	GetPublishedServices() []string
}

// RegisterDeployment registers a deployment topology from code
// This is the code-equivalent of YAML deployment definition
// It builds the topology and stores it for runtime use
func (g *GlobalRegistry) RegisterDeployment(deploymentName string, config DeploymentConfig) error {
	// Auto-generate router definitions for published services
	// Collect all published services from all servers
	publishedServicesMap := make(map[string]bool)
	for _, serverConfig := range config.GetServers() {
		// Collect apps (from Apps slice + shorthand fields)
		apps := serverConfig.GetApps()

		// If shorthand fields are set, create an app from them
		if serverConfig.GetAddr() != "" {
			shorthandApp := &shorthandAppConfig{
				addr:              serverConfig.GetAddr(),
				routers:           serverConfig.GetRouters(),
				publishedServices: serverConfig.GetPublishedServices(),
			}
			// Prepend shorthand app
			apps = append([]AppConfig{shorthandApp}, apps...)
		}

		for _, appConfig := range apps {
			for _, serviceName := range appConfig.GetPublishedServices() {
				publishedServicesMap[serviceName] = true
			}
		}
	}

	// Define routers for each published service
	for serviceName := range publishedServicesMap {
		routerName := serviceName + "-router"

		// Check if router already defined manually
		if g.GetRouterDef(routerName) != nil {
			continue // Skip, use existing definition
		}

		// Check if service is registered
		if !g.HasLazyService(serviceName) {
			return fmt.Errorf("published service '%s' not found in service registry", serviceName)
		}

		// Get service definition to find service type
		serviceDef := g.GetDeferredServiceDef(serviceName)
		if serviceDef == nil {
			return fmt.Errorf("published service '%s' definition not found", serviceName)
		}

		// Get service metadata from factory registration
		metadata := g.GetServiceMetadata(serviceDef.Type)

		// Build router definition
		var resourceName, resourcePlural, convention string

		// Use metadata from RegisterServiceType if available
		if metadata != nil && metadata.Resource != "" {
			resourceName = metadata.Resource
			resourcePlural = metadata.ResourcePlural
			convention = metadata.Convention
		} else {
			// Fallback: auto-generate from service name
			resourceName = strings.TrimSuffix(serviceName, "-service")
			resourcePlural = resourceName + "s" // Simple pluralization
			convention = "rest"
		}

		// Define router with metadata
		routerDef := &schema.RouterDef{
			Convention:     convention,
			Resource:       resourceName,
			ResourcePlural: resourcePlural,
		}

		// Add metadata overrides if available
		if metadata != nil {
			routerDef.PathPrefix = metadata.PathPrefix
			routerDef.Middlewares = metadata.MiddlewareNames
			routerDef.Hidden = metadata.HiddenMethods

			// Convert RouteOverrides to Custom routes
			if len(metadata.RouteOverrides) > 0 {
				routerDef.Custom = make([]schema.RouteDef, 0, len(metadata.RouteOverrides))
				for methodName, routeMeta := range metadata.RouteOverrides {
					routerDef.Custom = append(routerDef.Custom, schema.RouteDef{
						Name:        methodName,
						Method:      routeMeta.Method,
						Path:        routeMeta.Path,
						Middlewares: routeMeta.Middlewares,
					})
				}
			}
		}

		g.DefineRouter(routerName, routerDef)
	}

	// Build service location registry (service-name → base-url)
	// This maps published services to their server URLs for remote service resolution
	serviceLocations := make(map[string]string)

	for _, serverConfig := range config.GetServers() {
		// Collect apps (from Apps slice + shorthand fields)
		apps := serverConfig.GetApps()

		// If shorthand fields are set, create an app from them
		if serverConfig.GetAddr() != "" {
			shorthandApp := &shorthandAppConfig{
				addr:              serverConfig.GetAddr(),
				routers:           serverConfig.GetRouters(),
				publishedServices: serverConfig.GetPublishedServices(),
			}
			// Prepend shorthand app
			apps = append([]AppConfig{shorthandApp}, apps...)
		}

		for _, appConfig := range apps {
			for _, serviceName := range appConfig.GetPublishedServices() {
				// Build full URL: base-url + addr
				fullURL := serverConfig.GetBaseURL() + appConfig.GetAddr()
				serviceLocations[serviceName] = fullURL
			}
		}
	}

	// Create deployment topology
	deployTopo := &DeploymentTopology{
		Name:            deploymentName,
		ConfigOverrides: make(map[string]any),
		Servers:         make(map[string]*ServerTopology),
	}

	// Copy config overrides
	for key, value := range config.GetConfigOverrides() {
		deployTopo.ConfigOverrides[key] = value
	}

	// Build server topologies
	for serverName, serverConfig := range config.GetServers() {
		serverTopo := &ServerTopology{
			Name:           serverName,
			DeploymentName: deploymentName,
			BaseURL:        serverConfig.GetBaseURL(),
			Services:       make([]string, 0),
			RemoteServices: make(map[string]string),
			Apps:           make([]*AppTopology, 0),
		}

		// Collect apps (from Apps slice + shorthand fields)
		apps := serverConfig.GetApps()

		// If shorthand fields are set, create an app from them
		if serverConfig.GetAddr() != "" {
			shorthandApp := &shorthandAppConfig{
				addr:              serverConfig.GetAddr(),
				routers:           serverConfig.GetRouters(),
				publishedServices: serverConfig.GetPublishedServices(),
			}
			// Prepend shorthand app
			apps = append([]AppConfig{shorthandApp}, apps...)
		}

		// Collect SERVER-LEVEL services (published services only)
		serviceMap := make(map[string]bool)
		for _, appConfig := range apps {
			for _, svcName := range appConfig.GetPublishedServices() {
				serviceMap[svcName] = true
			}
		}

		// Convert to slice
		for svcName := range serviceMap {
			serverTopo.Services = append(serverTopo.Services, svcName)
		}

		// Build RemoteServices map (services published on OTHER servers)
		for otherServerName, otherServerConfig := range config.GetServers() {
			if otherServerName == serverName {
				continue // Skip own server
			}

			// Collect apps from other server
			otherApps := otherServerConfig.GetApps()
			if otherServerConfig.GetAddr() != "" {
				shorthandApp := &shorthandAppConfig{
					addr:              otherServerConfig.GetAddr(),
					routers:           otherServerConfig.GetRouters(),
					publishedServices: otherServerConfig.GetPublishedServices(),
				}
				otherApps = append([]AppConfig{shorthandApp}, otherApps...)
			}

			// Add remote services
			for _, appConfig := range otherApps {
				for _, svcName := range appConfig.GetPublishedServices() {
					// Build full URL: base-url + addr
					remoteURL := otherServerConfig.GetBaseURL() + appConfig.GetAddr()
					serverTopo.RemoteServices[svcName] = remoteURL
				}
			}
		}

		// Build app topologies
		for _, appConfig := range apps {
			appTopo := &AppTopology{
				Addr:    appConfig.GetAddr(),
				Routers: make([]string, 0),
			}

			// Collect routers
			appTopo.Routers = append(appTopo.Routers, appConfig.GetRouters()...)

			// Auto-generated routers from published services
			for _, serviceName := range appConfig.GetPublishedServices() {
				routerName := serviceName + "-router"
				appTopo.Routers = append(appTopo.Routers, routerName)
			}

			serverTopo.Apps = append(serverTopo.Apps, appTopo)
		}

		deployTopo.Servers[serverName] = serverTopo
	}

	// Store topology in global registry
	g.StoreDeploymentTopology(deployTopo)

	return nil
}

// shorthandAppConfig is a helper struct for shorthand app creation
type shorthandAppConfig struct {
	addr              string
	routers           []string
	publishedServices []string
}

func (a *shorthandAppConfig) GetAddr() string                { return a.addr }
func (a *shorthandAppConfig) GetRouters() []string           { return a.routers }
func (a *shorthandAppConfig) GetPublishedServices() []string { return a.publishedServices }

// StoreDeploymentTopology stores deployment topology in global registry
func (g *GlobalRegistry) StoreDeploymentTopology(topology *DeploymentTopology) {
	g.deploymentTopologies.Store(topology.Name, topology)

	// Also store server topologies with composite keys
	for serverName, serverTopo := range topology.Servers {
		compositeKey := topology.Name + "." + serverName
		g.serverTopologies.Store(compositeKey, serverTopo)
	}
}

// GetDeploymentTopology retrieves deployment topology by name
func (g *GlobalRegistry) GetDeploymentTopology(deploymentName string) (*DeploymentTopology, bool) {
	if v, ok := g.deploymentTopologies.Load(deploymentName); ok {
		return v.(*DeploymentTopology), true
	}
	return nil, false
}

// GetServerTopology retrieves server topology by composite key "deployment.server"
func (g *GlobalRegistry) GetServerTopology(compositeKey string) (*ServerTopology, bool) {
	if v, ok := g.serverTopologies.Load(compositeKey); ok {
		return v.(*ServerTopology), true
	}
	return nil, false
}

// SetCurrentCompositeKey sets the current server context for runtime resolution
func (g *GlobalRegistry) SetCurrentCompositeKey(compositeKey string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.currentCompositeKey = compositeKey
}

// GetCurrentCompositeKey returns the current server context
func (g *GlobalRegistry) GetCurrentCompositeKey() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentCompositeKey
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
// Services are shutdown in reverse order of their registration (LIFO) to respect dependencies.
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
func (g *GlobalRegistry) ShutdownServices() {
	// Create a snapshot to avoid issues during shutdown
	var snapshot []struct {
		name string
		svc  any
	}

	g.serviceInstances.Range(func(key, value any) bool {
		snapshot = append(snapshot, struct {
			name string
			svc  any
		}{
			name: key.(string),
			svc:  value,
		})
		return true
	})

	// Shutdown in reverse order (LIFO)
	for i := len(snapshot) - 1; i >= 0; i-- {
		item := snapshot[i]
		if shutdownable, ok := item.svc.(Shutdownable); ok {
			if err := shutdownable.Shutdown(); err != nil {
				fmt.Printf("[ShutdownServices] Failed to shutdown service %s: %v\n", item.name, err)
			} else {
				fmt.Printf("[ShutdownServices] Successfully shutdown service: %s\n", item.name)
			}
		}
	}
	fmt.Println("[ShutdownServices] Gracefully shutdown all services.")
}

// ===== HELPER FUNCTIONS =====

// convertMiddlewareNames converts middleware names to middleware instances
// Returns []any containing middleware functions resolved from registry
func convertMiddlewareNames(g *GlobalRegistry, names []string) []any {
	if len(names) == 0 {
		return nil
	}

	middlewares := make([]any, 0, len(names))
	for _, name := range names {
		// Get middleware instance from registry
		if mw, ok := g.GetMiddleware(name); ok {
			middlewares = append(middlewares, mw)
		} else {
			// If middleware not found, add name as string (will be resolved at runtime)
			middlewares = append(middlewares, name)
		}
	}
	return middlewares
}
