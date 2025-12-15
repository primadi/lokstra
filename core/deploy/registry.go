package deploy

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/internal/registry"
)

// GlobalRegistry stores all global definitions (configs, middlewares, services, etc.)
// These are shared across all deployments
type GlobalRegistry struct {
	mu sync.RWMutex

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

	// Lazy router factories (for deferred router creation)
	lazyRouterFactories sync.Map // map[string]func() router.Router

	// Definitions (YAML or code-defined)
	routers map[string]*schema.RouterDef
	// Note: routerOverrides removed - overrides are now inline in RouterDef
	// Note: middlewares map removed - use middlewareEntries sync.Map (unified API)
	// Note: serviceDefs removed - unified with lazyServiceFactories (2-phase resolution)
	// Note: configs map removed - use resolvedConfigs only (simplified)

	// Config values (runtime and YAML-loaded configs)
	// All configs are stored here after loader's 2-step resolution
	resolvedConfigs map[string]any

	// Topology storage (2-Layer Architecture)
	// Single source of truth for runtime topology
	deploymentTopologies sync.Map // map[deploymentName]*DeploymentTopology
	serverTopologies     sync.Map // map[compositeKey]*ServerTopology (key: "deployment.server")

	// Original config (for inline definitions normalization)
	deployConfig *schema.DeployConfig

	// Current server context (for runtime service resolution)
	currentCompositeKey string // "deployment.server" - set by SetCurrentServer
}

// deferredServiceDef holds service definition for deferred instantiation
// Used when RegisterLazyService is called with string factory type
// Instantiation is deferred until first access, allowing auto-detect LOCAL/REMOTE
// ServiceFactoryEntry holds local and remote factory functions plus metadata
type ServiceFactoryEntry struct {
	Local    ServiceFactory
	Remote   ServiceFactory
	Metadata *ServiceMetadata // Optional metadata for auto-router generation
}

// LazyServiceEntry holds a lazy service factory and its config
// Supports 2-phase resolution: unresolved (FactoryType) → resolved (Factory)
type LazyServiceEntry struct {
	// Phase 1: Set by LoadConfig (unresolved)
	FactoryType string // Service factory type (e.g., "email_smtp")

	// Phase 2: Set by RegisterDefinitionsForRuntime (resolved)
	Factory func(deps, config map[string]any) any

	Config   map[string]any
	Deps     map[string]string // Dependency mapping: key in factory -> service name in registry
	resolved bool              // Has factory been resolved from FactoryType?
}

// IsResolved returns true if the factory has been resolved from FactoryType
func (e *LazyServiceEntry) IsResolved() bool {
	return e.resolved
}

// ResolveFactory sets the factory function and marks the entry as resolved
func (e *LazyServiceEntry) ResolveFactory(factory func(deps, config map[string]any) any) {
	e.Factory = factory
	e.resolved = true
}

// ServiceMetadata holds metadata for service auto-generation
// ServiceMetadata holds metadata for a service type registration
// Can be populated from ServiceTypeConfig or legacy functional options
type ServiceMetadata struct {
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
	Name            string
	DeploymentName  string
	BaseURL         string
	ConfigOverrides map[string]any    // Server-level config overrides (highest priority)
	Services        []string          // Service names (server-level, shared)
	RemoteServices  map[string]string // serviceName -> remoteBaseURL (empty string if local)
	Apps            []*AppTopology
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
		serviceFactories:    make(map[string]*ServiceFactoryEntry),
		middlewareFactories: make(map[string]MiddlewareFactory),
		routers:             make(map[string]*schema.RouterDef),
		resolvedConfigs:     make(map[string]any),
		// Topology maps and middlewareEntries use sync.Map, no initialization needed
	}
}

// ResetGlobalRegistryForTesting resets the global registry singleton to a fresh state.
// WARNING: This function is ONLY for testing purposes!
// Do NOT use in production code as it will clear all registered services, middlewares, and configs.
func ResetGlobalRegistryForTesting() {
	globalRegistry = NewGlobalRegistry()
	registry.SetGlobal(globalRegistry)
}

// ===== FACTORY REGISTRATION (CODE) =====

