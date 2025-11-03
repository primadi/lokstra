package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/middleware/cors"
)

func main() {
	router := lokstra.NewRouter("api")

	// Configure CORS middleware - allow specific origins
	allowedOrigins := []string{
		"http://localhost:3001",
		"http://localhost:8080",
		"https://myapp.com",
	}

	// Add CORS middleware globally
	router.Use(cors.Middleware(allowedOrigins))

	// Register routes
	router.GET("/users", func() map[string]any {
		return map[string]any{
			"users": []string{"Alice", "Bob", "Charlie"},
		}
	})

	router.POST("/users", func() map[string]any {
		return map[string]any{
			"message": "User created",
			"id":      123,
		}
	})

	router.PUT("/users/1", func() map[string]any {
		return map[string]any{
			"message": "User updated",
			"id":      1,
		}
	})

	router.DELETE("/users/1", func() map[string]any {
		return map[string]any{
			"message": "User deleted",
			"id":      1,
		}
	})

	// Create app
	app := lokstra.NewApp("cors-demo", ":3000", router)

	log.Println("🚀 CORS Middleware Demo")
	log.Println("🌐 Demonstrates CORS configuration")
	log.Println()
	log.Println("CORS Configuration:")
	log.Println("  Allowed Origins:")
	log.Println("    - http://localhost:3001")
	log.Println("    - http://localhost:8080")
	log.Println("    - https://myapp.com")
	log.Println()
	log.Println("  Allowed Methods: GET, POST, PUT, DELETE, OPTIONS")
	log.Println("  Allowed Headers: Content-Type, Authorization, X-API-Key")
	log.Println("  Credentials: Allowed")
	log.Println()
	log.Println("Endpoints:")
	log.Println("  GET    /users")
	log.Println("  POST   /users")
	log.Println("  PUT    /users/1")
	log.Println("  DELETE /users/1")
	log.Println()
	log.Println("Test with browser console or curl with Origin header:")
	log.Println(`  curl -H "Origin: http://localhost:3001" http://localhost:3000/users`)
	log.Println()
	log.Println("Server: http://localhost:3000")

	// Run
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
