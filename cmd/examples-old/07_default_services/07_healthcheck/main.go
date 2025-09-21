package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/health_check"
)

func main() {
	fmt.Println("🏥 Lokstra Health Check Example - YAML Configuration")
	fmt.Println("===================================================")

	// 1. Create registration context with all default services
	regCtx := lokstra.NewGlobalRegistrationContext()

	// 2. Register health check module (required for YAML config)
	health_check.GetModule().Register(regCtx)

	// 3. Register application handlers
	setupApplicationHandlers(regCtx)

	// 4. Load configuration and create server
	server := newServerFromConfig(regCtx, ".")

	// 5. Setup health checks if health service is available
	if healthService, err := regCtx.GetService("health-service"); err == nil {
		if health, ok := healthService.(serviceapi.HealthService); ok {
			setupHealthChecks(health)
			fmt.Println("✅ Health service configured from YAML and health checks registered")
		}
	} else {
		// Fallback: create health service manually if not in config
		healthService, err := regCtx.CreateService("health_check", "health-service", true, nil)
		if err != nil {
			log.Fatalf("❌ Failed to create health service: %v", err)
		}
		health := healthService.(serviceapi.HealthService)
		setupHealthChecks(health)
		fmt.Println("✅ Health service created manually and health checks registered")
	}

	fmt.Println("\n🚀 Starting Health Check Example Application...")
	fmt.Println("📊 Health Endpoints (configured via YAML):")
	fmt.Println("   - GET /health                - Main health check")
	fmt.Println("   - GET /health/liveness       - Kubernetes liveness probe")
	fmt.Println("   - GET /health/readiness      - Kubernetes readiness probe")
	fmt.Println("   - GET /health/detailed       - Detailed health information")
	fmt.Println("   - GET /health/list           - List all health checks")
	fmt.Println("   - GET /health/check/{name}   - Individual health check")
	fmt.Println("   - GET /health/metrics        - Prometheus metrics")
	fmt.Println("\n📝 Application Endpoints:")
	fmt.Println("   - GET /                      - Application info")
	fmt.Println("   - GET /api/status            - API status")
	fmt.Println("   - POST /api/simulate-error   - Simulate service error")
	fmt.Println("   - POST /api/recover          - Recover from error")
	fmt.Println("\n🌐 Server running on: http://localhost:8080")
	fmt.Println("💡 Try: curl http://localhost:8080/health")

	if err := server.Start(true); err != nil {
		log.Fatalf("❌ Failed to start application: %v", err)
	}
}

func newServerFromConfig(ctx lokstra.RegistrationContext, dir string) *lokstra.Server {
	cfg, err := lokstra.LoadConfigDir(dir)
	if err != nil {
		panic(fmt.Sprintf("Failed to load config from %s: %v", dir, err))
	}

	server, err := lokstra.NewServerFromConfig(ctx, cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to create server from config: %v", err))
	}

	fmt.Println("Config loaded successfully:")
	fmt.Printf("- Server: %+v\n", cfg.Server)
	fmt.Printf("- Apps: %d\n", len(cfg.Apps))
	fmt.Printf("- Services: %d\n", len(cfg.Services))
	fmt.Printf("- Modules: %d\n", len(cfg.Modules))

	return server
}

// setupHealthChecks registers various health checks to demonstrate different scenarios
func setupHealthChecks(health serviceapi.HealthService) {
	fmt.Println("\n📋 Registering Health Checks...")

	// 1. Application health check
	health.RegisterCheck("application", health_check.ApplicationHealthChecker("health-check-example"))
	fmt.Println("   ✅ Application health check")

	// 2. Memory health check (512MB limit)
	health.RegisterCheck("memory", health_check.MemoryHealthChecker(512))
	fmt.Println("   ✅ Memory health check (512MB limit)")

	// 3. Disk health check (90% limit on temp directory)
	health.RegisterCheck("disk", health_check.DiskHealthChecker("/tmp", 90.0))
	fmt.Println("   ✅ Disk health check (/tmp, 90% limit)")

	// 4. Simulated database health check
	health.RegisterCheck("database", createSimulatedDatabaseChecker())
	fmt.Println("   ✅ Simulated database health check")

	// 5. Simulated external service health check
	health.RegisterCheck("external_api", createSimulatedExternalServiceChecker())
	fmt.Println("   ✅ Simulated external service health check")

	// 6. Business logic health check
	health.RegisterCheck("business_logic", createBusinessLogicChecker())
	fmt.Println("   ✅ Business logic health check")

	// 7. Periodic task health check
	health.RegisterCheck("periodic_tasks", createPeriodicTaskChecker())
	fmt.Println("   ✅ Periodic task health check")

	fmt.Println("✅ All health checks registered successfully")
}