// Factory signatures (auto-wrapped by framework):
//   - func(deps, cfg map[string]any) any - full control (canonical)
//   - func(cfg map[string]any) any       - only config
//   - func() any                          - no params
//
// Both local and remote factories support all three signatures.
// RegisterRouterServiceType registers a service type with HTTP routing configuration.
// Use this for services that expose HTTP endpoints (annotated with @RouterService).
// For simple infrastructure services (DB, Redis, etc), use RegisterServiceType instead.
//
// Parameters:
//   - serviceType: Unique identifier for this service type
//   - local: Factory for local deployment (same process)
//   - remote: Factory for remote deployment (HTTP client)
//   - config: Optional routing configuration (path prefix, middlewares, route overrides)
func (g *GlobalRegistry) RegisterRouterServiceType(serviceType string, local, remote any,
	config *ServiceTypeConfig) {
	logger.LogDebug("[RegisterRouterServiceType CALLED] serviceType=%s", serviceType)

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.serviceFactories[serviceType]; exists {
		panic(fmt.Sprintf("service type %s already registered", serviceType))
	}

	// Build metadata from config
	metadata := &ServiceMetadata{}

	if config != nil {
		metadata.PathPrefix = config.PathPrefix
		metadata.MiddlewareNames = config.Middlewares
		metadata.HiddenMethods = config.Hidden

		// Convert RouteConfig to RouteMetadata
		if len(config.RouteOverrides) > 0 {
			metadata.RouteOverrides = make(map[string]RouteMetadata)
			for methodName, routeConfig := range config.RouteOverrides {
				// Parse "METHOD /path" format from Path field if Method is empty
				// This handles annotation-generated code format: Path: "POST /users/{id}"
				method := routeConfig.Method
				path := routeConfig.Path

				if method == "" && path != "" {
					parts := strings.SplitN(path, " ", 2)
					if len(parts) == 2 {
						method = parts[0]
						path = parts[1]
					}
				}

				metadata.RouteOverrides[methodName] = RouteMetadata{
					Method:      method,
					Path:        path,
					Middlewares: routeConfig.Middlewares,
				}
			}
		}
	}

	// Debug: log metadata before filtering
	logger.LogDebug("[RegisterServiceType] %s (before filter): PathPrefix='%s', RouteOverrides=%d",
		serviceType, metadata.PathPrefix, len(metadata.RouteOverrides))

	// Store metadata if any meaningful configuration is provided
	var metadataPtr *ServiceMetadata
	hasConfig := metadata.PathPrefix != "" ||
		len(metadata.RouteOverrides) > 0 ||
		len(metadata.MiddlewareNames) > 0 ||
		len(metadata.HiddenMethods) > 0

	if hasConfig {
		metadataPtr = metadata
		// Debug log
		logger.LogDebug("[RegisterServiceType] %s: STORED - PathPrefix=%s, RouteOverrides count=%d",
			serviceType, metadata.PathPrefix, len(metadata.RouteOverrides))
		for methodName, route := range metadata.RouteOverrides {
			logger.LogDebug("  - %s: method=%s, path=%s", methodName, route.Method, route.Path)
		}
	} else {
		logger.LogDebug("[RegisterServiceType] %s: NOT STORED (no meaningful config)", serviceType)
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

	logger.LogDebug("[RegisterRouterServiceType] %s: registered (local=%v, remote=%v)",
		serviceType, localFactory != nil, remoteFactory != nil)
}

// RegisterServiceType registers a simple service type without HTTP routing.
// Use this for infrastructure services like database pools, Redis clients, metrics, etc.
// For services that expose HTTP endpoints, use RegisterRouterServiceType instead.
//
// Parameters:
//   - serviceType: Unique identifier for this service type
//   - factory: Factory function (supports multiple signatures - see normalizeServiceFactory)
func (g *GlobalRegistry) RegisterServiceType(serviceType string, factory any) {
	g.RegisterRouterServiceType(serviceType, factory, nil, nil)
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

// ===== DEFINITION REGISTRATION (YAML OR CODE) =====

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
		logger.LogDebug("[GetServiceMetadata] serviceType '%s' NOT FOUND", serviceType)
		return nil
	}

	if entry.Metadata != nil {
		logger.LogDebug("[GetServiceMetadata] serviceType '%s' FOUND: PathPrefix=%s, RouteOverrides=%d",
			serviceType, entry.Metadata.PathPrefix, len(entry.Metadata.RouteOverrides))
	} else {
		logger.LogDebug("[GetServiceMetadata] serviceType '%s' FOUND but Metadata=nil", serviceType)
	}

	return entry.Metadata
}

// GetRouterDef returns a router definition
func (g *GlobalRegistry) GetRouterDef(name string) *schema.RouterDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.routers[name]
}

// ===== CONFIG RESOLUTION =====

// SetConfig sets a runtime configuration value
// Useful for:
//   - Runtime detection results (mode, environment)
//   - Computed values (expensive calculations)
//   - Dynamic service discovery
//   - Feature flags
//
// If value is a map[string]any, also flattens nested values automatically:
//
//	SetConfig("db", {"host": "x", "port": 5432})
//	→ Also sets: "db.host" = "x", "db.port" = 5432
//
// IMPORTANT: When setting a map, deletes all existing "key.*" entries first
// to prevent stale nested values:
//
//	SetConfig("db", {"host": "x", "port": 5432}) // Creates db, db.host, db.port
//	SetConfig("db", {"host": "y"})               // Deletes db.port (stale), keeps only db, db.host
//
// Key is automatically converted to lowercase for case-insensitive access
func (g *GlobalRegistry) SetConfig(key string, value any) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.resolvedConfigs == nil {
		g.resolvedConfigs = make(map[string]any)
	}

	// Store with lowercase key for case-insensitive access
	lowerKey := strings.ToLower(key)

	// If value is a map, delete all existing nested keys first (prevent stale data)
	if _, ok := value.(map[string]any); ok {
		g.deleteNestedKeys(lowerKey)
	}

	// Store the value
	g.resolvedConfigs[lowerKey] = value

	// If value is a map, also flatten nested values
	if nestedMap, ok := value.(map[string]any); ok {
		g.flattenAndStoreNested(lowerKey, nestedMap)
	}
}

// deleteNestedKeys deletes all keys with prefix "key.*"
// Called before setting a map value to prevent stale nested data
func (g *GlobalRegistry) deleteNestedKeys(prefix string) {
	prefixDot := prefix + "."
	for key := range g.resolvedConfigs {
		if key == prefix || strings.HasPrefix(key, prefixDot) {
			delete(g.resolvedConfigs, key)
		}
	}
}

// flattenAndStoreNested recursively flattens nested map values
// Called internally by SetConfig when value is a map
func (g *GlobalRegistry) flattenAndStoreNested(prefix string, values map[string]any) {
	for key, value := range values {
		fullKey := prefix + "." + strings.ToLower(key)
		g.resolvedConfigs[fullKey] = value

		// Recurse if value is also a map
		if nestedMap, ok := value.(map[string]any); ok {
			g.flattenAndStoreNested(fullKey, nestedMap)
		}
	}
}

