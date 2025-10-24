package deploy

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/primadi/lokstra/core/deploy/resolver"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
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

	// Definitions (YAML or code-defined)
	configs         map[string]*schema.ConfigDef
	middlewares     map[string]*schema.MiddlewareDef
	services        map[string]*schema.ServiceDef
	routers         map[string]*schema.RouterDef
	routerOverrides map[string]*schema.RouterOverrideDef

	// Resolved config values (after resolver processing)
	resolvedConfigs map[string]any

	// Topology storage (2-Layer Architecture)
	// Single source of truth for runtime topology
	deploymentTopologies sync.Map // map[deploymentName]*DeploymentTopology
	serverTopologies     sync.Map // map[compositeKey]*ServerTopology (key: "deployment.server")
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
type ServiceMetadata struct {
	Resource        string            // Singular resource name (e.g., "user")
	ResourcePlural  string            // Plural resource name (e.g., "users")
	Convention      string            // Convention type (e.g., "rest", "rpc")
	RouteOverrides  map[string]string // Method name -> custom path
	HiddenMethods   []string          // Methods to hide from router
	PathPrefix      string            // Path prefix for all routes
	MiddlewareNames []string          // Middleware names to apply
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
		middlewares:         make(map[string]*schema.MiddlewareDef),
		services:            make(map[string]*schema.ServiceDef),
		routers:             make(map[string]*schema.RouterDef),
		routerOverrides:     make(map[string]*schema.RouterOverrideDef),
		resolvedConfigs:     make(map[string]any),
		// Topology maps (deploymentTopologies, serverTopologies) use sync.Map, no initialization needed
	}
}

// ===== FACTORY REGISTRATION (CODE) =====

// RegisterServiceType registers a service factory with optional metadata
// Supports three factory signatures (auto-wrapped by framework):
//   - func(deps, cfg map[string]any) any - full control (canonical)
//   - func(cfg map[string]any) any       - only config
//   - func() any                          - no params
//
// Both local and remote factories support all three signatures.
func (g *GlobalRegistry) RegisterServiceType(serviceType string, local, remote any, options ...RegisterServiceTypeOption) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.serviceFactories[serviceType]; exists {
		panic(fmt.Sprintf("service type %s already registered", serviceType))
	}

	// Build metadata from options
	metadata := &ServiceMetadata{
		Convention: "rest", // Default convention
	}
	for _, opt := range options {
		opt(metadata)
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

// DefineMiddleware defines a middleware instance
func (g *GlobalRegistry) DefineMiddleware(def *schema.MiddlewareDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.middlewares[def.Name]; exists {
		panic(fmt.Sprintf("middleware %s already defined", def.Name))
	}

	g.middlewares[def.Name] = def
}

// DefineService defines a service instance
func (g *GlobalRegistry) DefineService(def *schema.ServiceDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.services[def.Name]; exists {
		panic(fmt.Sprintf("service %s already defined", def.Name))
	}

	g.services[def.Name] = def
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

// DefineRouterOverride defines router overrides
func (g *GlobalRegistry) DefineRouterOverride(name string, def *schema.RouterOverrideDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.routerOverrides[name]; exists {
		panic(fmt.Sprintf("router override %s already defined", name))
	}

	g.routerOverrides[name] = def
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

// GetMiddlewareDef returns a middleware definition
func (g *GlobalRegistry) GetMiddlewareDef(name string) *schema.MiddlewareDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.middlewares[name]
}

// GetServiceDef returns a service definition
func (g *GlobalRegistry) GetServiceDef(name string) *schema.ServiceDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.services[name]
}

// GetRouterDef returns a router definition
func (g *GlobalRegistry) GetRouterDef(name string) *schema.RouterDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.routers[name]
}

// GetRouterOverride returns a router override definition
func (g *GlobalRegistry) GetRouterOverride(name string) *schema.RouterOverrideDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.routerOverrides[name]
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
//   - func() any                    - no params (simplest!)
//
// Dependencies are resolved manually via lokstra_registry.MustGetService() inside factory.
//
// Example with config:
//
//	lokstra_registry.RegisterLazyService("db-main", func(cfg map[string]any) any {
//	    return db.NewConnection(cfg["dsn"].(string))
//	}, map[string]any{"dsn": "postgresql://localhost/main"})
//
// Example without params:
//
//	lokstra_registry.RegisterLazyService("user-repo", func() any {
//	    return repository.NewUserRepository()
//	}, nil)
//
// For explicit dependency injection, use RegisterLazyServiceWithDeps instead.
func (g *GlobalRegistry) RegisterLazyService(name string, factory any, config map[string]any) {
	g.RegisterLazyServiceWithDeps(name, factory, nil, config)
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
func (g *GlobalRegistry) GetServiceAny(name string) (any, bool) {
	// Check eager registry first
	if svc, ok := g.serviceInstances.Load(name); ok {
		return svc, true
	}

	// Check lazy registry and create if needed
	onceAny, hasOnce := g.lazyServiceOnce.Load(name)
	if !hasOnce {
		return nil, false
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

	// Try to create from middleware definition (YAML config)
	g.mu.RLock()
	mwDef, defExists := g.middlewares[name]
	g.mu.RUnlock()

	if !defExists {
		return nil
	}

	// Get factory
	factory := g.GetMiddlewareFactory(mwDef.Type)
	if factory == nil {
		return nil
	}

	// Create instance
	mw := factory(mwDef.Config)
	if handlerFunc, ok := mw.(request.HandlerFunc); ok {
		// Cache it
		g.RegisterMiddleware(name, handlerFunc)
		return handlerFunc
	}

	return nil
}

// ===== TOPOLOGY MANAGEMENT (2-Layer Architecture) =====

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