// setupApplicationHandlers registers additional application endpoints
func setupApplicationHandlers(regCtx lokstra.RegistrationContext) {
	// Health check handlers (mapped to YAML config)
	regCtx.RegisterHandler("health.check", func(ctx *lokstra.Context) error {
		// Get health service
		healthService, err := regCtx.GetService("health-service")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		health := healthService.(serviceapi.HealthService)
		result := health.CheckHealth(context.Background())

		// Return appropriate HTTP status based on overall health
		switch result.Status {
		case serviceapi.HealthStatusHealthy:
			return ctx.Ok(result)
		case serviceapi.HealthStatusDegraded:
			ctx.SetStatusCode(200) // Still OK for load balancers
			return ctx.Ok(result)
		default:
			ctx.SetStatusCode(503) // Service Unavailable
			return ctx.Ok(result)
		}
	})

	regCtx.RegisterHandler("health.liveness", func(ctx *lokstra.Context) error {
		// Liveness probe - just check if service is running
		return ctx.Ok(map[string]any{
			"status":    "alive",
			"timestamp": time.Now(),
			"uptime":    time.Since(startTime).String(),
		})
	})

	regCtx.RegisterHandler("health.readiness", func(ctx *lokstra.Context) error {
		// Readiness probe - check if service can handle requests
		healthService, err := regCtx.GetService("health-service")
		if err != nil {
			ctx.SetStatusCode(503)
			return ctx.Ok(map[string]any{
				"status": "not_ready",
				"reason": "health service unavailable",
			})
		}

		health := healthService.(serviceapi.HealthService)
		result := health.CheckHealth(context.Background())

		if result.Status == serviceapi.HealthStatusUnhealthy {
			ctx.SetStatusCode(503)
			return ctx.Ok(map[string]any{
				"status": "not_ready",
				"reason": "unhealthy",
				"checks": result.Checks,
			})
		}

		return ctx.Ok(map[string]any{
			"status":    "ready",
			"timestamp": time.Now(),
			"checks":    len(result.Checks),
		})
	})

	regCtx.RegisterHandler("health.detailed", func(ctx *lokstra.Context) error {
		healthService, err := regCtx.GetService("health-service")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		health := healthService.(serviceapi.HealthService)
		result := health.CheckHealth(context.Background())
		return ctx.Ok(result)
	})

	regCtx.RegisterHandler("health.list", func(ctx *lokstra.Context) error {
		healthService, err := regCtx.GetService("health-service")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		health := healthService.(serviceapi.HealthService)
		checks := health.ListChecks()
		return ctx.Ok(map[string]any{
			"checks":    checks,
			"count":     len(checks),
			"timestamp": time.Now(),
		})
	})

	regCtx.RegisterHandler("health.check_by_name", func(ctx *lokstra.Context) error {
		name := ctx.GetPathParam("name")
		if name == "" {
			return ctx.ErrorBadRequest("Check name is required")
		}

		healthService, err := regCtx.GetService("health-service")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		health := healthService.(serviceapi.HealthService)
		result, found := health.GetCheck(context.Background(), name)

		if !found {
			return ctx.ErrorNotFound("Health check not found")
		}

		return ctx.Ok(result)
	})

	regCtx.RegisterHandler("health.metrics", func(ctx *lokstra.Context) error {
		healthService, err := regCtx.GetService("health-service")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		health := healthService.(serviceapi.HealthService)
		result := health.CheckHealth(context.Background())

		// Generate Prometheus-style metrics
		metrics := generatePrometheusMetrics(result)
		return ctx.WriteRaw("text/plain; charset=utf-8", 200, []byte(metrics))
	})

	// Application info endpoint (root handler)
	regCtx.RegisterHandler("app.home", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"application": "Lokstra Health Check Example",
			"version":     "1.0.0",
			"description": "Demonstrates health check capabilities with YAML configuration",
			"features": []string{
				"YAML-configured health endpoints",
				"Multiple health check types",
				"Kubernetes-ready probes",
				"Prometheus metrics",
				"Custom business logic checks",
			},
			"endpoints": map[string]string{
				"health":           "/health",
				"liveness":         "/health/liveness",
				"readiness":        "/health/readiness",
				"detailed_health":  "/health/detailed",
				"health_list":      "/health/list",
				"individual_check": "/health/check/{name}",
				"metrics":          "/health/metrics",
				"app_status":       "/api/status",
			},
			"timestamp": time.Now(),
		})
	})

	// API status endpoint
	regCtx.RegisterHandler("app.status", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status":    "operational",
			"uptime":    time.Since(startTime).String(),
			"requests":  requestCount,
			"timestamp": time.Now(),
		})
	})

	// Simulate error endpoint (for testing degraded/unhealthy states)
	regCtx.RegisterHandler("app.simulate-error", func(ctx *lokstra.Context) error {
		simulateError = true
		errorStartTime = time.Now()
		return ctx.Ok(map[string]any{
			"message":   "Error simulation activated",
			"timestamp": time.Now(),
			"note":      "Health checks will now report degraded/unhealthy status",
		})
	})

	// Recover from error endpoint
	regCtx.RegisterHandler("app.recover", func(ctx *lokstra.Context) error {
		simulateError = false
		return ctx.Ok(map[string]any{
			"message":        "Recovered from simulated error",
			"error_duration": time.Since(errorStartTime).String(),
			"timestamp":      time.Now(),
			"note":           "Health checks will return to healthy status",
		})
	})
}