// GetConfig returns a config value
// Supports both flat access ("db_main.dsn") and nested access ("db_main" returns map)
// Key lookup is case-insensitive
func (g *GlobalRegistry) GetConfig(name string) (any, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Convert to lowercase for case-insensitive lookup
	lowerName := strings.ToLower(name)

	// Try direct lookup first (flat access)
	if value, ok := g.resolvedConfigs[lowerName]; ok {
		return value, true
	}

	// Try nested access: collect all keys starting with "name."
	prefix := lowerName + "."
	nested := make(map[string]any)
	hasNested := false

	for key, value := range g.resolvedConfigs {
		if after, ok := strings.CutPrefix(key, prefix); ok {
			// Remove prefix and reconstruct nested structure
			subKey := after

			// Handle further nesting (e.g., "db_main.connection.pool.size")
			if strings.Contains(subKey, ".") {
				setNestedValue(nested, subKey, value)
			} else {
				nested[subKey] = value
			}
			hasNested = true
		}
	}

	if hasNested {
		return nested, true
	}

	return nil, false
}

// setNestedValue sets a value in a nested map using dot notation
// Example: setNestedValue(map, "connection.pool.size", 10) creates {"connection": {"pool": {"size": 10}}}
func setNestedValue(target map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := target

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part: set value
			current[part] = value
		} else {
			// Intermediate part: ensure map exists
			if _, exists := current[part]; !exists {
				current[part] = make(map[string]any)
			}
			// Navigate deeper
			if nextMap, ok := current[part].(map[string]any); ok {
				current = nextMap
			} else {
				// Type conflict: cannot navigate deeper
				return
			}
		}
	}
}

// SimpleResolver resolves variables in the format ${key} or ${key:default}
// by looking up values from the config registry via GetConfig().
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
//	  auth-prefix: ${AUTH_PREFIX:/api/auth}  # ENV variable with default
//	  api-version: /api                      # Static value
//	  db-host: ${@VAULT:database/host:localhost}  # Vault with default
func (g *GlobalRegistry) SimpleResolver(input string) string {
	return os.Expand(input, func(placeholder string) string {
		// Parse placeholder: "key" or "key:default"
		parts := strings.SplitN(placeholder, ":", 2)
		key := parts[0]
		defaultValue := ""

		if len(parts) == 2 {
			defaultValue = parts[1]
		}

		// Look up from resolved config registry
		value, ok := g.GetConfig(key)
		if !ok {
			return defaultValue
		}

		// Convert value to string
		if strValue, ok := value.(string); ok {
			return strValue
		}

		return defaultValue
	})
}

// ===== RUNTIME INSTANCE REGISTRATION =====

// RegisterRouter registers a router instance
// If a RouterDef with the same name exists and has a PathPrefix, it will be applied
func (g *GlobalRegistry) RegisterRouter(name string, r router.Router) {
	if _, exists := g.routerInstances.Load(name); exists {
		panic(fmt.Sprintf("router %s already registered", name))
	}

	// Check if RouterDef exists with PathPrefix
	if routerDef := g.GetRouterDef(name); routerDef != nil {
		if routerDef.PathPrefix != "" {
			// Apply PathPrefix from RouterDef (YAML router-definitions)
			logger.LogDebug("🔧 Applying PathPrefix '%s' to router '%s' from router-definitions", routerDef.PathPrefix, name)
			r = r.SetPathPrefix(routerDef.PathPrefix)
		}

		// Apply PathRewrites if defined
		if len(routerDef.PathRewrites) > 0 {
			rewrites := make(map[string]string)
			for _, rewrite := range routerDef.PathRewrites {
				rewrites[rewrite.Pattern] = rewrite.Replacement
			}
			r = r.SetPathRewrites(rewrites)
		}
	}

	logger.LogDebug("🔧 RegisterRouter: storing router '%s' at %p (type=%T)", name, r, r)
	g.routerInstances.Store(name, r)
}

// RegisterRouterFactory registers a lazy router factory that will be instantiated
// when the runtime is ready (after all services are resolved).
// This allows router registration to depend on services that need runtime resolution.
//
// Example:
//
//	lokstra_registry.RegisterRouterFactory("email-router", func() lokstra.Router {
//	    emailService := lokstra_registry.GetService[EmailService]("email-api-service")
//	    return emailService.GetRouter()
//	})
func (g *GlobalRegistry) RegisterRouterFactory(name string, factory func() router.Router) {
	logger.LogDebug("🔧 RegisterRouterFactory: registering lazy router '%s'", name)
	g.lazyRouterFactories.Store(name, factory)
}

// instantiateLazyRouters creates router instances from registered factories
// func (g *GlobalRegistry) InstantiateLazyRouters() {
// 	logger.LogDebug("🔧 InstantiateLazyRouters: starting lazy router instantiation")
// 	count := 0
// 	g.lazyRouterFactories.Range(func(nameAny, factoryAny any) bool {
// 		name := nameAny.(string)
// 		factory := factoryAny.(func() router.Router)

// 		// Skip if already instantiated
// 		if _, exists := g.routerInstances.Load(name); exists {
// 			logger.LogDebug("🔧 Lazy router '%s': already instantiated, skipping", name)
// 			return true
// 		}

// 		logger.LogDebug("🔧 Instantiating lazy router: '%s'", name)
// 		r := factory()
// 		logger.LogDebug("🔧 Lazy router '%s': factory returned %T, registering", name, r)
// 		g.RegisterRouter(name, r)
// 		count++
// 		return true
// 	})
// 	logger.LogDebug("🔧 InstantiateLazyRouters: completed, instantiated %d routers", count)
// }

