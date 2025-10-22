package deploy

import (
	"fmt"
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

	// Runtime instances (registered routers, services, middlewares)
	routerInstances     sync.Map // map[string]router.Router
	serviceInstances    sync.Map // map[string]any
	middlewareInstances sync.Map // map[string]request.HandlerFunc

	// Definitions (YAML or code-defined)
	configs         map[string]*schema.ConfigDef
	middlewares     map[string]*schema.MiddlewareDef
	services        map[string]*schema.ServiceDef
	routers         map[string]*schema.RouterDef
	routerOverrides map[string]*schema.RouterOverrideDef

	// Resolved config values (after resolver processing)
	resolvedConfigs map[string]any

	// Deployments (loaded from config)
	deployments map[string]*Deployment
}

// ServiceFactoryEntry holds local and remote factory functions plus metadata
type ServiceFactoryEntry struct {
	Local    ServiceFactory
	Remote   ServiceFactory
	Metadata *ServiceMetadata // Optional metadata for auto-router generation
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
		deployments:         make(map[string]*Deployment),
	}
}

// ===== FACTORY REGISTRATION (CODE) =====

// RegisterServiceType registers a service factory with optional metadata
func (g *GlobalRegistry) RegisterServiceType(serviceType string, local, remote ServiceFactory, options ...RegisterServiceTypeOption) {
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

	g.serviceFactories[serviceType] = &ServiceFactoryEntry{
		Local:    local,
		Remote:   remote,
		Metadata: metadataPtr,
	}
}

// RegisterMiddlewareType registers a middleware factory
func (g *GlobalRegistry) RegisterMiddlewareType(middlewareType string, factory MiddlewareFactory) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.middlewareFactories[middlewareType]; exists {
		panic(fmt.Sprintf("middleware type %s already registered", middlewareType))
	}

	g.middlewareFactories[middlewareType] = factory
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

// GetServiceAny retrieves a service instance by name as any
func (g *GlobalRegistry) GetServiceAny(name string) (any, bool) {
	return g.serviceInstances.Load(name)
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

	// Try to create from middleware definition
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

// ===== DEPLOYMENT MANAGEMENT =====

// RegisterDeployment registers a deployment in the global registry
func (g *GlobalRegistry) RegisterDeployment(name string, deployment *Deployment) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.deployments[name]; exists {
		panic(fmt.Sprintf("deployment %s already registered", name))
	}

	g.deployments[name] = deployment
}

// GetDeployment retrieves a deployment by name
func (g *GlobalRegistry) GetDeployment(name string) (*Deployment, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	dep, ok := g.deployments[name]
	return dep, ok
}

// GetAllDeployments returns all registered deployments
func (g *GlobalRegistry) GetAllDeployments() map[string]*Deployment {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make(map[string]*Deployment, len(g.deployments))
	for name, dep := range g.deployments {
		result[name] = dep
	}
	return result
}