// generatePrometheusMetrics creates Prometheus-style metrics from health check results
func generatePrometheusMetrics(result serviceapi.HealthResult) string {
	metrics := ""

	// Overall health status
	overallStatus := 1.0
	switch result.Status {
	case serviceapi.HealthStatusDegraded:
		overallStatus = 0.5
	case serviceapi.HealthStatusUnhealthy:
		overallStatus = 0.0
	}

	metrics += "# HELP health_status Overall health status (1=healthy, 0.5=degraded, 0=unhealthy)\n"
	metrics += "# TYPE health_status gauge\n"
	metrics += fmt.Sprintf("health_status %f\n", overallStatus)

	metrics += "# HELP health_checks_total Total number of health checks\n"
	metrics += "# TYPE health_checks_total gauge\n"
	metrics += fmt.Sprintf("health_checks_total %d\n", len(result.Checks))

	// Individual check statuses
	metrics += "# HELP health_check_status Individual health check status\n"
	metrics += "# TYPE health_check_status gauge\n"

	for _, check := range result.Checks {
		status := 1.0
		switch check.Status {
		case serviceapi.HealthStatusDegraded:
			status = 0.5
		case serviceapi.HealthStatusUnhealthy:
			status = 0.0
		}
		metrics += fmt.Sprintf("health_check_status{name=\"%s\"} %f\n", check.Name, status)
	}

	// Check durations
	metrics += "# HELP health_check_duration_seconds Health check duration in seconds\n"
	metrics += "# TYPE health_check_duration_seconds gauge\n"

	for _, check := range result.Checks {
		metrics += fmt.Sprintf("health_check_duration_seconds{name=\"%s\"} %f\n",
			check.Name, check.Duration.Seconds())
	}

	return metrics
}

// Global variables for simulation
var (
	startTime      = time.Now()
	requestCount   = 0
	simulateError  = false
	errorStartTime time.Time
)

// createSimulatedDatabaseChecker creates a health checker that simulates database connectivity
func createSimulatedDatabaseChecker() serviceapi.HealthChecker {
	return health_check.CustomHealthChecker("database", func(ctx context.Context) (bool, string, map[string]any) {
		// Simulate database check delay
		time.Sleep(5 * time.Millisecond)

		if simulateError && rand.Float32() < 0.3 {
			return false, "Database connection failed", map[string]any{
				"host":       "localhost:5432",
				"database":   "lokstra_example",
				"error":      "connection timeout",
				"last_check": time.Now(),
			}
		}

		// Simulate occasional slow responses
		responseTime := 5 + rand.Intn(20) // 5-25ms
		status := responseTime < 15       // Degraded if > 15ms

		message := "Database connection is healthy"
		if !status {
			message = "Database connection is slow"
		}

		return status, message, map[string]any{
			"host":          "localhost:5432",
			"database":      "lokstra_example",
			"response_time": fmt.Sprintf("%dms", responseTime),
			"pool_size":     10,
			"active_conns":  rand.Intn(8) + 1,
			"last_check":    time.Now(),
		}
	})
}