// GetRouter retrieves a router instance by name
// If not found in routerInstances, checks lazyRouterFactories and instantiates if needed
func (g *GlobalRegistry) GetRouter(name string) router.Router {
	// Check if already instantiated
	if v, ok := g.routerInstances.Load(name); ok {
		r := v.(router.Router)
		logger.LogDebug("🔍 GetRouter('%s'): found router %p (type=%T)", name, r, r)
		return r
	}

	// Check lazy router factories and instantiate if found
	if factoryAny, ok := g.lazyRouterFactories.Load(name); ok {
		factory := factoryAny.(func() router.Router)
		logger.LogDebug("🔍 GetRouter('%s'): found lazy factory, instantiating...", name)
		r := factory()
		logger.LogDebug("🔍 GetRouter('%s'): lazy factory returned %T, registering", name, r)
		g.RegisterRouter(name, r)
		return r
	}

	// Check if this is a service router (format: "serviceName-router")
	// If so, try to instantiate the service first
	if strings.HasSuffix(name, "-router") {
		serviceName := strings.TrimSuffix(name, "-router")

		// Check if service exists (lazy or instance)
		if g.HasService(serviceName) {
			// Try to instantiate service (this will trigger router creation in the service factory)
			logger.LogDebug("🔍 GetRouter('%s'): service '%s' exists, attempting to instantiate...", name, serviceName)

			// Get service instance (this will instantiate if lazy)
			serviceInstance, ok := g.GetServiceAny(serviceName)
			logger.LogDebug("🔍 GetRouter('%s'): GetServiceAny returned ok=%v, instance=%v", name, ok, serviceInstance != nil)

			if ok && serviceInstance != nil {
				logger.LogDebug("🔍 GetRouter('%s'): service '%s' instantiated successfully, checking router again...", name, serviceName)

				// Check if router was created during service instantiation
				if v, ok := g.routerInstances.Load(name); ok {
					r := v.(router.Router)
					logger.LogDebug("🔍 GetRouter('%s'): router created during service instantiation", name)
					return r
				}

				// Service instantiated but router not found - may need to create manually
				logger.LogDebug("🔍 GetRouter('%s'): service instantiated but router not auto-created, checking metadata...", name)

				// Get service definition
				serviceDef := g.GetDeferredServiceDef(serviceName)
				if serviceDef == nil {
					logger.LogDebug("🔍 GetRouter('%s'): service definition not found", name)
					return nil
				}

				// Get service metadata
				metadata := g.GetServiceMetadata(serviceDef.Type)
				if metadata == nil {
					logger.LogDebug("🔍 GetRouter('%s'): service metadata not found", name)
					return nil
				}

				// Check if service has router config
				hasRouterConfig := len(metadata.RouteOverrides) > 0 || metadata.PathPrefix != ""
				if !hasRouterConfig {
					logger.LogDebug("🔍 GetRouter('%s'): service has no router configuration", name)
					return nil
				}

				// Create router from service using autogen
				logger.LogDebug("🔍 GetRouter('%s'): creating router from service instance", name)

				// Get RouterDef if exists
				routerDef := g.GetRouterDef(name)
				finalPrefix := metadata.PathPrefix
				if routerDef != nil && routerDef.PathPrefix != "" {
					finalPrefix = routerDef.PathPrefix
				}

				// Build ServiceRouterOptions
				// Convert RouteMetadata to RouteMeta
				routeOverrides := make(map[string]router.RouteMeta)
				for methodName, routeMeta := range metadata.RouteOverrides {
					middlewares := make([]any, len(routeMeta.Middlewares))
					for i, mw := range routeMeta.Middlewares {
						middlewares[i] = mw
					}

					routeOverrides[methodName] = router.RouteMeta{
						HTTPMethod:  routeMeta.Method,
						Path:        routeMeta.Path,
						Middlewares: middlewares,
					}
				}

				// Convert router-level middleware names to string slice
				routerMiddlewares := metadata.MiddlewareNames

				opts := &router.ServiceRouterOptions{
					Prefix:         finalPrefix,
					Middlewares:    routerMiddlewares,
					RouteOverrides: routeOverrides,
				}

				// Create router using NewFromService
				r := router.NewFromService(serviceInstance, opts)
				g.RegisterRouter(name, r)
				logger.LogDebug("🔍 GetRouter('%s'): router created and registered", name)
				return r
			}
		}
	}

	logger.LogDebug("🔍 GetRouter('%s'): NOT FOUND", name)
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
	logger.LogDebug("ℹ️  Registered service instance: '%s'\n", name)
}

