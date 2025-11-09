package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
)

func main() {
	// API app on port 8081
	apiRouter := lokstra.NewRouter("api")
	apiRouter.GET("/health", func() map[string]any {
		return map[string]any{"status": "ok"}
	})
	apiRouter.GET("/users", func() map[string]any {
		return map[string]any{
			"users": []string{"Alice", "Bob", "Charlie"},
		}
	})
	apiApp := lokstra.NewApp("api-app", ":8081", apiRouter)

	// Admin app on port 8082
	adminRouter := lokstra.NewRouter("admin")
	adminRouter.GET("/dashboard", func() map[string]any {
		return map[string]any{"dashboard": "admin"}
	})
	adminRouter.GET("/users", func() map[string]any {
		return map[string]any{
			"users":   []string{"Alice", "Bob"},
			"actions": []string{"edit", "delete"},
		}
	})
	adminApp := lokstra.NewApp("admin-app", ":8082", adminRouter)

	// Metrics app on port 8083
	metricsRouter := lokstra.NewRouter("metrics")
	metricsRouter.GET("/metrics", func() map[string]any {
		return map[string]any{
			"cpu":      "25%",
			"memory":   "512MB",
			"requests": 1234,
		}
	})
	metricsApp := lokstra.NewApp("metrics-app", ":8083", metricsRouter)

	// Server manages all apps
	server := lokstra.NewServer("main-server", apiApp, adminApp, metricsApp)

	log.Println("🚀 Server starting all apps...")
	log.Println("  API:     http://localhost:8081")
	log.Println("  Admin:   http://localhost:8082")
	log.Println("  Metrics: http://localhost:8083")
	log.Println()
	log.Println("🛑 Press Ctrl+C to stop all")

	// All apps start together, shutdown together
	if err := server.Run(30 * time.Second); err != nil {
		log.Fatal(err)
	}
}
