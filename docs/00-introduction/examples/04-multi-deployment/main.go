package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/deploy/loader"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/appservice"
)

func main() {
	// Parse command line flag
	deployment := flag.String("deployment", "order-service", "Deployment to run: monolith, user-service, or order-service")
	flag.Parse()

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════╗\n")
	fmt.Printf("║   LOKSTRA MULTI-DEPLOYMENT DEMO               ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	// 1. Get global registry
	reg := deploy.Global()

	// 2. Register service factories
	reg.RegisterServiceType("database-factory", DatabaseFactory, nil)
	reg.RegisterServiceType("user-service-factory", UserServiceFactory, nil)
	reg.RegisterServiceType("order-service-factory", OrderServiceFactory, nil)

	// 3. Load and build deployment from YAML (services defined in config.yaml)
	dep, err := loader.LoadAndBuild(
		[]string{"config.yaml"},
		*deployment,
		reg,
	)
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	// 5. Run the deployment
	switch *deployment {
	case "monolith":
		fmt.Println("📦 Deployment: MONOLITH")
		fmt.Println("   • All services in one process")
		fmt.Println("   • Port: 3003")
		fmt.Println()
		runMonolith(dep)

	case "user-service":
		fmt.Println("🔷 Deployment: USER-SERVICE")
		fmt.Println("   • Only user endpoints")
		fmt.Println("   • Port: 3004")
		fmt.Println()
		runUserService(dep)

	case "order-service":
		fmt.Println("🔶 Deployment: ORDER-SERVICE")
		fmt.Println("   • Only order endpoints + remote user service")
		fmt.Println("   • Port: 3005")
		fmt.Println()
		runOrderService(dep)

	default:
		log.Fatalf("Unknown deployment: %s\nUse: monolith, user-service, or order-service", *deployment)
	}
}

// ========================================
// MONOLITH: All services in one process
// ========================================

func runMonolith(dep *deploy.Deployment) {
	log.Println("🚀 Starting MONOLITH server")

	// Get app
	server, ok := dep.GetServer("api")
	if !ok {
		log.Fatal("❌ Failed to get server 'api'")
	}
	app := server.Apps()[0]

	// Lazy load services
	userService := service.LazyLoadFrom[appservice.UserService](app, "user-service")
	orderService := service.LazyLoadFrom[appservice.OrderService](app, "order-service")

	// Create handlers
	userHandler := NewUserHandler(userService)
	orderHandler := NewOrderHandler(orderService)

	// Create router
	r := lokstra.NewRouter("monolith")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"server":  "monolith",
			"message": "All services running in one process",
			"endpoints": map[string]any{
				"users": []string{
					"GET /users",
					"GET /users/{id}",
				},
				"orders": []string{
					"GET /orders/{id}",
					"GET /users/{user_id}/orders",
				},
			},
		}
	})

	// Register routes
	r.GET("/users", userHandler.list)
	r.GET("/users/{id}", userHandler.get)
	r.GET("/orders/{id}", orderHandler.get)
	r.GET("/users/{user_id}/orders", orderHandler.getUserOrders)

	// Run
	lokstraApp := lokstra.NewApp("monolith", ":3003", r)
	lokstraApp.PrintStartInfo()
	if err := lokstraApp.Run(30 * time.Second); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

// ========================================
// USER SERVICE: Only user endpoints
// ========================================

func runUserService(dep *deploy.Deployment) {
	log.Println("🚀 Starting USER-SERVICE server")

	// Get app
	server, ok := dep.GetServer("user-api")
	if !ok {
		log.Fatal("❌ Failed to get server 'user-api'")
	}
	app := server.Apps()[0]

	// Lazy load services
	userService := service.LazyLoadFrom[appservice.UserService](app, "user-service")

	// Create handlers
	userHandler := NewUserHandler(userService)

	// Create router
	r := lokstra.NewRouter("user-service")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"server": "user-service",
			"endpoints": []string{
				"GET /users",
				"GET /users/{id}",
			},
		}
	})

	// Register routes
	r.GET("/users", userHandler.list)
	r.GET("/users/{id}", userHandler.get)

	// Run
	lokstraApp := lokstra.NewApp("user-service", ":3004", r)
	lokstraApp.PrintStartInfo()
	if err := lokstraApp.Run(30 * time.Second); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}

// ========================================
// ORDER SERVICE: Order endpoints + remote user service
// ========================================

func runOrderService(dep *deploy.Deployment) {
	log.Println("🚀 Starting ORDER-SERVICE server")

	// Get app
	server, ok := dep.GetServer("order-api")
	if !ok {
		log.Fatal("❌ Failed to get server 'order-api'")
	}
	app := server.Apps()[0]

	// Lazy load services
	orderService := service.LazyLoadFrom[appservice.OrderService](app, "order-service")

	// Create handlers
	orderHandler := NewOrderHandler(orderService)

	// Create router
	r := lokstra.NewRouter("order-service")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"server": "order-service",
			"endpoints": []string{
				"GET /orders/{id}",
				"GET /users/{user_id}/orders",
			},
			"dependencies": []string{
				"user-service (remote at http://localhost:3004)",
			},
		}
	})

	// Register routes
	r.GET("/orders/{id}", orderHandler.get)
	r.GET("/users/{user_id}/orders", orderHandler.getUserOrders)

	// Run
	lokstraApp := lokstra.NewApp("order-service", ":3005", r)
	lokstraApp.PrintStartInfo()
	if err := lokstraApp.Run(30 * time.Second); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