// UnregisterService removes a service instance from the registry
func (g *GlobalRegistry) UnregisterService(name string) {
	g.serviceInstances.Delete(name)
	logger.LogDebug("ℹ️  Unregistered service instance: '%s'\n", name)
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
//	lokstra_registry.RegisterLazyService("db_main", func(cfg map[string]any) any {
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
	// Extract depends-on from config if present
	var deps map[string]string
	if depsRaw, ok := config["depends-on"]; ok {
		var dependsOn []string
		switch depsVal := depsRaw.(type) {
		case []string:
			dependsOn = depsVal
		case []any:
			// Handle YAML unmarshaling []any
			dependsOn = make([]string, len(depsVal))
			for i, d := range depsVal {
				dependsOn[i] = d.(string)
			}
		}

		// Create deps map: key = service name, value = service name
		// Support "paramName:serviceName" notation (e.g., "cfg:@store.implementation")
		if len(dependsOn) > 0 {
			deps = make(map[string]string, len(dependsOn))
			for _, dep := range dependsOn {
				// Parse "paramName:serviceName" or just "serviceName"
				parts := strings.SplitN(dep, ":", 2)
				if len(parts) == 2 {
					// "cfg:@store.implementation" -> deps["cfg"] = "@store.implementation"
					deps[parts[0]] = parts[1]
				} else {
					// "logger" -> deps["logger"] = "logger"
					deps[dep] = dep
				}
			}
			logger.LogDebug("📦 RegisterLazyService '%s': extracted %d dependencies from config: %v", name, len(deps), dependsOn)
		}
	} // Delegate to RegisterLazyServiceWithDeps
	g.RegisterLazyServiceWithDeps(name, factory, deps, config)
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
		case []any:
			// Handle YAML unmarshaling []any
			dependsOn = make([]string, len(deps))
			for i, d := range deps {
				dependsOn[i] = d.(string)
			}
		}
	}

	// Create unresolved lazy service entry (Phase 1: store FactoryType string)
	// Will be resolved to actual Factory function in RegisterDefinitionsForRuntime
	depsMap := make(map[string]string)
	for _, dep := range dependsOn {
		depsMap[dep] = dep
	}

	entry := &LazyServiceEntry{
		FactoryType: factoryType,
		Factory:     nil, // Unresolved - will be set in Phase 2
		Config:      config,
		Deps:        depsMap,
		resolved:    false,
	}

	g.lazyServiceFactories.Store(name, entry)
	// NOTE: Do NOT create sync.Once here!
	// sync.Once will be created in GetServiceAny when entry is resolved
	// This prevents premature instantiation before factory is available
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
	// Type detection: string factory type name vs inline function
	if factoryTypeName, ok := factory.(string); ok {
		// String factory type - store definition for deferred instantiation
		g.registerDeferredService(name, factoryTypeName, config)
		return
	}

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

	// If deps is nil OR empty, try to extract from config["depends-on"]
	// This handles RegisterLazyService calls where deps=nil
	if deps == nil && config != nil {
		if depsRaw, ok := config["depends-on"]; ok {
			deps = make(map[string]string)
			switch depsArray := depsRaw.(type) {
			case []string:
				for _, depStr := range depsArray {
					// Parse "paramName:serviceName" or just "serviceName"
					// serviceName can be "@config.key" for config-based resolution
					parts := strings.SplitN(depStr, ":", 2)
					if len(parts) == 2 {
						paramName := parts[0]
						serviceName := parts[1]
						deps[paramName] = serviceName
					} else {
						// No explicit param name - use service name as key
						deps[depStr] = depStr
					}
				}
			case []any:
				for _, d := range depsArray {
					if depStr, ok := d.(string); ok {
						// Parse "paramName:serviceName" or just "serviceName"
						parts := strings.SplitN(depStr, ":", 2)
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
			}
		}
	}

	entry := &LazyServiceEntry{
		Factory:  normFactory,
		Config:   config,
		Deps:     deps, // Store dependency mapping
		resolved: true, // Already has Factory function
	}

	logger.LogDebug("📦 RegisterLazyServiceWithDeps '%s': stored with %d dependencies: %v", name, len(deps), deps)
	g.lazyServiceFactories.Store(name, entry)
	g.lazyServiceOnce.Store(name, &sync.Once{})
}

// RegisterLazyServiceUnresolved stores an unresolved lazy service entry
// This is called during config loading when we only have the factory type name
// The actual factory function will be resolved later in RegisterDefinitionsForRuntime
func (g *GlobalRegistry) RegisterLazyServiceUnresolved(name, factoryType string, deps map[string]string, config map[string]any) {
	if name == "" || factoryType == "" {
		panic(fmt.Sprintf("service name and factory type must not be empty (name=%s, type=%s)", name, factoryType))
	}

	entry := &LazyServiceEntry{
		FactoryType: factoryType,
		Factory:     nil, // Will be resolved later
		Config:      config,
		Deps:        deps,
		resolved:    false, // Mark as unresolved
	}

	g.lazyServiceFactories.Store(name, entry)
	g.lazyServiceOnce.Store(name, &sync.Once{})
}

// GetLazyServiceEntry retrieves a lazy service entry by name (for resolution checking)
func (g *GlobalRegistry) GetLazyServiceEntry(name string) *LazyServiceEntry {
	if entryAny, ok := g.lazyServiceFactories.Load(name); ok {
		return entryAny.(*LazyServiceEntry)
	}
	return nil
}

// GetServiceAny retrieves a service instance by name as any
// If not found in eager registry, checks lazy registry and instantiates
// If still not found, checks service-definitions and auto-creates lazy service
func (g *GlobalRegistry) GetServiceAny(name string) (any, bool) {
	return g.getServiceAnyWithStack(name, []string{})
}

// getServiceAnyWithStack is internal version with circular dependency detection
func (g *GlobalRegistry) getServiceAnyWithStack(name string, resolutionStack []string) (any, bool) {
	logger.LogDebug("🔍 GetServiceAny('%s'): starting resolution, stack=%v", name, resolutionStack)

	// Handle @ prefix - resolve actual service name from config
	// Example: "@store.order-repository" reads config key "store.order-repository"
	// and gets the actual service name to inject
	if after, ok := strings.CutPrefix(name, "@"); ok {
		logger.LogDebug("🔍 GetServiceAny('%s'): has @ prefix, resolving from config key '%s'", name, after)
		configKey := after
		configValue, ok := g.GetConfig(configKey)
		if !ok {
			logger.LogDebug("🔍 GetServiceAny('%s'): config key '%s' NOT FOUND", name, configKey)
			return nil, false
		}

		actualServiceName, ok := configValue.(string)
		if !ok || actualServiceName == "" {
			logger.LogDebug("🔍 GetServiceAny('%s'): config value is not string or empty: %v", name, configValue)
			return nil, false
		}

		logger.LogDebug("🔍 GetServiceAny('%s'): resolved to actual service '%s'", name, actualServiceName)
		// Recursively resolve the actual service (add to stack to detect circular deps)
		return g.getServiceAnyWithStack(actualServiceName, append(resolutionStack, name))
	}

	// Check for circular dependency
	for _, svcName := range resolutionStack {
		if svcName == name {
			// Build dependency chain for error message
			chain := utils.NewSliceAndAppend(resolutionStack, name)
			panic(fmt.Sprintf("circular dependency detected: %s", strings.Join(chain, " → ")))
		}
	}

	// Check eager registry first
	if svc, ok := g.serviceInstances.Load(name); ok {
		logger.LogDebug("🔍 GetServiceAny('%s'): found in eager registry (already instantiated)", name)
		return svc, true
	}

	// Add to resolution stack
	newStack := utils.NewSliceAndAppend(resolutionStack, name)

	// Check lazy registry and create if needed
	onceAny, hasOnce := g.lazyServiceOnce.Load(name)
	if !hasOnce {
		logger.LogDebug("🔍 GetServiceAny('%s'): NOT in lazyServiceOnce, checking lazyServiceFactories...", name)

		// Not in lazy registry - check if in lazyServiceFactories with unresolved entry
		if entryAny, exists := g.lazyServiceFactories.Load(name); exists {
			logger.LogDebug("🔍 GetServiceAny('%s'): found in lazyServiceFactories", name)
			entry := entryAny.(*LazyServiceEntry)

			// If unresolved (Phase 1 - from registerDeferredService), resolve it now
			if !entry.resolved {
				logger.LogDebug("🔍 GetServiceAny('%s'): entry UNRESOLVED, resolving factory type '%s'...", name, entry.FactoryType)
				// Get factory for the service type
				factory := g.GetServiceFactory(entry.FactoryType, true) // true = local factory
				if factory == nil {
					logger.LogDebug("🔍 GetServiceAny('%s'): factory '%s' NOT FOUND!", name, entry.FactoryType)
					panic(fmt.Sprintf("service factory '%s' not registered for service '%s'", entry.FactoryType, name))
				}

				// Resolve the factory (modifies the entry in-place since it's a pointer)
				entry.Factory = factory
				entry.resolved = true
			}

			// Create sync.Once if not exists (handles case where entry was resolved externally)
			if _, hasOnceAlready := g.lazyServiceOnce.Load(name); !hasOnceAlready {
				logger.LogDebug("🔍 GetServiceAny('%s'): creating sync.Once (entry was resolved=%v)", name, entry.resolved)
				g.lazyServiceOnce.Store(name, &sync.Once{})
			}

			// Now proceed with instantiation below
			onceAny, hasOnce = g.lazyServiceOnce.Load(name)
		} else {
			// Not in lazy registry - try auto-registration from serviceFactories
			// Convention: service name = factory type (e.g., "email-smtp" service uses "email-smtp" factory)
			g.mu.RLock()
			factoryEntry, hasFactory := g.serviceFactories[name]
			g.mu.RUnlock()

			if hasFactory && factoryEntry.Local != nil {
				// Auto-register as lazy service with default config
				logger.LogDebug("🔧 Auto-registering service '%s' from factory type '%s' (default config)", name, name)
				entry := &LazyServiceEntry{
					FactoryType: name,
					Factory:     factoryEntry.Local,
					Config:      make(map[string]any), // Empty config
					Deps:        make(map[string]string),
					resolved:    true,
				}

				g.lazyServiceFactories.Store(name, entry)
				g.lazyServiceOnce.Store(name, &sync.Once{})

				onceAny, hasOnce = g.lazyServiceOnce.Load(name)
			}
		}

		if !hasOnce {
			logger.LogDebug("🔍 GetServiceAny('%s'): NOT FOUND in any registry, returning false", name)
			return nil, false
		}
	} else {
		logger.LogDebug("🔍 GetServiceAny('%s'): found in lazyServiceOnce, will instantiate", name)
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

		// If unresolved inside once.Do, resolve now (handles race condition)
		if !entry.resolved {
			factory := g.GetServiceFactory(entry.FactoryType, true)
			if factory == nil {
				panic(fmt.Sprintf("service factory '%s' not registered for service '%s'", entry.FactoryType, name))
			}
			entry.Factory = factory
			entry.resolved = true
		}

		// Resolve dependencies if specified
		var resolvedDeps map[string]any
		if len(entry.Deps) > 0 {
			logger.LogDebug("📦 Service '%s': resolving %d dependencies: %v", name, len(entry.Deps), entry.Deps)
			resolvedDeps = make(map[string]any, len(entry.Deps))
			for factoryKey, serviceName := range entry.Deps {
				// Recursively resolve dependency with circular detection
				// @ prefix is handled automatically by getServiceAnyWithStack
				logger.LogDebug("📦 Service '%s': resolving dependency '%s' -> '%s'", name, factoryKey, serviceName)
				depSvc, ok := g.getServiceAnyWithStack(serviceName, newStack)
				if !ok {
					panic(fmt.Sprintf("lazy service %s: dependency %s not found", name, serviceName))
				}
				logger.LogDebug("📦 Service '%s': dependency '%s' resolved to: %T", name, factoryKey, depSvc)
				// Use factoryKey (may include @ prefix) as key for factory lookup
				resolvedDeps[factoryKey] = depSvc
			}
			logger.LogDebug("📦 Service '%s': all dependencies resolved, calling factory", name)
		} else {
			logger.LogDebug("📦 Service '%s': no dependencies, calling factory directly", name)
		}

		// Call factory with resolved deps or nil
		// Check if this is a remote service (has "remote" in config)
		if _, isRemote := entry.Config["remote"]; isRemote {
			logger.LogDebug("📦 Creating remote service wrapper: '%s'", name)
		} else {
			logger.LogDebug("📦 Creating service instance: '%s'", name)
		}
		instance := entry.Factory(resolvedDeps, entry.Config)
		logger.LogDebug("📦 Service '%s' created: instance=%p, type=%T", name, instance, instance)
		g.serviceInstances.Store(name, instance)
	})

	// Return cached instance
	svc, ok := g.serviceInstances.Load(name)
	return svc, ok
}

