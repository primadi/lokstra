package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/primadi/lokstra/core/deploy/loader"
	svc "github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/05-remote-router/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse flags
	serverFlag := flag.String("server", "app.api-server", "Server to run (deployment.server format)")
	flag.Parse()

	// Register service factory
	lokstra_registry.RegisterServiceType("weather-service-factory",
		svc.WeatherServiceFactory)

	// Load config and build deployment topology
	if err := loader.LoadAndBuild([]string{"config.yaml"}); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Print info and run server
	printStartInfo()

	// Run server (compositeKey format: "deployment.server")
	if err := lokstra_registry.RunServer(*serverFlag, 30*time.Second); err != nil {
		log.Fatal(err)
	}
}

func printStartInfo() {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("🌐 Example 07 - Remote Router (proxy.Router)")
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("This example demonstrates:")
	fmt.Println("  ✅ proxy.Router for quick API access")
	fmt.Println("  ✅ No service wrapper needed")
	fmt.Println("  ✅ Direct HTTP calls to external API")
	fmt.Println("  ✅ router-definitions in config")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("📋 Prerequisites:")
	fmt.Println("  1. Start mock weather API first:")
	fmt.Println("     cd mock-weather-api && go run main.go")
	fmt.Println("     (Runs on http://localhost:9001)")
	fmt.Println()
	fmt.Println("  2. Then start this server:")
	fmt.Println("     go run main.go")
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
	fmt.Println("🔗 API Endpoints:")
	fmt.Println()
	fmt.Println("  Weather Reports:")
	fmt.Println("    POST   http://localhost:3001/weather-reports")
	fmt.Println("           ?city=jakarta&forecast=true&days=5")
	fmt.Println()
	fmt.Println("💡 How it works:")
	fmt.Println("  1. GetWeatherReport uses proxy.Router.DoJSON()")
	fmt.Println("  2. Direct HTTP calls to weather API (no wrapper)")
	fmt.Println("  3. Simple and quick integration")
	fmt.Println()
	fmt.Println("📝 Test:")
	fmt.Println("  # Get weather report")
	fmt.Println(`  curl -X POST "http://localhost:3001/weather-reports?city=jakarta&forecast=true&days=3"`)
	fmt.Println()
	fmt.Println("  # Without forecast")
	fmt.Println(`  curl -X POST "http://localhost:3001/weather-reports?city=bandung"`)
	fmt.Println()
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println()
}
