package lokstra_registry

import (
	"fmt"
	"log"
	"maps"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

// ServiceLocation represents where a service is deployed
type ServiceLocation struct {
	ServerName string // Which server hosts this service
	BaseURL    string // Base URL of the server
	IsLocal    bool   // true if service is on current server
}

var (
	serviceLocationMap  map[string]*ServiceLocation // Map of serviceName -> location
	currentDeploymentID string                      // Auto-detected from current server
)

func init() {
	serviceLocationMap = make(map[string]*ServiceLocation)
}

// SetCurrentServer sets the current server name.
// This must be called before RegisterConfig() to enable automatic local/remote inference.
func SetCurrentServer(serverName string) {
	currentServerName = serverName
	log.Printf("🖥️  Current server set to: %s", serverName)
}

// GetCurrentServer returns the current server name
func GetCurrentServer() string {
	return currentServerName
}

// GetServiceLocation returns location info for a service
func GetServiceLocation(serviceName string) *ServiceLocation {
	return serviceLocationMap[serviceName]
}

// handles service configurations (supports both simple and layered formats)
func processServices(services *config.ServicesConfig) error {
	// Validate service dependencies first
	if err := config.ValidateServices(services); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	// Flatten services (handles both simple and layered)
	flatServices := services.Flatten()

	// Build dependency map for lazy injection
	depMap := buildDependencyMap(flatServices)

	// Register each service with auto local/remote inference
	for _, svc := range flatServices {
		if err := registerService(svc, depMap); err != nil {
			return fmt.Errorf("service %s: %w", svc.Name, err)
		}
	}

	return nil
}

// registerService registers a single service with automatic local/remote inference
func registerService(svc *config.Service, depMap map[string]any) error {
	if !svc.IsEnabled() {
		return nil
	}

	svcType := svc.Type
	if svcType == "" {
		svcType = svc.Name
	}

	// Get service location
	location := serviceLocationMap[svc.Name]
	if location == nil {
		return fmt.Errorf("service %s not assigned to any server (check servers[].apps[].services)", svc.Name)
	}

	// Inject lazy dependencies into config
	cfg := injectDependencies(svc, depMap)

	if location.IsLocal {
		// LOCAL SERVICE: Use provided config
		log.Printf("   ✅ Registering LOCAL service factory: %s (type=%s)", svc.Name, svcType)
		RegisterLazyService(svc.Name, svcType, cfg, AllowOverride(true))

	} else {
		// REMOTE SERVICE: Use base URL + router name + auto-router config for remote service creation
		log.Printf("   ✅ Registering REMOTE service factory: %s (type=%s, url=%s)",
			svc.Name, svcType, location.BaseURL)

		remoteConfig := map[string]any{
			"base_url":     location.BaseURL,
			"service_name": svc.Name,
			"router":       svc.Name, // Router name is the same as service name
		}

		// Pass auto-router configuration to remote service
		// Always get convention, path-prefix, and resource-name (with fallback logic)
		convention := svc.GetConvention("")
		if convention != "" {
			remoteConfig["convention"] = convention
		}

		pathPrefix := svc.GetPathPrefix()
		if pathPrefix != "" {
			remoteConfig["path-prefix"] = pathPrefix
		}

		resourceName := svc.GetResourceName()
		if resourceName != "" {
			remoteConfig["resource-name"] = resourceName
		}

		pluralResourceName := svc.GetPluralResourceName()
		if pluralResourceName != "" {
			remoteConfig["plural-resource-name"] = pluralResourceName
		}

		// Pass route overrides if available
		routeOverrides := svc.GetRouteOverrides()
		if len(routeOverrides) > 0 {
			routes := make([]any, len(routeOverrides))
			for i, routeOverride := range routeOverrides {
				routes[i] = map[string]any{
					"name":   routeOverride.Name,
					"method": routeOverride.Method,
					"path":   routeOverride.Path,
				}
			}
			remoteConfig["routes"] = routes
		}

		RegisterLazyService(svc.Name, svcType, remoteConfig, AllowOverride(true))
	}

	return nil
}

// injectDependencies clones config and injects lazy service placeholders
// Supports both "service-name" and "local_key:service-name" formats in depends-on
func injectDependencies(svc *config.Service, depMap map[string]any) map[string]any {
	cfg := svc.Config
	if cfg == nil {
		cfg = make(map[string]any)
	}

	// Clone config to avoid mutating original
	injectedCfg := make(map[string]any)
	maps.Copy(injectedCfg, cfg)

	// Process each depends-on entry
	if depMap != nil {
		for _, depEntry := range svc.DependsOn {
			// Parse dependency entry: "local_key:service-name" or just "service-name"
			localKey, serviceName := parseDependencyEntry(depEntry)

			// Get lazy service placeholder for the actual service name
			if lazySvc, ok := depMap[serviceName]; ok {
				// Check if config already has the local key with service name value
				if existingVal, exists := injectedCfg[localKey]; exists {
					if strVal, ok := existingVal.(string); ok && strVal == serviceName {
						// Replace with lazy service placeholder
						injectedCfg[localKey] = lazySvc
					}
				} else {
					// Auto-inject if not in config (convenience feature)
					injectedCfg[localKey] = lazySvc
				}
			}
		}
	}

	return injectedCfg
}

// parseDependencyEntry parses depends-on entry supporting both formats:
// - "service-name" returns ("service-name", "service-name")
// - "local_key:service-name" returns ("local_key", "service-name")
func parseDependencyEntry(entry string) (localKey, serviceName string) {
	if idx := strings.Index(entry, ":"); idx > 0 {
		return entry[:idx], entry[idx+1:]
	}
	return entry, entry
}

// generateRoutersFromCurrentServerServices auto-generates routers only for services explicitly listed in app.routers
func generateRoutersFromCurrentServerServices(c *config.Config) error {
	log.Printf("   🔀 Auto-generating routers from app.routers in current server: %s", currentServerName)

	// Find current server configuration
	var currentServer *config.Server
	for _, srv := range c.Servers {
		if srv.DeploymentID != currentDeploymentID {
			continue // Skip servers not in current deployment
		}
		if srv.Name == currentServerName {
			currentServer = srv
			break
		}
	}
	if currentServer == nil {
		return fmt.Errorf("current server %s not found in configuration", currentServerName)
	}

	// Collect all services deployed locally (for validation)
	localServices := make(map[string]bool)
	for _, app := range currentServer.Apps {
		for _, serviceName := range app.Services {
			localServices[serviceName] = true
		}
	}

	// Collect all routers to be generated from app.routers
	routersToGenerate := make(map[string]bool)
	for _, app := range currentServer.Apps {
		for _, routerName := range app.Routers {
			routersToGenerate[routerName] = true
		}
	}

	// Flatten all services to find their configurations
	flatServices := c.Services.Flatten()
	serviceConfigMap := make(map[string]*config.Service)
	for _, svc := range flatServices {
		serviceConfigMap[svc.Name] = svc
	}

	// Generate routers only for services listed in app.routers
	for routerName := range routersToGenerate {
		// Validate that the router name corresponds to a local service
		if !localServices[routerName] {
			return fmt.Errorf("router %s is declared in app.routers but service %s is not available locally (not in app.services)", routerName, routerName)
		}

		svc := serviceConfigMap[routerName]
		if svc == nil {
			log.Printf("   ⚠️  Router %s declared but service not found in services configuration", routerName)
			continue
		}

		if !svc.IsEnabled() {
			log.Printf("   ⏭️  Service %s is disabled, skipping router generation", routerName)
			continue
		}

		// Generate router from service
		if err := generateRouterFromService(svc); err != nil {
			return fmt.Errorf("failed to generate router for service %s: %w", svc.Name, err)
		}
	}

	return nil
}

// generateRouterFromService creates a router for a single service using convention system
func generateRouterFromService(svc *config.Service) error {
	// Determine convention and prefix from service configuration
	convention := svc.GetConvention("") // No global default, use service-specific or fallback to "rest"
	servicePrefix := svc.GetPathPrefix()
	resourceName := svc.GetResourceName()
	pluralResourceName := svc.GetPluralResourceName()

	log.Printf("   🔀 Generating router for service %s (convention=%s, prefix=%s, resource=%s, plural=%s)",
		svc.Name, convention, servicePrefix, resourceName, pluralResourceName)

	// Get the service factory to create service instance
	serviceFactory := GetServiceFactory(svc.Type, svc.Name)
	if serviceFactory == nil {
		return fmt.Errorf("service factory %s not found", svc.Type)
	}

	// IMPORTANT: Use GetService to get the already-registered service instance
	// This ensures we use the same instance that was created with injected dependencies
	// during registerService() call in processServices()
	var serviceInstance any
	serviceInstance, ok := TryGetService[any](svc.Name)
	if !ok {
		// Fallback: create new instance (shouldn't happen if processServices ran first)
		serviceConfig := svc.Config
		if serviceConfig == nil {
			serviceConfig = make(map[string]any)
		}
		serviceInstance = serviceFactory(serviceConfig)
		if serviceInstance == nil {
			return fmt.Errorf("failed to create service instance for %s", svc.Name)
		}
	}

	// Create router using convention system
	options := router.DefaultServiceRouterOptions().
		WithConvention(convention).
		WithPrefix(servicePrefix).
		WithResourceName(resourceName)

	// Apply plural resource name if specified
	if pluralResourceName != "" {
		options = options.WithPluralResourceName(pluralResourceName)
	}

	// Apply route overrides from config
	routeOverrides := svc.GetRouteOverrides()
	if len(routeOverrides) > 0 {
		log.Printf("   🔧 Applying %d route override(s) for service %s", len(routeOverrides), svc.Name)
		for _, override := range routeOverrides {
			routeMeta := router.RouteMeta{}
			if override.Method != "" {
				routeMeta.HTTPMethod = override.Method
			}
			if override.Path != "" {
				routeMeta.Path = override.Path
			}
			options = options.WithRouteOverride(override.Name, routeMeta)
			log.Printf("      - Override %s: method=%s, path=%s", override.Name, override.Method, override.Path)
		}
	}

	// Create router from service
	serviceRouter := router.NewFromService(serviceInstance, options)

	// Register the router with service name as router name
	RegisterRouter(svc.Name, serviceRouter)

	log.Printf("   ✅ Router generated and registered for service %s", svc.Name)

	return nil
}

// creates a map of service name -> lazy service placeholder
func buildDependencyMap(services []*config.Service) map[string]any {
	depMap := make(map[string]any)

	for _, svc := range services {
		if svc.IsEnabled() {
			depMap[svc.Name] = svc.Name
		}
	}

	return depMap
}

// preRegisterClientRouters registers ClientRouters for all servers in current deployment
// This must be called before processServices() so GetServiceFactory can determine local vs remote
func preRegisterClientRouters(c *config.Config) {
	// log.Printf("   🔍 Pre-registering ClientRouters for deployment: %s", currentDeploymentID)
	for _, srvConfig := range c.Servers {
		if srvConfig.DeploymentID != currentDeploymentID {
			continue // Skip servers not in current deployment
		}

		for _, appConfig := range srvConfig.Apps {
			// Register ClientRouter for each router that will be created
			for _, routerName := range appConfig.Routers {
				// Router may not exist yet, but we know it will be created
				// log.Printf("      Registering ClientRouter: %s @ %s (url=%s%s)",
				// 	routerName, srvConfig.Name, srvConfig.BaseUrl, appConfig.Addr)
				RegisterClientRouter(routerName, srvConfig.Name,
					srvConfig.BaseUrl, appConfig.Addr, 0)
			}
		}
	}

	// Build running registry now that ClientRouters are registered
	// log.Printf("   📋 Building running ClientRouter registry...")
	buildRunningClientRouterRegistry()
}

// Applies the configuration to the lokstra_registry
func RegisterConfig(c *config.Config, serverName string) {

	// Step 1: Apply general configs and find current-server from config if not set
	// This allows overriding current server via config file
	for _, gc := range c.Configs {
		if serverName == "" && gc.Name == "current-server" {
			if cs, ok := gc.Value.(string); ok && cs != "" {
				serverName = cs
			}
		}
		SetConfig(gc.Name, gc.Value)
	}

	// find servername
	if serverName == "" {
		serverName = currentServerName
	}
	if serverName == "" {
		// fallback to first server if not set
		if len(c.Servers) > 0 {
			serverName = c.Servers[0].Name
		} else {
			panic("no servers defined in config")
		}
	}
	if serverName == "" {
		panic("cannot determine current server name")
	}
	SetCurrentServer(serverName)

	// Auto-detect current deployment ID from current server
	found := false
	for _, srv := range c.Servers {
		// Detect current deployment ID
		if srv.Name == currentServerName {
			currentDeploymentID = srv.DeploymentID
			found = true
			break
		}
	}
	if !found {
		panic("current server " + currentServerName + " not found in config servers")
	}

	// Step 2: Build service location map from server configs
	buildServiceLocationMap(c)

	// Step 2.5: Pre-register ClientRouters for ALL servers in current deployment
	// This is needed so GetServiceFactory can determine local vs remote during service creation
	preRegisterClientRouters(c)

	// Step 3: Apply services (flattened, with auto local/remote)
	if err := processServices(&c.Services); err != nil {
		panic("failed to process services: " + err.Error())
	}

	// Step 3.5: Auto-generate routers from services that are deployed in current server
	if err := generateRoutersFromCurrentServerServices(c); err != nil {
		panic("failed to generate routers from current server services: " + err.Error())
	}

	// Step 4: Apply middlewares
	for _, mw := range c.Middlewares {
		if mw.IsEnabled() {
			mwType := mw.Type
			if mwType == "" {
				mwType = mw.Name
			}
			RegisterMiddlewareName(mw.Name, mwType,
				mw.Config, AllowOverride(true))
		}
	}

	// Step 4.5: Process routers and store their config (path-prefix, middlewares)
	// This will be used when routers are retrieved
	routerConfigs := make(map[string]*config.Router)
	for _, routerCfg := range c.Routers {
		routerConfigs[routerCfg.Name] = routerCfg
	}

	// Step 5: Register routers on current server only
	for _, srvConfig := range c.Servers {
		if srvConfig.Name != currentServerName {
			continue // Only register current server
		}
		apps := make([]*app.App, 0, len(srvConfig.Apps))
		for i, appConfig := range srvConfig.Apps {
			routers := make([]router.Router, 0, len(appConfig.Routers))

			// Add routers with their config (path-prefix, middlewares)
			for _, routerName := range appConfig.Routers {
				r := GetRouter(routerName)
				if r == nil {
					panic("router " + routerName + " not found for app " + appConfig.Name)
				}

				// Apply router config if available
				if routerCfg, exists := routerConfigs[routerName]; exists {
					// Clone router to avoid side effects
					r = r.Clone()

					// Combine path prefix: config prefix + code prefix (additive)
					if routerCfg.PathPrefix != "" {
						existingPrefix := r.PathPrefix()
						combinedPrefix := routerCfg.PathPrefix + existingPrefix
						r.SetPathPrefix(combinedPrefix)
					}

					// Apply middlewares from router config
					r.Use(utils.ToAnySlice(routerCfg.Middlewares)...)
				}

				routers = append(routers, r)
			}

			a := app.New(appConfig.GetName(i), appConfig.Addr, routers...)

			// Add reverse proxies if configured
			if len(appConfig.ReverseProxies) > 0 {
				proxies := make([]*app.ReverseProxyConfig, len(appConfig.ReverseProxies))
				for j, rp := range appConfig.ReverseProxies {
					proxy := &app.ReverseProxyConfig{
						Prefix:      rp.Prefix,
						StripPrefix: rp.StripPrefix,
						Target:      rp.Target,
					}
					// Add rewrite config if specified
					if rp.Rewrite != nil && rp.Rewrite.From != "" {
						proxy.Rewrite = &app.ReverseProxyRewrite{
							From: rp.Rewrite.From,
							To:   rp.Rewrite.To,
						}
					}
					proxies[j] = proxy
				}
				a.AddReverseProxies(proxies)
			}

			r := a.GetRouter()
			for r != nil {
				RegisterClientRouter(r.Name(), srvConfig.Name,
					srvConfig.BaseUrl, appConfig.Addr, 0)
				r = r.GetNextChain()
			}
			apps = append(apps, a)
		}
		srv := server.New(srvConfig.Name, apps...)
		srv.DeploymentID = srvConfig.DeploymentID
		srv.BaseUrl = srvConfig.BaseUrl
		RegisterServer(srvConfig.Name, srv, AllowOverride(true))
	}

	log.Printf("✅ Configuration registered for deployment: %s, server: %s",
		currentDeploymentID, currentServerName)
}

// builds a map of serviceName -> ServiceLocation
// by scanning all servers and their apps.services configuration
func buildServiceLocationMap(c *config.Config) {
	serviceLocationMap = make(map[string]*ServiceLocation)
	// Track services per deployment-id to detect duplicates within same deployment
	deploymentServices := make(map[string]map[string]string) // deploymentID -> serviceName -> serverName

	for _, srv := range c.Servers {
		if srv.DeploymentID != currentDeploymentID {
			continue // Skip servers not in current deployment
		}

		// Initialize map for this deployment if not exists
		if deploymentServices[srv.DeploymentID] == nil {
			deploymentServices[srv.DeploymentID] = make(map[string]string)
		}

		for _, app := range srv.Apps {
			for _, serviceName := range app.Services {
				// Check for duplicates ONLY within the same deployment-id
				if existingServer, exists := deploymentServices[srv.DeploymentID][serviceName]; exists {
					panic(fmt.Sprintf(
						"service %s is declared in multiple locations within deployment-id '%s': %s and %s",
						serviceName, srv.DeploymentID, existingServer, srv.Name))
				}

				// Track this service in the deployment
				deploymentServices[srv.DeploymentID][serviceName] = srv.Name

				isLocal := srv.Name == currentServerName
				serviceLocationMap[serviceName] = &ServiceLocation{
					ServerName: srv.Name,
					BaseURL:    srv.BaseUrl,
					IsLocal:    isLocal,
				}

				if isLocal {
					log.Printf("   📍 LOCAL service: %s", serviceName)
				} else {
					log.Printf("   🌐 REMOTE service: %s (server=%s, url=%s)",
						serviceName, srv.Name, srv.BaseUrl)
				}
			}
		}
	}
}