// HasService checks if a service is registered in the lazy service registry
// or instantiated in the eager registry.
func (g *GlobalRegistry) HasService(name string) bool {
	// Check if defined in lazy registry (resolved or unresolved)
	if _, ok := g.lazyServiceFactories.Load(name); ok {
		return true
	}

	// Check if instantiated in eager registry
	if _, ok := g.serviceInstances.Load(name); ok {
		return true
	}

	return false
}

// MergeRegistryServicesToConfig merges services from registry (RegisterLazyService)
// into config.ServiceDefinitions. This allows services registered via code to be
// available in config for dependency resolution and topology checks.
func (g *GlobalRegistry) MergeRegistryServicesToConfig(config *schema.DeployConfig) {
	g.lazyServiceFactories.Range(func(key, value any) bool {
		serviceName := key.(string)
		entry := value.(*LazyServiceEntry)

		// Skip if already exists in config (YAML takes priority)
		if _, exists := config.ServiceDefinitions[serviceName]; exists {
			return true // continue iteration
		}

		// Skip if already resolved (has inline factory function)
		// Resolved entries don't need to be merged to config because they already have
		// the factory function ready to instantiate - no factory type lookup needed
		if entry.resolved {
			logger.LogDebug("⏭️  Skipping merge for '%s': already resolved with inline factory", serviceName)
			return true // continue iteration
		}

		// Convert Deps map to DependsOn slice
		dependsOn := make([]string, 0, len(entry.Deps))
		for dep := range entry.Deps {
			dependsOn = append(dependsOn, dep)
		}

		// Add to config.ServiceDefinitions
		config.ServiceDefinitions[serviceName] = &schema.ServiceDef{
			Name:      serviceName,
			Type:      entry.FactoryType,
			DependsOn: dependsOn,
			Config:    entry.Config,
		}

		return true // continue iteration
	})
}

