package main

import (
	"log"
	"sync"
	"time"

	"github.com/primadi/lokstra"
)

// Health check example

var (
	startTime      time.Time
	isHealthy      = true
	healthMu       sync.RWMutex
	dbConnected    = true
	cacheConnected = true
)

// Basic health check
func HealthHandler() map[string]any {
	healthMu.RLock()
	defer healthMu.RUnlock()

	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	}

	return map[string]any{
		"status": status,
	}
}

// Detailed health check
func HealthDetailedHandler() map[string]any {
	healthMu.RLock()
	defer healthMu.RUnlock()

	checks := map[string]any{
		"database": map[string]any{
			"status":  getStatus(dbConnected),
			"latency": "2ms",
		},
		"cache": map[string]any{
			"status":  getStatus(cacheConnected),
			"latency": "1ms",
		},
		"disk": map[string]any{
			"status":    "healthy",
			"usage":     "45%",
			"available": "100GB",
		},
		"memory": map[string]any{
			"status": "healthy",
			"usage":  "512MB",
		},
	}

	overallStatus := "healthy"
	if !dbConnected || !cacheConnected {
		overallStatus = "degraded"
	}
	if !isHealthy {
		overallStatus = "unhealthy"
	}

	return map[string]any{
		"status":    overallStatus,
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).Seconds(),
		"checks":    checks,
	}
}

// Readiness check (for load balancers)
func ReadinessHandler() map[string]any {
	healthMu.RLock()
	defer healthMu.RUnlock()

	ready := isHealthy && dbConnected && cacheConnected

	return map[string]any{
		"ready":  ready,
		"uptime": time.Since(startTime).Seconds(),
		"services": map[string]bool{
			"database": dbConnected,
			"cache":    cacheConnected,
		},
	}
}

// Liveness check (for orchestrators)
func LivenessHandler() map[string]any {
	healthMu.RLock()
	defer healthMu.RUnlock()

	return map[string]any{
		"alive":  isHealthy,
		"uptime": time.Since(startTime).Seconds(),
	}
}

// Simulate health changes
func ToggleHealthHandler() map[string]any {
	healthMu.Lock()
	isHealthy = !isHealthy
	healthMu.Unlock()

	return map[string]any{
		"message": "Health status toggled",
		"healthy": isHealthy,
	}
}

func ToggleDBHandler() map[string]any {
	healthMu.Lock()
	dbConnected = !dbConnected
	healthMu.Unlock()

	return map[string]any{
		"message":      "Database status toggled",
		"db_connected": dbConnected,
	}
}

func getStatus(connected bool) string {
	if connected {
		return "healthy"
	}
	return "unhealthy"
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Health Checks Example</h1>
		<p>Demonstrates different types of health checks</p>
		<h2>Health Check Endpoints</h2>
		<ul>
			<li><a href="/health">Health</a> - Basic health check</li>
			<li><a href="/health/detailed">Detailed Health</a> - Comprehensive health info</li>
			<li><a href="/readiness">Readiness</a> - Ready to serve traffic?</li>
			<li><a href="/liveness">Liveness</a> - Is process alive?</li>
		</ul>
		<h2>Simulate Issues</h2>
		<ul>
			<li><a href="/toggle-health">Toggle Health</a> - Simulate app health issue</li>
			<li><a href="/toggle-db">Toggle DB</a> - Simulate DB connection issue</li>
		</ul>
		<h2>Use Cases</h2>
		<ul>
			<li><strong>/health</strong> - Load balancers</li>
			<li><strong>/readiness</strong> - Kubernetes readiness probe</li>
			<li><strong>/liveness</strong> - Kubernetes liveness probe</li>
		</ul>
	</body>
	</html>
	`
}

func main() {
	startTime = time.Now()

	// Create router
	router := lokstra.NewRouter("main")
	router.GET("/", HomeHandler)
	router.GET("/health", HealthHandler)
	router.GET("/health/detailed", HealthDetailedHandler)
	router.GET("/readiness", ReadinessHandler)
	router.GET("/liveness", LivenessHandler)
	router.GET("/toggle-health", ToggleHealthHandler)
	router.GET("/toggle-db", ToggleDBHandler)

	// Create app
	app := lokstra.NewApp("health-checks", ":3100", router)

	log.Println("Server starting on :3100")
	log.Println("Health check endpoints:")
	log.Println("  /health - Basic health")
	log.Println("  /health/detailed - Detailed health")
	log.Println("  /readiness - Ready for traffic")
	log.Println("  /liveness - Process alive")

	if err := app.Run(0); err != nil {
		log.Fatal(err)
	}
}
