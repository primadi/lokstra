package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/primadi/lokstra"
)

// Multiple servers example

// API handlers
func APIHomeHandler() string {
	return `
	<html>
	<body>
		<h1>API Server</h1>
		<p>Port: 3080</p>
		<ul>
			<li><a href="/api/users">Users</a></li>
			<li><a href="/api/orders">Orders</a></li>
		</ul>
	</body>
	</html>
	`
}

func GetUsersHandler() map[string]any {
	return map[string]any{
		"users": []map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		},
		"server": "api-server",
	}
}

func GetOrdersHandler() map[string]any {
	return map[string]any{
		"orders": []map[string]any{
			{"id": 1, "item": "Book"},
			{"id": 2, "item": "Laptop"},
		},
		"server": "api-server",
	}
}

// Admin handlers
func AdminHomeHandler() string {
	return `
	<html>
	<body>
		<h1>Admin Server</h1>
		<p>Port: 3081</p>
		<ul>
			<li><a href="/admin/stats">Stats</a></li>
			<li><a href="/admin/config">Config</a></li>
		</ul>
	</body>
	</html>
	`
}

func GetStatsHandler() map[string]any {
	return map[string]any{
		"total_users":  150,
		"total_orders": 75,
		"uptime":       "24h",
		"server":       "admin-server",
	}
}

func GetConfigHandler() map[string]any {
	return map[string]any{
		"environment": "development",
		"log_level":   "DEBUG",
		"server":      "admin-server",
	}
}

// Metrics handlers
func MetricsHomeHandler() string {
	return `
	<html>
	<body>
		<h1>Metrics Server</h1>
		<p>Port: 3082</p>
		<ul>
			<li><a href="/metrics">Metrics</a></li>
			<li><a href="/health">Health</a></li>
		</ul>
	</body>
	</html>
	`
}

func GetMetricsHandler() map[string]any {
	return map[string]any{
		"requests_total": 1234,
		"errors_total":   5,
		"response_time":  "25ms",
		"server":         "metrics-server",
	}
}

func GetHealthHandler() map[string]any {
	return map[string]any{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"server":    "metrics-server",
	}
}

func main() {
	var wg sync.WaitGroup

	// API Server (port 3080)
	apiRouter := lokstra.NewRouter("api")
	apiRouter.GET("/", APIHomeHandler)
	apiRouter.GET("/api/users", GetUsersHandler)
	apiRouter.GET("/api/orders", GetOrdersHandler)
	apiApp := lokstra.NewApp("api-server", ":3080", apiRouter)

	// Admin Server (port 3081)
	adminRouter := lokstra.NewRouter("admin")
	adminRouter.GET("/", AdminHomeHandler)
	adminRouter.GET("/admin/stats", GetStatsHandler)
	adminRouter.GET("/admin/config", GetConfigHandler)
	adminApp := lokstra.NewApp("admin-server", ":3081", adminRouter)

	// Metrics Server (port 3082)
	metricsRouter := lokstra.NewRouter("metrics")
	metricsRouter.GET("/", MetricsHomeHandler)
	metricsRouter.GET("/metrics", GetMetricsHandler)
	metricsRouter.GET("/health", GetHealthHandler)
	metricsApp := lokstra.NewApp("metrics-server", ":3082", metricsRouter)

	log.Println("Starting multiple servers...")
	log.Println("📡 API Server:     http://localhost:3080")
	log.Println("🔧 Admin Server:   http://localhost:3081")
	log.Println("📊 Metrics Server: http://localhost:3082")

	// Start API server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := apiApp.Run(0); err != nil {
			log.Printf("API Server error: %v", err)
		}
	}()

	// Start Admin server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := adminApp.Run(0); err != nil {
			log.Printf("Admin Server error: %v", err)
		}
	}()

	// Start Metrics server
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := metricsApp.Run(0); err != nil {
			log.Printf("Metrics Server error: %v", err)
		}
	}()

	fmt.Println("All servers running. Press Ctrl+C to stop.")
	wg.Wait()
}