// createSimulatedExternalServiceChecker simulates an external service dependency
func createSimulatedExternalServiceChecker() serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		// Simulate external service call delay
		delay := time.Duration(10+rand.Intn(50)) * time.Millisecond
		time.Sleep(delay)

		status := serviceapi.HealthStatusHealthy
		message := "External service is responsive"

		// Simulate various scenarios
		if simulateError {
			if rand.Float32() < 0.5 {
				status = serviceapi.HealthStatusUnhealthy
				message = "External service is unreachable"
			} else {
				status = serviceapi.HealthStatusDegraded
				message = "External service is responding slowly"
			}
		} else if delay > 30*time.Millisecond {
			status = serviceapi.HealthStatusDegraded
			message = "External service response time is elevated"
		}

		return serviceapi.HealthCheck{
			Name:    "external_api",
			Status:  status,
			Message: message,
			Details: map[string]any{
				"endpoint":      "https://api.example.com/v1/status",
				"response_time": delay.String(),
				"timeout":       "5s",
				"last_success":  time.Now().Add(-delay),
				"retry_count":   0,
			},
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}
}

// createBusinessLogicChecker creates a health check for custom business logic
func createBusinessLogicChecker() serviceapi.HealthChecker {
	return health_check.CustomHealthChecker("business_logic", func(ctx context.Context) (bool, string, map[string]any) {
		requestCount++

		// Simulate business logic validation
		isHealthy := true
		message := "Business logic is operating normally"

		details := map[string]any{
			"request_count": requestCount,
			"uptime":        time.Since(startTime).String(),
			"last_check":    time.Now(),
			"version":       "1.0.0",
			"environment":   "development",
		}

		if simulateError {
			// Simulate business logic issues
			if rand.Float32() < 0.4 {
				isHealthy = false
				message = "Business logic validation failed"
				details["error"] = "Data consistency check failed"
				details["affected_records"] = rand.Intn(100) + 1
			}
		}

		// Check if request rate is too high (simulate overload)
		if requestCount > 100 && time.Since(startTime) < 30*time.Second {
			isHealthy = false
			message = "System is overloaded"
			details["requests_per_second"] = float64(requestCount) / time.Since(startTime).Seconds()
		}

		return isHealthy, message, details
	})
}

// createPeriodicTaskChecker monitors background task health
func createPeriodicTaskChecker() serviceapi.HealthChecker {
	return func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		status := serviceapi.HealthStatusHealthy
		message := "All periodic tasks are running normally"

		// Simulate periodic task monitoring
		lastJobTime := time.Now().Add(-time.Duration(rand.Intn(300)) * time.Second)
		timeSinceLastJob := time.Since(lastJobTime)

		details := map[string]any{
			"last_job_run":    lastJobTime,
			"time_since_last": timeSinceLastJob.String(),
			"next_job_in":     (5*time.Minute - timeSinceLastJob).String(),
			"completed_jobs":  rand.Intn(50) + 10,
			"failed_jobs":     rand.Intn(3),
			"queue_size":      rand.Intn(10),
		}

		if simulateError || timeSinceLastJob > 10*time.Minute {
			status = serviceapi.HealthStatusDegraded
			message = "Periodic tasks are delayed"
			details["issue"] = "Job scheduler experiencing delays"
		}

		if timeSinceLastJob > 30*time.Minute {
			status = serviceapi.HealthStatusUnhealthy
			message = "Periodic tasks have stopped"
			details["issue"] = "Job scheduler is not responding"
		}

		return serviceapi.HealthCheck{
			Name:      "periodic_tasks",
			Status:    status,
			Message:   message,
			Details:   details,
			CheckedAt: start,
			Duration:  time.Since(start),
		}
	}
}
