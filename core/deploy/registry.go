package deploy

import (
	"fmt"
	"sync"

	"github.com/primadi/lokstra/core/deploy/resolver"
	"github.com/primadi/lokstra/core/deploy/schema"
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

	// Definitions (YAML or code-defined)
	configs         map[string]*schema.ConfigDef
	middlewares     map[string]*schema.MiddlewareDef
	services        map[string]*schema.ServiceDef
	routers         map[string]*schema.RouterDef
	routerOverrides map[string]*schema.RouterOverrideDef
	serviceRouters  map[string]*schema.ServiceRouterDef

	// Resolved config values (after resolver processing)
	resolvedConfigs map[string]any
}

// ServiceFactoryEntry holds local and remote factory functions
type ServiceFactoryEntry struct {
	Local  ServiceFactory
	Remote ServiceFactory
}

// ServiceFactory creates a service instance
// deps: dependencies resolved as map[paramName]*service.Cached[any]
// config: configuration for this service instance
// Dependencies are lazy-loaded - call .Get() to resolve
type ServiceFactory func(deps map[string]any, config map[string]any) any

// MiddlewareFactory creates a middleware instance
type MiddlewareFactory func(config map[string]any) any

var (
	globalRegistry     *GlobalRegistry
	globalRegistryOnce sync.Once
)

// Global returns the singleton global registry
func Global() *GlobalRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewGlobalRegistry()
	})
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
		serviceRouters:      make(map[string]*schema.ServiceRouterDef),
		resolvedConfigs:     make(map[string]any),
	}
}

// ===== FACTORY REGISTRATION (CODE) =====

// RegisterServiceType registers a service factory
func (g *GlobalRegistry) RegisterServiceType(serviceType string, local, remote ServiceFactory) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.serviceFactories[serviceType]; exists {
		panic(fmt.Sprintf("service type %s already registered", serviceType))
	}

	g.serviceFactories[serviceType] = &ServiceFactoryEntry{
		Local:  local,
		Remote: remote,
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

// DefineRouter defines a manual router
func (g *GlobalRegistry) DefineRouter(def *schema.RouterDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.routers[def.Name]; exists {
		panic(fmt.Sprintf("router %s already defined", def.Name))
	}

	g.routers[def.Name] = def
}

// DefineRouterOverride defines router overrides
func (g *GlobalRegistry) DefineRouterOverride(def *schema.RouterOverrideDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.routerOverrides[def.Name]; exists {
		panic(fmt.Sprintf("router override %s already defined", def.Name))
	}

	g.routerOverrides[def.Name] = def
}

// DefineServiceRouter defines a service router
func (g *GlobalRegistry) DefineServiceRouter(def *schema.ServiceRouterDef) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.serviceRouters[def.Name]; exists {
		panic(fmt.Sprintf("service router %s already defined", def.Name))
	}

	g.serviceRouters[def.Name] = def
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

// GetMiddleware returns a middleware definition
func (g *GlobalRegistry) GetMiddleware(name string) *schema.MiddlewareDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.middlewares[name]
}

// GetService returns a service definition
func (g *GlobalRegistry) GetService(name string) *schema.ServiceDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.services[name]
}

// GetRouter returns a router definition
func (g *GlobalRegistry) GetRouter(name string) *schema.RouterDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.routers[name]
}

// GetRouterOverride returns router overrides
func (g *GlobalRegistry) GetRouterOverride(name string) *schema.RouterOverrideDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.routerOverrides[name]
}

// GetServiceRouter returns a service router definition
func (g *GlobalRegistry) GetServiceRouter(name string) *schema.ServiceRouterDef {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.serviceRouters[name]
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
