package main

import (
	"fmt"
	"os"

	"github.com/primadi/lokstra/core/config"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test-config-loading.go <config-file>")
		fmt.Println("Example:")
		fmt.Println("  go run test-config-loading.go test-simple-services.yaml")
		fmt.Println("  go run test-config-loading.go test-layered-services.yaml")
		os.Exit(1)
	}

	configFile := os.Args[1]
	fmt.Printf("Loading config from: %s\n\n", configFile)

	// Load config
	cfg := &config.Config{}
	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Config loaded successfully!")
	fmt.Println()

	// Debug: check what we actually got
	fmt.Printf("DEBUG: Simple=%v, Layered=%v, Order=%v\n",
		len(cfg.Services.Simple),
		len(cfg.Services.Layered),
		len(cfg.Services.Order))
	fmt.Println()

	// Check service mode
	if cfg.Services.IsSimple() {
		fmt.Println("📋 Service Mode: SIMPLE (array)")
		fmt.Printf("   Services: %d\n", len(cfg.Services.Simple))
		for i, svc := range cfg.Services.Simple {
			fmt.Printf("   %d. %s (type: %s, enabled: %v)\n", i+1, svc.Name, svc.Type, svc.IsEnabled())
			if len(svc.Config) > 0 {
				fmt.Printf("      Config: %v\n", svc.Config)
			}
		}
	} else if cfg.Services.IsLayered() {
		fmt.Println("📋 Service Mode: LAYERED (map)")
		fmt.Printf("   Layers: %d\n", len(cfg.Services.Order))
		for _, layerName := range cfg.Services.Order {
			services := cfg.Services.Layered[layerName]
			fmt.Printf("\n   Layer: %s (%d services)\n", layerName, len(services))
			for i, svc := range services {
				fmt.Printf("   %d. %s (type: %s, enabled: %v)\n", i+1, svc.Name, svc.Type, svc.IsEnabled())
				if len(svc.DependsOn) > 0 {
					fmt.Printf("      Depends on: %v\n", svc.DependsOn)
				}
				if len(svc.Config) > 0 {
					fmt.Printf("      Config: %v\n", svc.Config)
				}
			}
		}

		// Validate layered services
		fmt.Println("\n🔍 Validating layered services...")
		if err := config.ValidateLayeredServices(&cfg.Services); err != nil {
			fmt.Printf("   ❌ Validation failed: %v\n", err)
		} else {
			fmt.Println("   ✅ Validation passed!")
		}
	}

	fmt.Println()
	fmt.Printf("️  Servers: %d\n", len(cfg.Servers))
	fmt.Println()
	fmt.Println("✅ All checks passed!")
}
