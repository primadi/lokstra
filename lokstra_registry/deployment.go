package lokstra_registry

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/core/server"
)

var (
	currentDeployment *deploy.Deployment
	currentServerName string
)

// LoadAndBuild loads config and builds ALL deployments into Global registry
func LoadAndBuild(configPaths []string) error {
	return loader.LoadAndBuild(configPaths)
}

// SetCurrentServer sets the current server using composite key: "deploymentName.serverName"
// Example: SetCurrentServer("order-service.order-api")
func SetCurrentServer(compositeKey string) error {
	parts := strings.Split(compositeKey, ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid server key format, expected 'deployment.server', got: %s", compositeKey)
	}

	deploymentName := parts[0]
	serverName := parts[1]

	// Find deployment from global registry
	dep, ok := deploy.Global().GetDeployment(deploymentName)
	if !ok {
		return fmt.Errorf("deployment not found: %s", deploymentName)
	}

	// Find server
	_, ok = dep.GetServer(serverName)
	if !ok {
		return fmt.Errorf("server '%s' not found in deployment '%s'", serverName, deploymentName)
	}

	// Set current context
	currentDeployment = dep
	currentServerName = serverName
	return nil
}

// GetCurrentDeployment returns the current deployment
func GetCurrentDeployment() *deploy.Deployment {
	return currentDeployment
}

// GetCurrentServerName returns the current server name
func GetCurrentServerName() string {
	return currentServerName
}

// PrintCurrentServerInfo prints information about the current server configuration
func PrintCurrentServerInfo() error {
	if currentDeployment == nil {
		return fmt.Errorf("no deployment set - call SetCurrentServer first")
	}
	if currentServerName == "" {
		return fmt.Errorf("no server name set - call SetCurrentServer first")
	}

	server, ok := currentDeployment.GetServer(currentServerName)
	if !ok {
		return fmt.Errorf("server '%s' not found in deployment '%s'", currentServerName, currentDeployment.Name())
	}

	fmt.Println("┌─────────────────────────────────────────────┐")
	fmt.Printf("│ Server: %-35s │\n", currentServerName)
	fmt.Println("├─────────────────────────────────────────────┤")
	fmt.Printf("│ Deployment: %-31s │\n", currentDeployment.Name())
	fmt.Printf("│ Base URL: %-33s │\n", server.BaseURL())

	// Apps
	apps := server.Apps()
	if len(apps) > 0 {
		fmt.Println("│                                             │")
		fmt.Printf("│ Apps: %-37d │\n", len(apps))
		for i, app := range apps {
			fmt.Printf("│   App #%d:                                   │\n", i+1)
			fmt.Printf("│     Addr: %-33s │\n", app.Addr())

			// App services
			appServices := app.Services()
			if len(appServices) > 0 {
				fmt.Println("│     Services:                               │")
				for svcName := range appServices {
					fmt.Printf("│       • %-35s │\n", svcName)
				}
			}

			// App remote services
			appRemoteServices := app.RemoteServices()
			if len(appRemoteServices) > 0 {
				fmt.Println("│     Remote Services:                        │")
				for svcName := range appRemoteServices {
					fmt.Printf("│       • %-32s │\n", svcName)
				}
			}

			// Routers
			routersMap := app.Routers()
			if len(routersMap) > 0 {
				fmt.Println("│     Routers:                                │")
				for routerName := range routersMap {
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
	if currentDeployment == nil {
		return fmt.Errorf("no deployment set - call SetCurrentServer first")
	}
	if currentServerName == "" {
		return fmt.Errorf("no server name set - call SetCurrentServer first")
	}

	// Get server from deployment
	deployServer, ok := currentDeployment.GetServer(currentServerName)
	if !ok {
		return fmt.Errorf("server '%s' not found in deployment '%s'", currentServerName, currentDeployment.Name())
	}

	// Get apps
	apps := deployServer.Apps()
	if len(apps) == 0 {
		return fmt.Errorf("server '%s' has no apps configured", currentServerName)
	}

	// For now, support single-app servers only
	if len(apps) > 1 {
		return fmt.Errorf("multi-app servers not yet supported (server '%s' has %d apps)", currentServerName, len(apps))
	}

	ap := apps[0]

	// Build routers from registry
	routersMap := ap.Routers()
	if len(routersMap) == 0 {
		return fmt.Errorf("app has no routers configured")
	}

	var routers []router.Router
	for routerName := range routersMap {
		// Try to get manually registered router first
		r := GetRouter(routerName)

		// If not found, try to build from router definition (auto-generated)
		if r == nil {
			autoRouter, err := BuildRouterFromDefinition(routerName, ap)
			if err != nil {
				return fmt.Errorf("router '%s' not found in registry and failed to auto-build: %w", routerName, err)
			}
			r = autoRouter
			log.Printf("✨ Auto-generated router '%s' from service '%s'\n", routerName, deploy.Global().GetRouterDef(routerName).Service)
		}

		routers = append(routers, r)
	}

	// Create Lokstra App with all routers (app.New supports multiple routers)
	lokstraApp := app.New(currentServerName, ap.Addr(), routers...)

	// Create core Server and run (delegates to core/server/server.go)
	coreServer := server.New(currentServerName, lokstraApp)
	coreServer.PrintStartInfo()

	// Delegate to core Server.Run() - no code duplication!
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
