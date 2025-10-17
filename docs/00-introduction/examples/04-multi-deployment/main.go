package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
)

// ========================================
// Server Configurations
// ========================================

func runMonolithServer() {
	log.Println("🚀 Starting MONOLITH server")
	log.Println("   All services in one process")

	// Register services
	registerMonolithServices()

	// Create combined router
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

	// Register user routes
	r.GET("/users", listUsersHandler)
	r.GET("/users/{id}", getUserHandler)

	// Register order routes
	r.GET("/orders/{id}", getOrderHandler)
	r.GET("/users/{user_id}/orders", getUserOrdersHandler)

	// Run
	app := lokstra.NewApp("monolith", ":3003", r)
	app.PrintStartInfo()
	app.Run(30 * time.Second)
}

func runUserServiceServer() {
	log.Println("🚀 Starting USER-SERVICE server")
	log.Println("   Only user-related endpoints")

	// Register user service dependencies
	registerUserServices()

	// Create user router
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

	// Register user routes
	r.GET("/users", listUsersHandler)
	r.GET("/users/{id}", getUserHandler)

	// Run
	app := lokstra.NewApp("user-service", ":3004", r)
	app.PrintStartInfo()
	app.Run(30 * time.Second)
}

func runOrderServiceServer() {
	log.Println("🚀 Starting ORDER-SERVICE server")
	log.Println("   Only order-related endpoints")

	// Register order service dependencies
	registerOrderServices()

	// Create order router
	r := lokstra.NewRouter("order-service")

	r.GET("/", func() map[string]any {
		return map[string]any{
			"server": "order-service",
			"endpoints": []string{
				"GET /orders/{id}",
				"GET /users/{user_id}/orders",
			},
			"dependencies": []string{
				"user-service (for user data)",
			},
		}
	})

	// Register order routes
	r.GET("/orders/{id}", getOrderHandler)
	r.GET("/users/{user_id}/orders", getUserOrdersHandler)

	// Run
	app := lokstra.NewApp("order-service", ":3005", r)
	app.PrintStartInfo()
	app.Run(30 * time.Second)
}

// ========================================
// Main
// ========================================

func main() {
	server := flag.String("server", "order-service", "Server to run: monolith, user-service, or order-service")
	flag.Parse()

	fmt.Printf("\n")
	fmt.Printf("╔═══════════════════════════════════════════════╗\n")
	fmt.Printf("║   LOKSTRA MULTI-DEPLOYMENT DEMO               ║\n")
	fmt.Printf("╚═══════════════════════════════════════════════╝\n")
	fmt.Printf("\n")

	switch *server {
	case "monolith":
		fmt.Println("📦 Server: MONOLITH")
		fmt.Println("   • All services in one process")
		fmt.Println("   • Port: 3003")
		fmt.Println()
		runMonolithServer()

	case "user-service":
		fmt.Println("🔷 Server: USER-SERVICE")
		fmt.Println("   • Only user endpoints")
		fmt.Println("   • Port: 3004")
		fmt.Println()
		runUserServiceServer()

	case "order-service":
		fmt.Println("🔶 Server: ORDER-SERVICE")
		fmt.Println("   • Only order endpoints")
		fmt.Println("   • Port: 3005")
		fmt.Println()
		runOrderServiceServer()

	default:
		log.Fatalf("Unknown server: %s\nUse: monolith, user-service, or order-service", *server)
	}
}
