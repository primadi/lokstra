package main

import (
	"flag"
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/core/deploy/schema"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

func main() {
	// Parse command line flag
	mode := flag.String("mode", "config", "Run mode: 'code' (manual) or 'config' (YAML)")
	flag.Parse()

	log.Printf("🚀 Starting CRUD API in '%s' mode...\n", *mode)

	if *mode == "config" {
		runWithConfig()
	} else {
		runWithCode()
	}
}

// ========================================
// APPROACH 1: Run by Code (Manual)
// ========================================

func runWithCode() {
	log.Println("📝 APPROACH 1: Manual registration + Lazy loading (run by code)")

	// 1. Register service factories
	lokstra_registry.RegisterServiceType("database-factory", DatabaseFactory, nil)
	lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

	// 2. Define services in registry using ServiceDef (like YAML structure)
	lokstra_registry.DefineService(&schema.ServiceDef{
		Name: "database",
		Type: "database-factory",
	})
	lokstra_registry.DefineService(&schema.ServiceDef{
		Name:      "user-service",
		Type:      "user-service-factory",
		DependsOn: []string{"database"},
	})

	// 3. Build deployment manually (same structure as config mode)
	dep := deploy.New("development")
	server := dep.NewServer("api", "http://localhost")
	app := server.NewApp(":3002")

	// 4. Add services to app (lazy-loaded automatically)
	app.AddService("database")
	app.AddService("user-service")

	// 5. Lazy load service (SAME pattern as config mode)
	userService := service.LazyLoadFrom[*UserService](app, "user-service")

	log.Println("✅ Services configured from code (lazy - will be created on first HTTP request)")

	// 6. Create handler with injected service
	handler := NewUserHandler(userService)

	// 7. Setup router and run
	setupRouterAndRun(handler)
}

// ========================================
// APPROACH 2: Run by Config (YAML + Lazy DI)
// ========================================

func runWithConfig() {
	log.Println("⚙️  APPROACH 2: YAML Configuration + Lazy DI (run by config)")

	// 1. Register service factories
	lokstra_registry.RegisterServiceType("database-factory", DatabaseFactory, nil)
	lokstra_registry.RegisterServiceType("user-service-factory", UserServiceFactory, nil)

	// 2. Load and build deployment from YAML
	if err := loader.LoadAndBuild([]string{"config.yaml"}); err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// 3. Get app (service container) from global registry
	dep, ok := deploy.Global().GetDeployment("development")
	if !ok {
		log.Fatal("❌ Failed to get deployment 'development'")
	}
	server, ok := dep.GetServer("api")
	if !ok {
		log.Fatal("❌ Failed to get server 'api'")
	}
	app := server.Apps()[0]

	// Debug: Print available services
	log.Printf("📋 App has %d services", len(app.Services()))
	for _, svc := range app.Services() {
		log.Printf("   - %s", svc.Name())
	}

	// 4. Lazy load service from app - Service created on FIRST HTTP request!
	userService := service.LazyLoadFrom[*UserService](app, "user-service")

	log.Println("✅ Services configured from YAML (lazy - will be created on first HTTP request)")

	// 5. Create handler with injected service
	handler := NewUserHandler(userService)

	// 6. Setup router and run
	setupRouterAndRun(handler)
}

// ========================================
// Router Setup (Shared by Both Approaches)
// ========================================

func setupRouterAndRun(handler *UserHandler) {
	// Create router
	r := lokstra.NewRouter("api")

	// CRUD routes - Using Group for clean organization
	r.Group("/api/v1/users", func(api lokstra.Router) {
		api.GET("/", handler.listUsers)
		api.GET("/{id}", handler.getUser)
		api.POST("/", handler.createUser)
		api.PUT("/{id}", handler.updateUser)
		api.DELETE("/{id}", handler.deleteUser)
	})

	// Info endpoint
	r.GET("/", func() map[string]any {
		return map[string]any{
			"service": "User CRUD API",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"GET /api/v1/users":         "List all users",
				"GET /api/v1/users/{id}":    "Get user by ID",
				"POST /api/v1/users":        "Create user",
				"PUT /api/v1/users/{id}":    "Update user",
				"DELETE /api/v1/users/{id}": "Delete user",
			},
		}
	})

	// Create and start app
	app := lokstra.NewApp("crud-api", ":3002", r)
	app.PrintStartInfo()

	// Handle error from Run (e.g., port already in use)
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
