package main

import (
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra/core/config"
)

func main() {
	fmt.Println("🔧 Lokstra YAML Config Demo")
	fmt.Println("============================")

	// Example 1: Load single config file
	fmt.Println("\n1. Loading single config file...")
	var cfg1 config.Config

	// Load the example config file
	if err := config.LoadConfigFile("core/config/example.yaml", &cfg1); err != nil {
		log.Printf("Failed to load config file: %v", err)
	} else {
		fmt.Printf("✅ Loaded %d routers, %d services, %d middlewares, %d servers\n",
			len(cfg1.Routers), len(cfg1.Services), len(cfg1.Middlewares), len(cfg1.Servers))

		// Show some details
		if len(cfg1.Routers) > 0 {
			router := cfg1.Routers[0]
			fmt.Printf("   First router: %s (enabled: %v, routes: %d)\n",
				router.Name, router.IsEnabled(), len(router.Routes))
		}
	}

	// Example 2: Create multi-file config structure
	fmt.Println("\n2. Creating multi-file config structure...")

	// Create temporary directory for multi-file config
	tmpDir := "./temp_config"
	os.MkdirAll(tmpDir, 0755)
	defer os.RemoveAll(tmpDir)

	// Create separate config files
	routersYAML := `
routers:
  - name: api-router
    use: [cors-mw, auth-mw]
    routes:
      - name: health-check
        path: /health
        handler: HealthHandler
      - name: users-list
        path: /users
        method: GET
        handler: GetUsersHandler
`

	servicesYAML := `
services:
  - name: database
    type: postgres
    config:
      dsn: "postgres://localhost/myapp"
  - name: cache
    type: redis
    config:
      addr: "localhost:6379"
`

	middlewaresYAML := `
middlewares:
  - name: cors-mw
    type: cors
    config:
      allowed_origins: ["*"]
  - name: auth-mw
    type: jwt
    config:
      secret: "my-secret"
`

	serversYAML := `
servers:
  - name: web-server
    description: "Main web server"
    services: [database, cache]
    apps:
      - name: api-app
        addr: ":8080"
        routers: [api-router]
`

	// Write files
	files := map[string]string{
		"routers.yaml":     routersYAML,
		"services.yaml":    servicesYAML,
		"middlewares.yaml": middlewaresYAML,
		"servers.yaml":     serversYAML,
	}

	for filename, content := range files {
		filepath := tmpDir + "/" + filename
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			log.Printf("Failed to write %s: %v", filename, err)
			return
		}
	}

	// Example 3: Load multi-file config
	fmt.Println("\n3. Loading multi-file config directory...")
	var cfg2 config.Config

	if err := config.LoadConfigDir(tmpDir, &cfg2); err != nil {
		log.Printf("Failed to load config directory: %v", err)
	} else {
		fmt.Printf("✅ Loaded %d routers, %d services, %d middlewares, %d servers\n",
			len(cfg2.Routers), len(cfg2.Services), len(cfg2.Middlewares), len(cfg2.Servers))

		// Show configuration details
		fmt.Println("\n   Configuration Summary:")
		for _, router := range cfg2.Routers {
			fmt.Printf("   • Router: %s (%d routes, uses: %v)\n",
				router.Name, len(router.Routes), router.Use)
		}

		for _, service := range cfg2.Services {
			fmt.Printf("   • Service: %s (type: %s, enabled: %v)\n",
				service.Name, service.Type, service.IsEnabled())
		}

		for _, middleware := range cfg2.Middlewares {
			fmt.Printf("   • Middleware: %s (type: %s)\n",
				middleware.Name, middleware.Type)
		}

		for _, server := range cfg2.Servers {
			fmt.Printf("   • Server: %s (%d apps, services: %v)\n",
				server.Name, len(server.Apps), server.Services)
		}
	}

	// Example 4: Apply configurations
	fmt.Println("\n4. Applying configurations...")

	// Apply specific components
	if err := config.ApplyMiddlewareConfig(&cfg2, "cors-mw", "auth-mw"); err != nil {
		fmt.Printf("❌ Failed to apply middlewares: %v\n", err)
	} else {
		fmt.Println("✅ Applied middlewares: cors-mw, auth-mw")
	}

	if err := config.ApplyServicesConfig(&cfg2, "database"); err != nil {
		fmt.Printf("❌ Failed to apply services: %v\n", err)
	} else {
		fmt.Println("✅ Applied services: database")
	}

	if err := config.ApplyRoutersConfig(&cfg2, "api-router"); err != nil {
		fmt.Printf("❌ Failed to apply routers: %v\n", err)
	} else {
		fmt.Println("✅ Applied routers: api-router")
	}

	if err := config.ApplyServerConfig(&cfg2, "web-server"); err != nil {
		fmt.Printf("❌ Failed to apply server config: %v\n", err)
	} else {
		fmt.Println("✅ Applied server config: web-server")
	}

	// Apply all configuration at once
	fmt.Println("\n5. Applying complete configuration...")
	if err := config.ApplyAllConfig(&cfg2, "web-server"); err != nil {
		fmt.Printf("❌ Failed to apply all config: %v\n", err)
	} else {
		fmt.Println("✅ Successfully applied complete configuration for web-server")
	}

	// Example 6: Demonstrate default values
	fmt.Println("\n6. Demonstrating default values...")
	router := config.Router{Name: "demo-router"}
	fmt.Printf("   Router enabled by default: %v\n", router.IsEnabled())
	fmt.Printf("   Router engine type default: %s\n", router.GetEngineType())

	service := config.Service{Name: "demo-service", Type: "memory"}
	fmt.Printf("   Service enabled by default: %v\n", service.IsEnabled())

	fmt.Println("\n🎉 Demo completed!")
	fmt.Println("\nNext steps:")
	fmt.Println("• Create your own YAML config files")
	fmt.Println("• Integrate with Lokstra registry system")
	fmt.Println("• Use config.ApplyAllConfig() in your main.go")
	fmt.Println("• See docs/yaml-config.md for detailed documentation")
}
