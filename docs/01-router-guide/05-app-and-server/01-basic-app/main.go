package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	// Create API router
	apiRouter := lokstra.NewRouter("api")
	apiRouter.GET("/users", func() map[string]any {
		return map[string]any{
			"users": []string{"Alice", "Bob"},
		}
	})
	apiRouter.GET("/products", func() map[string]any {
		return map[string]any{
			"products": []string{"Book", "Pen"},
		}
	})

	// Create admin router
	adminRouter := lokstra.NewRouter("admin")
	adminRouter.GET("/stats", func() map[string]any {
		return map[string]any{
			"requests": 1234,
			"uptime":   "2h",
		}
	})
	adminRouter.GET("/logs", func() map[string]any {
		return map[string]any{
			"logs": []string{"Log 1", "Log 2"},
		}
	})

	// Combine into one app
	app := lokstra.NewApp("web-app", ":8080", apiRouter, adminRouter)

	log.Println("🚀 Server starting on :8080")
	log.Println("📋 Endpoints:")
	log.Println("  GET /users")
	log.Println("  GET /products")
	log.Println("  GET /stats")
	log.Println("  GET /logs")
	log.Println()
	log.Println("🛑 Press Ctrl+C to stop")

	// Run with 30s graceful shutdown timeout
	if err := app.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