// GetDeferredServiceDef retrieves a service definition by name.
// Returns the definition if found in lazyServiceFactories, or nil if not found.
// This is primarily used by wrapper functions that need access to service metadata.
func (g *GlobalRegistry) GetDeferredServiceDef(name string) *schema.ServiceDef {
	logger.LogDebug("[GetDeferredServiceDef] looking for '%s'", name)
	if entryAny, ok := g.lazyServiceFactories.Load(name); ok {
		entry := entryAny.(*LazyServiceEntry)
		logger.LogDebug("[GetDeferredServiceDef] FOUND '%s': Type=%s", name, entry.FactoryType)

		// Convert Deps map to DependsOn slice
		dependsOn := make([]string, 0, len(entry.Deps))
		for dep := range entry.Deps {
			dependsOn = append(dependsOn, dep)
		}

		return &schema.ServiceDef{
			Name:      name,
			Type:      entry.FactoryType,
			DependsOn: dependsOn,
			Config:    entry.Config,
		}
	}
	logger.LogDebug("[GetDeferredServiceDef] NOT FOUND '%s'", name)
	return nil
}

// autoRegisterLazyService auto-registers a service from service-definitions as a lazy service
// This enables zero-config pattern - services are created on-demand from YAML definitions
// Logic: Check if published on another server → REMOTE, else → LOCAL from service-definitions
// func (g *GlobalRegistry) autoRegisterLazyService(name string, def *schema.ServiceDef) {
// 	// Get current deployment context
// 	currentKey := g.GetCurrentCompositeKey()
// 	logger.LogDebug("[autoRegisterLazyService] service '%s', currentKey='%s'", name, currentKey)
// 	if currentKey == "" {
// 		// No current context - default to LOCAL
// 		logger.LogDebug("[autoRegisterLazyService] No currentKey - registering '%s' as LOCAL", name)
// 		g.autoRegisterLocalService(name, def)
// 		return
// 	}

// 	// Get current server topology
// 	currentServerTopo, ok := g.GetServerTopology(currentKey)
// 	if !ok {
// 		// No topology found - default to LOCAL
// 		logger.LogDebug("[autoRegisterLazyService] No topology found for '%s' - registering '%s' as LOCAL", currentKey, name)
// 		g.autoRegisterLocalService(name, def)
// 		return
// 	}

// 	// Check if service is published on another server (REMOTE)
// 	remoteBaseURL, isRemote := currentServerTopo.RemoteServices[name]
// 	logger.LogDebug("[autoRegisterLazyService] service '%s': isRemote=%v, remoteBaseURL='%s'", name, isRemote, remoteBaseURL)
// 	if isRemote {
// 		// Register as REMOTE service (HTTP proxy)
// 		logger.LogDebug("[autoRegisterLazyService] Registering '%s' as REMOTE -> %s", name, remoteBaseURL)
// 		g.AutoRegisterRemoteService(name, def, remoteBaseURL)
// 		return
// 	}

// 	// Not remote - register as LOCAL
// 	logger.LogDebug("[autoRegisterLazyService] Registering '%s' as LOCAL", name)
// 	g.autoRegisterLocalService(name, def)
// }

// autoRegisterLocalService registers a service as LOCAL (from factory)
// func (g *GlobalRegistry) autoRegisterLocalService(name string, def *schema.ServiceDef) {
// 	// Get factory
// 	factory := g.GetServiceFactory(def.Type, true) // true = local factory
// 	if factory == nil {
// 		panic(fmt.Sprintf("service factory '%s' not registered for service '%s'", def.Type, name))
// 	}

// 	// Parse dependencies from DependsOn field
// 	deps := make(map[string]string)
// 	if len(def.DependsOn) > 0 {
// 		for _, depStr := range def.DependsOn {
// 			// Format: "paramName:serviceName" or just "serviceName"
// 			parts := strings.Split(depStr, ":")
// 			if len(parts) == 2 {
// 				paramName := parts[0]
// 				serviceName := parts[1]
// 				deps[paramName] = serviceName
// 			} else {
// 				// No explicit param name - use service name as key
// 				deps[depStr] = depStr
// 			}
// 		}
// 	}

// 	// Register as lazy service with wrapper factory
// 	// Factory expects service.Cached for dependencies, so we wrap resolved deps
// 	g.RegisterLazyServiceWithDeps(name, func(resolvedDeps, cfg map[string]any) any {
// 		// Wrap resolved dependencies as service.Cached
// 		// This allows factories to use service.Cast[T](deps["key"])
// 		lazyDeps := make(map[string]any)
// 		for key, depSvc := range resolvedDeps {
// 			depSvcCopy := depSvc // Capture for closure
// 			lazyDeps[key] = service.LazyLoadWith(func() any { return depSvcCopy })
// 		}

