package lokstra_registry

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

var (
	// Current server composite key: "deploymentName.serverName"
	currentCompositeKey string
)

// getFirstServerCompositeKey returns the first available server composite key from the global registry
func getFirstServerCompositeKey() string {
	registry := deploy.Global()
	return registry.GetFirstServerCompositeKey()
}

// LoadAndBuild loads config and builds ALL deployments into Global registry
func LoadAndBuild(configPaths []string) error {
	return loader.LoadAndBuild(configPaths)
}

// SetCurrentServer sets the current server using composite key: "deploymentName.serverName"
// If compositeKey is empty, it will automatically use the first deployment and server available
// Example: SetCurrentServer("order-service.order-api")
func SetCurrentServer(compositeKey string) error {
	// If compositeKey is empty, get the first deployment and server
	if compositeKey == "" {
		firstKey := getFirstServerCompositeKey()
		if firstKey == "" {
			return fmt.Errorf("no server topologies found in global registry")
		}
		compositeKey = firstKey
		log.Printf("🎯 Auto-selected first server: %s", compositeKey)
	}

	parts := strings.Split(compositeKey, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid server key format, expected 'deployment.server', got: %s", compositeKey)
	}

	// Validate that server topology exists in Global registry
	_, ok := deploy.Global().GetServerTopology(compositeKey)
	if !ok {
		return fmt.Errorf("server topology '%s' not found in global registry", compositeKey)
	}

	// Set current context (both in package variable and GlobalRegistry)
	currentCompositeKey = compositeKey
	deploy.Global().SetCurrentCompositeKey(compositeKey)
	return nil
}

// GetCurrentCompositeKey returns the current composite key "deployment.server"
func GetCurrentCompositeKey() string {
	return currentCompositeKey
}

