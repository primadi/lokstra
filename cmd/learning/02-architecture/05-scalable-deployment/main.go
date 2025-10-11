package main

import (
	"fmt"
	"os"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	fmt.Println("🚀 Lokstra Scalable Deployment Demo")
	fmt.Println("=====================================")

	// Get deployment mode from command line arg
	deploymentMode := "monolith-single"
	if len(os.Args) > 1 {
		deploymentMode = os.Args[1]
	}

	configFile := getConfigFile(deploymentMode)
	fmt.Printf("📋 Deployment Mode: %s\n", deploymentMode)
	fmt.Printf("📄 Config File: %s\n\n", configFile)

	// Register all service factories and routers
	setupFactories()
	setupRouters()

	// Load configuration
	cfg := config.New()
	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		return
	}

	lokstra_registry.RegisterConfig(cfg, "")

	// Print deployment info
	printDeploymentInfo(deploymentMode)

	// Start server
	fmt.Println("\n🌐 Starting Server...")
	fmt.Println("=====================================")
	lokstra_registry.PrintServerStartInfo()
	if err := lokstra_registry.StartServer(); err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}

func getConfigFile(mode string) string {
	switch mode {
	case "monolith-single":
		return "monolith-single.yaml"
	case "monolith-multi":
		return "monolith-multi.yaml"
	case "product-service":
		return "product-service.yaml"
	case "order-service":
		return "order-service.yaml"
	case "gateway":
		return "gateway.yaml"
	case "hybrid":
		return "hybrid.yaml"
	default:
		return "monolith-single.yaml"
	}
}

func printDeploymentInfo(mode string) {
	fmt.Println("\n📊 Deployment Information")
	fmt.Println("=====================================")

	switch mode {
	case "monolith-single":
		fmt.Println("✅ Type: Monolith Single Port")
		fmt.Println("   All services on one port (:8080)")
		fmt.Println("   Router calls: LOCAL (zero overhead)")
		fmt.Println("   Best for: Development, testing")

	case "monolith-multi":
		fmt.Println("✅ Type: Monolith Multi Port")
		fmt.Println("   Product API: :8081")
		fmt.Println("   Order API: :8082")
		fmt.Println("   Router calls: HTTP to localhost")
		fmt.Println("   Best for: Staging, load testing")

	case "product-service":
		fmt.Println("✅ Type: Microservice (Product)")
		fmt.Println("   Runs on: :8081")
		fmt.Println("   Routers: product-api, health-api")
		fmt.Println("   Calls order-api at: :8082")

	case "order-service":
		fmt.Println("✅ Type: Microservice (Order)")
		fmt.Println("   Runs on: :8082")
		fmt.Println("   Routers: order-api, health-api")
		fmt.Println("   Calls product-api at: :8081")

	case "gateway":
		fmt.Println("✅ Type: API Gateway")
		fmt.Println("   Runs on: :8080")
		fmt.Println("   Routes to: product-api (:8081), order-api (:8082)")
		fmt.Println("   No local routers")

	case "hybrid":
		fmt.Println("✅ Type: Hybrid (Public + Private)")
		fmt.Println("   Public API: :8080 (exposed)")
		fmt.Println("   Private APIs: :8081, :8082 (internal)")
		fmt.Println("   Best for: Production security")
	}
}
