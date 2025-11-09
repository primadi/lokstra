package main

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/request_logger"
)

func main() {
	// Create router
	router := lokstra.NewRouter("api")

	// Add logging middleware globally
	// This will log all requests to this router
	router.Use(request_logger.Middleware(nil))

	// Register routes
	router.GET("/users", func() map[string]any {
		return map[string]any{
			"users": []string{"Alice", "Bob", "Charlie"},
		}
	})

	router.GET("/products", func() map[string]any {
		return map[string]any{
			"products": []string{"Laptop", "Phone", "Tablet"},
		}
	})

	router.POST("/users", func() map[string]any {
		return map[string]any{
			"message": "User created",
			"id":      123,
		}
	})

	router.GET("/error", func() (map[string]any, error) {
		return nil, fmt.Errorf("something went wrong")
	})

	// Create app
	app := lokstra.NewApp("logging-demo", ":3000", router)

	log.Println("🚀 Logging Middleware Demo")
	log.Println("📝 All requests will be logged automatically")
	log.Println()
	log.Println("Try these endpoints:")
	log.Println("  GET  /users")
	log.Println("  GET  /products")
	log.Println("  POST /users")
	log.Println("  GET  /error     (will log error)")
	log.Println()
	log.Println("Watch the console for request logs!")
	log.Println()
	log.Println("Server: http://localhost:3000")

	// Run
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