// GetCurrentServerName returns the current server name
func GetCurrentServerName() string {
	if currentCompositeKey == "" {
		return ""
	}
	parts := strings.Split(currentCompositeKey, ".")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// GetCurrentDeploymentName returns the current deployment name
func GetCurrentDeploymentName() string {
	if currentCompositeKey == "" {
		return ""
	}
	parts := strings.Split(currentCompositeKey, ".")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// PrintCurrentServerInfo prints information about the current server configuration
func PrintCurrentServerInfo() error {
	if currentCompositeKey == "" {
		return fmt.Errorf("no server set - call SetCurrentServer first")
	}

	// Get topology from Global registry
	registry := deploy.Global()
	serverTopo, ok := registry.GetServerTopology(currentCompositeKey)
	if !ok {
		return fmt.Errorf("server topology '%s' not found in global registry", currentCompositeKey)
	}

	// Extract deployment and server names
	deploymentName := GetCurrentDeploymentName()
	serverName := GetCurrentServerName()

	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Printf("│ Server: %-35s │\n", serverName)
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│ Deployment: %-31s │\n", deploymentName)
	fmt.Printf("│ Base URL: %-33s │\n", serverTopo.BaseURL)

	// Server-level services (shared across all apps)
	if len(serverTopo.Services) > 0 {
		fmt.Println("│                                             │")
		fmt.Println("│ Services (server-level):                    │")
		for _, svcName := range serverTopo.Services {
			fmt.Printf("│   • %-39s │\n", svcName)
		}
	}

	// Apps from topology (only show addr + routers)
	if len(serverTopo.Apps) > 0 {
		fmt.Println("│                                             │")
		fmt.Printf("│ Apps: %-37d │\n", len(serverTopo.Apps))
		for i, appTopo := range serverTopo.Apps {
			fmt.Printf("│   App #%d:                                   │\n", i+1)
			fmt.Printf("│     Addr: %-33s │\n", appTopo.Addr)

			// Routers
			if len(appTopo.Routers) > 0 {
				fmt.Println("│     Routers:                                │")
				for _, routerName := range appTopo.Routers {
					fmt.Printf("│       • %-35s │\n", routerName)
				}
			}
		}
	}

	fmt.Println("└─────────────────────────────────────────────┘")
	fmt.Println()

	return nil
}

// RunCurrentServer builds and runs the current server based on deployment config
func RunCurrentServer(timeout time.Duration) error {
	if currentCompositeKey == "" {
		return fmt.Errorf("no server set - call SetCurrentServer first")
	}

	// Get server topology from Global registry
	registry := deploy.Global()
	serverTopo, ok := registry.GetServerTopology(currentCompositeKey)
	if !ok {
		return fmt.Errorf("server topology '%s' not found in global registry", currentCompositeKey)
	}

	// Get original config for inline definitions normalization
	config := registry.GetDeployConfig()
	if config != nil {
		// Extract deployment and server names from composite key
		deploymentName := GetCurrentDeploymentName()
		serverName := GetCurrentServerName()

		// Perform lazy normalization of inline definitions for this server only
		// This updates the config structure (moves inline definitions to global with normalized names)
		err := loader.NormalizeInlineDefinitionsForServer(config, deploymentName, serverName)
		if err != nil {
			return fmt.Errorf("failed to normalize inline definitions: %w", err)
		}

		// Perform runtime registration of all definitions (global + normalized inline)
		// This registers middlewares, services (with remote/local logic), and auto-generates routers
		err = loader.RegisterDefinitionsForRuntime(registry, config, deploymentName, serverName, serverTopo)
		if err != nil {
			return fmt.Errorf("failed to register definitions for runtime: %w", err)
		}

		log.Printf("📝 Normalized and registered definitions for server %s.%s", deploymentName, serverName)
	}

	// NOTE: registerLazyServicesForServer is NO LONGER NEEDED
	// All service registration (including remote/local logic) is now handled in RegisterDefinitionsForRuntime

	// Get apps from topology
	serverName := GetCurrentServerName()
	if len(serverTopo.Apps) == 0 {
		return fmt.Errorf("server '%s' has no apps configured", serverName)
	}

	// Build one core app per AppTopology and collect them
	var coreApps []*app.App
	for i, appTopo := range serverTopo.Apps {
		// Build routers for this app
		if len(appTopo.Routers) == 0 {
			return fmt.Errorf("app %d has no routers configured", i+1)
		}

		var routers []router.Router
		for _, routerName := range appTopo.Routers {
			// Try to get manually registered router first
			r := GetRouter(routerName)

			var isAutoGenerated bool

			// If not found, try to build from router definition (auto-generated)
			if r == nil {
				autoRouter, err := BuildRouterFromDefinition(routerName)
				if err != nil {
					return fmt.Errorf("router '%s' not found in registry and failed to auto-build: %w", routerName, err)
				}
				r = autoRouter
				isAutoGenerated = true
				// Derive service name from router name: "{service-name}-router" -> "{service-name}"
				serviceName := strings.TrimSuffix(routerName, "-router")
				log.Printf("✨ Auto-generated router '%s' from service '%s'\n", routerName, serviceName)
			}

			// Apply overrides from router-definitions (works for both manual and auto-generated routers)
			routerDef := deploy.Global().GetRouterDef(routerName)
			if routerDef != nil {
				// Apply path prefix override if specified
				if routerDef.PathPrefix != "" {
					r.SetPathPrefix(routerDef.PathPrefix)
					routerType := "manual"
					if isAutoGenerated {
						routerType = "auto-generated"
					}
					log.Printf("🔧 Applied path prefix override to %s router '%s': %s\n", routerType, routerName, routerDef.PathPrefix)
				}

				// Apply path rewrites if specified
				if len(routerDef.PathRewrites) > 0 {
					rewrites := make(map[string]string)
					for _, rw := range routerDef.PathRewrites {
						rewrites[rw.Pattern] = rw.Replacement
					}
					r.SetPathRewrites(rewrites)
					log.Printf("🔧 Applied %d path rewrite rule(s) to router '%s'\n", len(rewrites), routerName)
				}

				// Apply router-level middleware overrides if specified
				if len(routerDef.Middlewares) > 0 {
					// Convert middleware names to []any
					middlewares := make([]any, len(routerDef.Middlewares))
					for i, name := range routerDef.Middlewares {
						middlewares[i] = name
					}
					// Apply middleware overrides from YAML config
					router.ApplyMiddlewares(r, middlewares...)
					log.Printf("🔧 Applied router-level middlewares to '%s': %v\n", routerName, routerDef.Middlewares)
				}

				// Apply route-level overrides (custom routes)
				if len(routerDef.Custom) > 0 {
					for _, customRoute := range routerDef.Custom {
						var options []any

						// Add method override if specified
						if customRoute.Method != "" {
							options = append(options, route.WithMethodOption(customRoute.Method))
						}

						// Add path override if specified
						if customRoute.Path != "" {
							options = append(options, route.WithPathOption(customRoute.Path))
						}

						// Add middlewares if specified
						if len(customRoute.Middlewares) > 0 {
							for _, mwName := range customRoute.Middlewares {
								options = append(options, mwName)
							}
						}

						// Apply all options to the route
						if len(options) > 0 {
							err := r.UpdateRoute(customRoute.Name, options...)
							if err != nil {
								log.Printf("⚠️  Warning: Failed to update route '%s' in router '%s': %v\n",
									customRoute.Name, routerName, err)
							} else {
								log.Printf("🔧 Applied route overrides to '%s.%s' (method: %s, path: %s, middlewares: %v)\n",
									routerName, customRoute.Name, customRoute.Method, customRoute.Path, customRoute.Middlewares)
							}
						}
					}
				}
			}

			routers = append(routers, r)
		}

		// Create Lokstra App for this deploy app. Name it using serverName#index to keep unique names
		appName := fmt.Sprintf("%s#%s", serverName, strconv.Itoa(i+1))

		// Address is already resolved by ResolveConfigs()
		coreApp := app.New(appName, appTopo.Addr, routers...)
		coreApps = append(coreApps, coreApp)
	}

	// Create core Server and run (delegates to core/server/server.go)
	coreServer := server.New(serverName, coreApps...)
	coreServer.PrintStartInfo()

	// Delegate to coreServer.Run() - no code duplication!
	return coreServer.Run(timeout)
}

// RunServer is a convenience helper that combines SetCurrentServer, PrintCurrentServerInfo, and RunCurrentServer.
// The composite key format is: "deploymentName.serverName"
// Example: RunServer("order-service.order-api", 30*time.Second)
func RunServer(compositeKey string, timeout time.Duration) error {
	// Set current server (this validates deployment and server exist)
	if err := SetCurrentServer(compositeKey); err != nil {
		return err
	}

	// Print server info
	if err := PrintCurrentServerInfo(); err != nil {
		return err
	}

	// Run the server
	return RunCurrentServer(timeout)
}