// 		// Call original factory
// 		logger.LogDebug("📦 Creating service instance: '%s' (type: %s)", name, def.Type)
// 		return factory(lazyDeps, cfg)
// 	}, deps, def.Config)
// }

// AutoRegisterRemoteService registers a service as REMOTE (HTTP proxy)
func (g *GlobalRegistry) AutoRegisterRemoteService(name string, def *schema.ServiceDef, remoteBaseURL string) {
	logger.LogDebug("🌐 Creating remote service proxy: '%s' -> %s", name, remoteBaseURL)

	// Get remote factory
	factory := g.GetServiceFactory(def.Type, false) // false = remote factory
	if factory == nil {
		panic(fmt.Sprintf("remote service factory '%s' not registered for service '%s'", def.Type, name))
	}

	// Get service metadata for proxy.Service creation
	metadata := g.GetServiceMetadata(def.Type)

	// Create proxy.Service for HTTP calls with explicit route mappings
	var proxyService *proxy.Service
	if metadata != nil && len(metadata.RouteOverrides) > 0 {
		// Use explicit route mappings from RegisterServiceType
		routeMap := make(map[string]proxy.RouteMapping)
		for methodName, routeMeta := range metadata.RouteOverrides {
			// Build full path with prefix
			path := routeMeta.Path
			if metadata.PathPrefix != "" {
				path = strings.TrimSuffix(metadata.PathPrefix, "/") + "/" + strings.TrimPrefix(path, "/")
			}

			routeMap[methodName] = proxy.RouteMapping{
				HTTPMethod: routeMeta.Method,
				Path:       path,
			}
		}

		proxyService = proxy.NewService(remoteBaseURL, routeMap)

		// Apply hidden methods if specified
		if len(metadata.HiddenMethods) > 0 {
			proxyService = proxyService.WithHiddenMethods(metadata.HiddenMethods...)
		}
	} else {
		// No metadata - service must have explicit route mappings
		// Create empty proxy (routes must be added manually)
		logger.LogDebug("⚠️  Remote service '%s' has no route metadata - proxy created with empty routes", name)
		proxyService = proxy.NewService(remoteBaseURL, make(map[string]proxy.RouteMapping))
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
	GetConfigOverrides() map[string]any
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
		if !g.HasService(serviceName) {
			return fmt.Errorf("published service '%s' not found in service registry", serviceName)
		}

		// Get service definition to find service type
		serviceDef := g.GetDeferredServiceDef(serviceName)
		if serviceDef == nil {
			return fmt.Errorf("published service '%s' definition not found", serviceName)
		}

		// Get service metadata from factory registration
		if metadata := g.GetServiceMetadata(serviceDef.Type); metadata != nil {
			// Define router with metadata
			routerDef := &schema.RouterDef{
				PathPrefix:  metadata.PathPrefix,
				Middlewares: metadata.MiddlewareNames,
				Hidden:      metadata.HiddenMethods,
			}

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
			g.DefineRouter(routerName, routerDef)
		}
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
			Name:            serverName,
			DeploymentName:  deploymentName,
			BaseURL:         serverConfig.GetBaseURL(),
			ConfigOverrides: serverConfig.GetConfigOverrides(),
			Services:        make([]string, 0),
			RemoteServices:  make(map[string]string),
			Apps:            make([]*AppTopology, 0),
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

// Implement ServerConfig interface for code-based config (not used in shorthandAppConfig)
func (a *shorthandAppConfig) GetConfigOverrides() map[string]any { return nil }

var FirstServer string

// StoreDeploymentTopology stores deployment topology in global registry (case-insensitive)
func (g *GlobalRegistry) StoreDeploymentTopology(topology *DeploymentTopology) {
	lowerName := strings.ToLower(topology.Name)
	g.deploymentTopologies.Store(lowerName, topology)

	// Also store server topologies with composite keys (case-insensitive)
	for serverName, serverTopo := range topology.Servers {
		compositeKey := lowerName + "." + strings.ToLower(serverName)
		g.serverTopologies.Store(compositeKey, serverTopo)
		if FirstServer == "" {
			FirstServer = compositeKey
		}
	}
}

// GetDeploymentTopology retrieves deployment topology by name (case-insensitive)
func (g *GlobalRegistry) GetDeploymentTopology(deploymentName string) (*DeploymentTopology, bool) {
	lowerName := strings.ToLower(deploymentName)
	if v, ok := g.deploymentTopologies.Load(lowerName); ok {
		return v.(*DeploymentTopology), true
	}
	return nil, false
}

// GetServerTopology retrieves server topology by composite key "deployment.server"
// GetServerTopology retrieves server topology by composite key (case-insensitive)
func (g *GlobalRegistry) GetServerTopology(compositeKey string) (*ServerTopology, bool) {
	lowerKey := strings.ToLower(compositeKey)
	if v, ok := g.serverTopologies.Load(lowerKey); ok {
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

// StoreDeployConfig stores the original deploy configuration for inline definitions normalization
func (g *GlobalRegistry) StoreDeployConfig(config *schema.DeployConfig) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.deployConfig = config
}

// GetDeployConfig returns the stored deploy configuration
func (g *GlobalRegistry) GetDeployConfig() *schema.DeployConfig {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.deployConfig
}

// GetFirstServerCompositeKey returns the first available server composite key from server topologies
// Returns empty string if no server topologies are found
func (g *GlobalRegistry) GetFirstServerCompositeKey() string {
	return FirstServer
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
				logger.LogInfo("[ShutdownServices] Failed to shutdown service %s: %v\n", item.name, err)
			} else {
				logger.LogInfo("[ShutdownServices] Successfully shutdown service: %s\n", item.name)
			}
		}
	}
	logger.LogInfo("[ShutdownServices] Gracefully shutdown all services.")
}
