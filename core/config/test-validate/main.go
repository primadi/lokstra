package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/primadi/lokstra/core/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <yaml-file>")
		os.Exit(1)
	}

	yamlFile := os.Args[1]

	fmt.Printf("Loading and validating: %s\n", yamlFile)
	fmt.Println(strings.Repeat("=", 60))

	cfg := config.New()
	err := config.LoadConfigFile(yamlFile, cfg)

	if err != nil {
		fmt.Println("❌ Validation FAILED:")
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("✅ Validation PASSED!")
	fmt.Println()

	// Print summary
	fmt.Printf("Summary:\n")
	fmt.Printf("  - Configs: %d\n", len(cfg.Configs))
	fmt.Printf("  - Services: %d\n", len(cfg.Services.GetAllServices()))
	fmt.Printf("  - Middlewares: %d\n", len(cfg.Middlewares))
	fmt.Printf("  - Servers: %d\n", len(cfg.Servers))

	if len(cfg.Servers) > 0 {
		fmt.Println("\nServers:")
		for _, srv := range cfg.Servers {
			fmt.Printf("  - %s (%s)\n", srv.Name, srv.GetBaseUrl())
			for _, app := range srv.Apps {
				fmt.Printf("    - App: %s @ %s\n", app.Name, app.Addr)
			}
		}
	}
}
