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
	server, ok := currentDeployment.GetServer(currentServerName)
	if !ok {
		return fmt.Errorf("server '%s' not found in deployment '%s'", currentServerName, currentDeployment.Name())
	}

	// Get apps
	apps := server.Apps()
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
		router := GetRouter(routerName)
		if router == nil {
			return fmt.Errorf("router '%s' not found in registry - did you call RegisterRouter?", routerName)
		}
		routers = append(routers, router)
	}

	// Merge routers if multiple
	var finalRouter router.Router
	if len(routers) == 1 {
		finalRouter = routers[0]
	} else {
		// TODO: Implement router merging for multi-router apps
		return fmt.Errorf("multi-router apps not yet supported (app has %d routers)", len(routers))
	}

	// Create and run Lokstra app
	lokstraApp := app.New(currentServerName, ap.Addr(), finalRouter)
	lokstraApp.PrintStartInfo()

	log.Printf("🟢 Starting server '%s' on %s\n", currentServerName, ap.Addr())
	if err := lokstraApp.Run(timeout); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
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
