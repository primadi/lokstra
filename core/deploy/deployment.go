package deploy

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/service"
)

// Deployment represents a deployment configuration
type Deployment struct {
	name     string
	registry *GlobalRegistry

	// Config overrides for this deployment
	configOverrides map[string]any

	// Servers in this deployment
	servers map[string]*Server
}

// Server represents a server in a deployment
type Server struct {
	name       string
	baseURL    string
	deployment *Deployment

	// Apps running on this server
	apps []*App
}

// App represents an application running on a server
type App struct {
	port   int
	server *Server

	// Service instances (lazy-loaded)
	services map[string]*serviceInstance

	// Manual routers
	routers map[string]any // TODO: Use actual router type

	// Service routers (auto-generated)
	serviceRouters map[string]*serviceRouter

	// Remote services (proxies)
	remoteServices map[string]*remoteService
}

// serviceInstance represents a service instance in an app
type serviceInstance struct {
	name       string
	serviceDef *schema.ServiceDef
	instance   any // Cached instance
	resolved   bool
}

// serviceRouter represents an auto-generated router from a service
type serviceRouter struct {
	name       string
	service    string
	convention string
	overrides  string
	router     any // TODO: Use actual router type
}

// remoteService represents a remote service proxy
type remoteService struct {
	name       string
	url        string
	convention string
	overrides  string
	proxy      any // TODO: Use actual proxy type
}

// ===== DEPLOYMENT CREATION =====

// New creates a new deployment
func New(deploymentName string) *Deployment {
	return &Deployment{
		name:            deploymentName,
		registry:        Global(),
		configOverrides: make(map[string]any),
		servers:         make(map[string]*Server),
	}
}

// NewWithRegistry creates a deployment with a specific registry (for testing)
func NewWithRegistry(deploymentName string, registry *GlobalRegistry) *Deployment {
	return &Deployment{
		name:            deploymentName,
		registry:        registry,
		configOverrides: make(map[string]any),
		servers:         make(map[string]*Server),
	}
}

// ===== CONFIG OVERRIDES =====

// SetConfigOverride sets a config override for this deployment
func (d *Deployment) SetConfigOverride(key string, value any) *Deployment {
	d.configOverrides[key] = value
	return d
}

// SetConfigOverrides sets multiple config overrides
func (d *Deployment) SetConfigOverrides(overrides map[string]any) *Deployment {
	for k, v := range overrides {
		d.configOverrides[k] = v
	}
	return d
}

// GetConfig gets a config value (with deployment override support)
func (d *Deployment) GetConfig(key string) (any, bool) {
	// Check deployment overrides first
	if value, ok := d.configOverrides[key]; ok {
		return value, true
	}

	// Fall back to global registry
	return d.registry.GetResolvedConfig(key)
}

// ===== SERVER MANAGEMENT =====

// NewServer creates a new server in this deployment
func (d *Deployment) NewServer(serverName, baseURL string) *Server {
	if _, exists := d.servers[serverName]; exists {
		panic(fmt.Sprintf("server %s already exists in deployment %s", serverName, d.name))
	}

	server := &Server{
		name:       serverName,
		baseURL:    baseURL,
		deployment: d,
		apps:       make([]*App, 0),
	}

	d.servers[serverName] = server
	return server
}

// GetServer returns a server by name
func (d *Deployment) GetServer(serverName string) (*Server, bool) {
	server, ok := d.servers[serverName]
	return server, ok
}

// ===== APP MANAGEMENT =====

// NewApp creates a new app on this server
func (s *Server) NewApp(port int) *App {
	app := &App{
		port:           port,
		server:         s,
		services:       make(map[string]*serviceInstance),
		routers:        make(map[string]any),
		serviceRouters: make(map[string]*serviceRouter),
		remoteServices: make(map[string]*remoteService),
	}

	s.apps = append(s.apps, app)
	return app
}

// ===== SERVICE MANAGEMENT =====

// AddService adds a service to this app
// Uses service definition from global registry
func (a *App) AddService(serviceName string) *App {
	// Get service definition from registry
	serviceDef := a.server.deployment.registry.GetService(serviceName)
	if serviceDef == nil {
		panic(fmt.Sprintf("service %s not defined in global registry", serviceName))
	}

	if _, exists := a.services[serviceName]; exists {
		panic(fmt.Sprintf("service %s already added to app", serviceName))
	}

	a.services[serviceName] = &serviceInstance{
		name:       serviceName,
		serviceDef: serviceDef,
		instance:   nil,
		resolved:   false,
	}

	return a
}

// AddServices adds multiple services
func (a *App) AddServices(serviceNames ...string) *App {
	for _, name := range serviceNames {
		a.AddService(name)
	}
	return a
}

// GetService gets a service instance (instantiates if needed)
func (a *App) GetService(serviceName string) (any, error) {
	svcInst, ok := a.services[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found in app", serviceName)
	}

	// If already resolved, return cached instance
	if svcInst.resolved {
		return svcInst.instance, nil
	}

	// Instantiate service
	instance, err := a.instantiateService(svcInst)
	if err != nil {
		return nil, err
	}

	svcInst.instance = instance
	svcInst.resolved = true

	return instance, nil
}

