package main

import (
	"fmt"
	"os"

	"github.com/primadi/lokstra/core/deploy/loader"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <config.yaml>")
		os.Exit(1)
	}

	config, err := loader.LoadConfig(os.Args[1])
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Config loaded successfully!\n")
	fmt.Printf("   Service definitions: %d\n", len(config.ServiceDefinitions))
	fmt.Printf("   External service definitions: %d\n", len(config.ExternalServiceDefinitions))
	fmt.Printf("   Deployments: %d\n", len(config.Deployments))

	// Show deployment details
	for depName, dep := range config.Deployments {
		fmt.Printf("\n   📦 Deployment: %s\n", depName)
		for serverName, server := range dep.Servers {
			fmt.Printf("      🖥️  Server: %s (%s)\n", serverName, server.BaseURL)
			for i, app := range server.Apps {
				fmt.Printf("         App #%d: %s\n", i+1, app.Addr)
				if len(app.Routers) > 0 {
					fmt.Printf("            Routers: %v\n", app.Routers)
				}
				if len(app.PublishedServices) > 0 {
					fmt.Printf("            App published services: %v\n", app.PublishedServices)
				}
			}
		}
	}
}