// instantiateService creates a service instance with dependency injection
func (a *App) instantiateService(svcInst *serviceInstance) (any, error) {
	serviceDef := svcInst.serviceDef
	registry := a.server.deployment.registry

	// Get factory
	factory := registry.GetServiceFactory(serviceDef.Type, true) // TODO: Determine local vs remote
	if factory == nil {
		return nil, fmt.Errorf("service factory %s not registered", serviceDef.Type)
	}

	// Build lazy dependencies
	deps := make(map[string]any)
	for _, depStr := range serviceDef.DependsOn {
		paramName, serviceName := parseDependency(depStr)

		// Create lazy loader for dependency
		// Closure captures serviceName for lazy resolution
		lazyDep := service.LazyLoadWith(func() any {
			depInstance, err := a.GetService(serviceName)
			if err != nil {
				panic(fmt.Sprintf("failed to resolve lazy dependency %s for service %s: %v",
					serviceName, svcInst.name, err))
			}
			return depInstance
		})

		deps[paramName] = lazyDep
	}

	// Resolve config values
	resolvedConfig := make(map[string]any)
	for key, value := range serviceDef.Config {
		resolved, err := registry.ResolveConfigValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve config %s for service %s: %w",
				key, svcInst.name, err)
		}
		resolvedConfig[key] = resolved
	}

	// Call factory
	instance := factory(deps, resolvedConfig)
	return instance, nil
}

// parseDependency parses dependency string
// Format: "paramName:serviceName" or just "serviceName"
// Returns: paramName, serviceName
func parseDependency(depStr string) (string, string) {
	parts := strings.SplitN(depStr, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return depStr, depStr
}

// ===== ROUTER MANAGEMENT =====

// AddRouter adds a manual router to this app
func (a *App) AddRouter(routerName string, router any) *App {
	if _, exists := a.routers[routerName]; exists {
		panic(fmt.Sprintf("router %s already added to app", routerName))
	}

	a.routers[routerName] = router
	return a
}

// AddServiceRouter adds a service router (auto-generated from service)
// Uses service router definition from global registry
func (a *App) AddServiceRouter(serviceRouterName string) *App {
	// Get service router definition from registry
	srDef := a.server.deployment.registry.GetServiceRouter(serviceRouterName)
	if srDef == nil {
		panic(fmt.Sprintf("service router %s not defined in global registry", serviceRouterName))
	}

	if _, exists := a.serviceRouters[serviceRouterName]; exists {
		panic(fmt.Sprintf("service router %s already added to app", serviceRouterName))
	}

	a.serviceRouters[serviceRouterName] = &serviceRouter{
		name:       serviceRouterName,
		service:    srDef.Service,
		convention: srDef.Convention,
		overrides:  srDef.Overrides,
		router:     nil, // TODO: Create router
	}

	return a
}

// AddServiceRouters adds multiple service routers
func (a *App) AddServiceRouters(serviceRouterNames ...string) *App {
	for _, name := range serviceRouterNames {
		a.AddServiceRouter(name)
	}
	return a
}

// ===== REMOTE SERVICE MANAGEMENT =====

// AddRemoteService adds a remote service proxy
func (a *App) AddRemoteService(serviceName, url string, opts ...RemoteServiceOption) *App {
	remote := &remoteService{
		name: serviceName,
		url:  url,
	}

	for _, opt := range opts {
		opt.apply(remote)
	}

	if _, exists := a.remoteServices[serviceName]; exists {
		panic(fmt.Sprintf("remote service %s already added to app", serviceName))
	}

	a.remoteServices[serviceName] = remote
	return a
}

// RemoteServiceOption configures a remote service
type RemoteServiceOption interface {
	apply(*remoteService)
}

type remoteServiceOptionFunc func(*remoteService)

func (f remoteServiceOptionFunc) apply(rs *remoteService) {
	f(rs)
}

// WithConvention sets the convention for remote service
func WithConvention(convention string) RemoteServiceOption {
	return remoteServiceOptionFunc(func(rs *remoteService) {
		rs.convention = convention
	})
}

// WithOverrides sets the overrides for remote service
func WithOverrides(overrides string) RemoteServiceOption {
	return remoteServiceOptionFunc(func(rs *remoteService) {
		rs.overrides = overrides
	})
}

// ===== GETTERS =====

// Name returns the deployment name
func (d *Deployment) Name() string {
	return d.name
}

// Registry returns the global registry
func (d *Deployment) Registry() *GlobalRegistry {
	return d.registry
}

// Servers returns all servers
func (d *Deployment) Servers() map[string]*Server {
	return d.servers
}

// Name returns the server name
func (s *Server) Name() string {
	return s.name
}

// BaseURL returns the server base URL
func (s *Server) BaseURL() string {
	return s.baseURL
}

// Apps returns all apps on this server
func (s *Server) Apps() []*App {
	return s.apps
}

// Port returns the app port
func (a *App) Port() int {
	return a.port
}

// Services returns all services in this app
func (a *App) Services() map[string]*serviceInstance {
	return a.services
}

// Routers returns all manual routers
func (a *App) Routers() map[string]any {
	return a.routers
}

// ServiceRouters returns all service routers
func (a *App) ServiceRouters() map[string]*serviceRouter {
	return a.serviceRouters
}

// RemoteServices returns all remote services
func (a *App) RemoteServices() map[string]*remoteService {
	return a.remoteServices
}
