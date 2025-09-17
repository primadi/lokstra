package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/health_check"
	"github.com/primadi/lokstra/services/logger"
)

// This example demonstrates comprehensive usage of Lokstra's built-in health check service.
// It shows how to implement robust health monitoring for Kubernetes readiness/liveness probes,
// monitoring systems, and operational visibility.
//
// Learning Objectives:
// - Understand health check service configuration and patterns
// - Learn to implement custom health checkers for various components
// - See Kubernetes integration and monitoring best practices
// - Explore dependency health checking and cascade failures
// - Master operational health visibility and alerting
//
// Documentation: https://github.com/primadi/lokstra/blob/main/docs/services/health_check.md

func main() {
	fmt.Println("🏥 Health Check Service Example - Comprehensive Health Monitoring")
	fmt.Println("")

	// Create registration context
	regCtx := lokstra.NewGlobalRegistrationContext()

	// Register service modules
	regCtx.RegisterModule(health_check.GetModule)
	regCtx.RegisterModule(logger.GetModule)

	// Configure health check service
	healthConfig := map[string]interface{}{
		"endpoint": "/health",
		"timeout":  "30s",
	}

	// Create health check service
	_, err := regCtx.CreateService("health_check", "app-health", healthConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create health check service: %v", err)
	}

	// Create logger service
	loggerConfig := map[string]interface{}{
		"level":  "info",
		"format": "json",
		"output": "stdout",
	}
	_, err = regCtx.CreateService("logger", "app-logger", loggerConfig)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to create logger service: %v", err)
	}

	// Register custom health checks
	setupHealthChecks(regCtx)

	// Create and configure server
	app := lokstra.NewApp(regCtx, "main-app", ":8080")

	// Register routes with direct handler functions
	app.GET("/health", func(ctx *lokstra.Context) error {
		healthSvc, _ := serviceapi.GetService[serviceapi.HealthService](regCtx, "app-health")
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

		result := healthSvc.CheckHealthWithTimeout(10 * time.Second)

		// Set HTTP status based on overall health
		switch result.Status {
		case serviceapi.HealthStatusHealthy:
			ctx.SetStatusCode(200)
		case serviceapi.HealthStatusDegraded:
			ctx.SetStatusCode(200) // Still considered healthy for load balancers
		case serviceapi.HealthStatusUnhealthy:
			ctx.SetStatusCode(503) // Service Unavailable
		}

		logger.Infof("Health check completed: status=%s, duration=%s",
			result.Status, result.Duration)

		return ctx.Ok(result)
	})

	app.GET("/health/ready", func(ctx *lokstra.Context) error {
		healthSvc, _ := serviceapi.GetService[serviceapi.HealthService](regCtx, "app-health")

		// For readiness, we only care about critical dependencies
		criticalChecks := []string{"database", "redis"}

		allHealthy := true
		for _, checkName := range criticalChecks {
			check, exists := healthSvc.GetCheck(context.Background(), checkName)
			if !exists || check.Status == serviceapi.HealthStatusUnhealthy {
				allHealthy = false
				break
			}
		}

		if allHealthy {
			ctx.SetStatusCode(200)
			return ctx.Ok(map[string]interface{}{
				"status":    "ready",
				"timestamp": time.Now(),
			})
		}

		ctx.SetStatusCode(503)
		return ctx.ErrorBadRequest("Service not ready")
	})

	app.GET("/health/live", func(ctx *lokstra.Context) error {
		// For liveness, we check if the application itself is functioning
		isAlive := true // In real implementation, check application state

		if isAlive {
			ctx.SetStatusCode(200)
			return ctx.Ok(map[string]interface{}{
				"status":    "alive",
				"timestamp": time.Now(),
				"uptime":    time.Since(time.Now().Add(-time.Hour)).String(), // Mock uptime
			})
		}

		ctx.SetStatusCode(503)
		return ctx.ErrorBadRequest("Service not alive")
	})

	app.GET("/health/detailed", func(ctx *lokstra.Context) error {
		healthSvc, _ := serviceapi.GetService[serviceapi.HealthService](regCtx, "app-health")
		logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

		result := healthSvc.CheckHealthWithTimeout(15 * time.Second)

		// Add additional system information
		detailed := map[string]interface{}{
			"health_result": result,
			"system_info": map[string]interface{}{
				"go_version":  "go1.21.0",
				"app_version": "1.0.0",
				"build_time":  "2025-09-18T10:00:00Z",
				"environment": "production",
			},
			"registered_checks": healthSvc.ListChecks(),
			"check_history":     generateMockHistory(),
		}

		// Set status code based on health
		switch result.Status {
		case serviceapi.HealthStatusHealthy:
			ctx.SetStatusCode(200)
		case serviceapi.HealthStatusDegraded:
			ctx.SetStatusCode(200)
		case serviceapi.HealthStatusUnhealthy:
			ctx.SetStatusCode(503)
		}

		logger.Infof("Detailed health check requested")

		return ctx.Ok(detailed)
	})

	app.GET("/health/summary", func(ctx *lokstra.Context) error {
		healthSvc, _ := serviceapi.GetService[serviceapi.HealthService](regCtx, "app-health")

		result := healthSvc.CheckHealthWithTimeout(5 * time.Second)

		summary := map[string]interface{}{
			"overall_status":   result.Status,
			"total_checks":     len(result.Checks),
			"healthy_checks":   0,
			"degraded_checks":  0,
			"unhealthy_checks": 0,
			"check_duration":   result.Duration.String(),
			"timestamp":        result.CheckedAt,
		}

		// Count status distribution
		for _, check := range result.Checks {
			switch check.Status {
			case serviceapi.HealthStatusHealthy:
				summary["healthy_checks"] = summary["healthy_checks"].(int) + 1
			case serviceapi.HealthStatusDegraded:
				summary["degraded_checks"] = summary["degraded_checks"].(int) + 1
			case serviceapi.HealthStatusUnhealthy:
				summary["unhealthy_checks"] = summary["unhealthy_checks"].(int) + 1
			}
		}

		return ctx.Ok(summary)
	})

	app.GET("/health/check/:check", func(ctx *lokstra.Context) error {
		healthSvc, _ := serviceapi.GetService[serviceapi.HealthService](regCtx, "app-health")

		checkName := ctx.GetPathParam("check")
		if checkName == "" {
			return ctx.ErrorBadRequest("Check name required")
		}

		check, exists := healthSvc.GetCheck(context.Background(), checkName)
		if !exists {
			return ctx.ErrorNotFound(fmt.Sprintf("Health check '%s' not found", checkName))
		}

		// Set status code based on check result
		switch check.Status {
		case serviceapi.HealthStatusHealthy:
			ctx.SetStatusCode(200)
		case serviceapi.HealthStatusDegraded:
			ctx.SetStatusCode(200)
		case serviceapi.HealthStatusUnhealthy:
			ctx.SetStatusCode(503)
		}

		return ctx.Ok(check)
	})

	// Start server
	server := lokstra.NewServer(regCtx, "health-check-server")
	server.AddApp(app)
	if err != nil {
		lokstra.Logger.Fatalf("Failed to load server: %v", err)
	}

	fmt.Println("🎯 Starting server with comprehensive health monitoring...")
	fmt.Println("🏥 Health endpoints:")
	fmt.Println("   - GET /health (overall health)")
	fmt.Println("   - GET /health/ready (readiness probe)")
	fmt.Println("   - GET /health/live (liveness probe)")
	fmt.Println("   - GET /health/detailed (detailed health report)")
	fmt.Println("")

	// Start the server
	if err := server.Start(); err != nil {
		lokstra.Logger.Fatalf("Failed to start server: %v", err)
	}
}

// setupHealthChecks registers comprehensive health checks for various system components
func setupHealthChecks(regCtx lokstra.RegistrationContext) {
	healthSvc, _ := serviceapi.GetService[serviceapi.HealthService](regCtx, "app-health")
	logger, _ := serviceapi.GetService[serviceapi.Logger](regCtx, "app-logger")

	// 1. Database Health Check
	healthSvc.RegisterCheck("database", func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		// Simulate database connection check
		connectionTime := time.Duration(rand.Intn(100)+10) * time.Millisecond
		time.Sleep(connectionTime)

		// Simulate occasional database issues
		isHealthy := rand.Float32() > 0.1 // 90% success rate

		if isHealthy {
			return serviceapi.HealthCheck{
				Name:      "database",
				Status:    serviceapi.HealthStatusHealthy,
				Message:   "Database connection successful",
				Duration:  time.Since(start),
				CheckedAt: time.Now(),
				Details: map[string]any{
					"connection_time":    connectionTime.String(),
					"active_connections": rand.Intn(20) + 5,
					"max_connections":    100,
					"version":            "PostgreSQL 15.2",
				},
			}
		}

		return serviceapi.HealthCheck{
			Name:      "database",
			Status:    serviceapi.HealthStatusUnhealthy,
			Message:   "Database connection failed",
			Duration:  time.Since(start),
			CheckedAt: time.Now(),
			Error:     "connection timeout after 5s",
			Details: map[string]any{
				"last_successful_check": time.Now().Add(-2 * time.Minute),
				"retry_count":           3,
			},
		}
	})

	// 2. Redis Cache Health Check
	healthSvc.RegisterCheck("redis", func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		// Simulate Redis connection and ping
		pingTime := time.Duration(rand.Intn(50)+5) * time.Millisecond
		time.Sleep(pingTime)

		// Simulate Redis health with occasional degradation
		random := rand.Float32()
		var status serviceapi.HealthStatus
		var message string
		var details map[string]any

		switch {
		case random > 0.85: // 15% chance of issues
			status = serviceapi.HealthStatusUnhealthy
			message = "Redis connection failed"
			details = map[string]any{
				"error":     "connection refused",
				"ping_time": "timeout",
			}
		case random > 0.7: // 15% chance of degraded performance
			status = serviceapi.HealthStatusDegraded
			message = "Redis responding slowly"
			details = map[string]any{
				"ping_time":         pingTime.String(),
				"memory_usage":      "85%",
				"connected_clients": rand.Intn(100) + 50,
				"warning":           "high memory usage",
			}
		default: // 70% healthy
			status = serviceapi.HealthStatusHealthy
			message = "Redis connection healthy"
			details = map[string]any{
				"ping_time":         pingTime.String(),
				"memory_usage":      fmt.Sprintf("%d%%", rand.Intn(50)+20),
				"connected_clients": rand.Intn(50) + 10,
				"version":           "Redis 7.0.5",
			}
		}

		return serviceapi.HealthCheck{
			Name:      "redis",
			Status:    status,
			Message:   message,
			Duration:  time.Since(start),
			CheckedAt: time.Now(),
			Details:   details,
		}
	})

	// 3. External API Health Check
	healthSvc.RegisterCheck("external_api", func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		// Simulate external API call
		responseTime := time.Duration(rand.Intn(500)+100) * time.Millisecond
		time.Sleep(responseTime)

		// Simulate API health
		isHealthy := rand.Float32() > 0.2 // 80% success rate
		httpStatus := 200
		if !isHealthy {
			httpStatus = 500 + rand.Intn(4) // 500-504
		}

		var status serviceapi.HealthStatus
		var message string

		if isHealthy {
			status = serviceapi.HealthStatusHealthy
			message = "External API responding"
		} else {
			status = serviceapi.HealthStatusUnhealthy
			message = fmt.Sprintf("External API error (HTTP %d)", httpStatus)
		}

		return serviceapi.HealthCheck{
			Name:      "external_api",
			Status:    status,
			Message:   message,
			Duration:  time.Since(start),
			CheckedAt: time.Now(),
			Details: map[string]any{
				"endpoint":             "https://api.example.com/health",
				"http_status":          httpStatus,
				"response_time":        responseTime.String(),
				"rate_limit_remaining": rand.Intn(1000),
			},
		}
	})

	// 4. Disk Space Health Check
	healthSvc.RegisterCheck("disk_space", func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		// Simulate disk space check
		usedPercent := rand.Intn(100)

		var status serviceapi.HealthStatus
		var message string

		switch {
		case usedPercent > 90:
			status = serviceapi.HealthStatusUnhealthy
			message = "Disk space critically low"
		case usedPercent > 80:
			status = serviceapi.HealthStatusDegraded
			message = "Disk space running low"
		default:
			status = serviceapi.HealthStatusHealthy
			message = "Disk space sufficient"
		}

		return serviceapi.HealthCheck{
			Name:      "disk_space",
			Status:    status,
			Message:   message,
			Duration:  time.Since(start),
			CheckedAt: time.Now(),
			Details: map[string]any{
				"used_percent": usedPercent,
				"free_space":   fmt.Sprintf("%.1fGB", float64(100-usedPercent)*0.5),
				"total_space":  "50GB",
				"mount_point":  "/var/lib/app",
			},
		}
	})

	// 5. Memory Usage Health Check
	healthSvc.RegisterCheck("memory", func(ctx context.Context) serviceapi.HealthCheck {
		start := time.Now()

		// Simulate memory usage check
		memoryPercent := rand.Intn(100)

		var status serviceapi.HealthStatus
		var message string

		switch {
		case memoryPercent > 95:
			status = serviceapi.HealthStatusUnhealthy
			message = "Memory usage critically high"
		case memoryPercent > 85:
			status = serviceapi.HealthStatusDegraded
			message = "Memory usage high"
		default:
			status = serviceapi.HealthStatusHealthy
			message = "Memory usage normal"
		}

		return serviceapi.HealthCheck{
			Name:      "memory",
			Status:    status,
			Message:   message,
			Duration:  time.Since(start),
			CheckedAt: time.Now(),
			Details: map[string]any{
				"used_percent":   memoryPercent,
				"used_bytes":     memoryPercent * 10485760, // ~10MB per percent
				"total_bytes":    1073741824,               // 1GB
				"gc_collections": rand.Intn(1000),
			},
		}
	})

	logger.Infof("Registered %d health checks", len(healthSvc.ListChecks()))
}

// generateMockHistory creates mock health check history for demonstration
func generateMockHistory() []map[string]interface{} {
	history := make([]map[string]interface{}, 5)

	for i := 0; i < 5; i++ {
		status := "healthy"
		if rand.Float32() > 0.8 {
			status = "unhealthy"
		} else if rand.Float32() > 0.9 {
			status = "degraded"
		}

		history[i] = map[string]interface{}{
			"timestamp": time.Now().Add(-time.Duration(i+1) * time.Minute),
			"status":    status,
			"duration":  fmt.Sprintf("%dms", rand.Intn(500)+50),
			"issues":    rand.Intn(3),
		}
	}

	return history
}

// Health Check Service Key Concepts:
//
// 1. Health Check Types:
//    - Liveness: Is the application alive and not deadlocked?
//    - Readiness: Is the application ready to receive traffic?
//    - Startup: Has the application finished starting up?
//    - Custom: Application-specific health indicators
//
// 2. Dependency Health:
//    - Database connectivity and query performance
//    - Cache availability and response times
//    - External API accessibility and response
//    - Disk space and resource availability
//    - Message queue connectivity
//
// 3. Health Status Levels:
//    - Healthy: All systems operational
//    - Degraded: Functional but with issues
//    - Unhealthy: Critical failures affecting service
//
// 4. Check Implementation:
//    - Fast execution (< 30s timeout)
//    - Meaningful error messages
//    - Detailed diagnostic information
//    - Proper error handling and recovery

// Kubernetes Integration:
//
// 1. Liveness Probe:
//    - Restart container if probe fails
//    - Check application deadlock/hang state
//    - Avoid external dependency checks
//    - Fast and lightweight checks only
//
// 2. Readiness Probe:
//    - Remove from service endpoints if failing
//    - Check external dependencies
//    - Verify initialization completion
//    - Can include slower dependency checks
//
// 3. Startup Probe:
//    - Allow longer startup times
//    - Check initialization progress
//    - Prevent premature liveness checks
//    - Useful for slow-starting applications

// Monitoring and Alerting:
//
// 1. Health Metrics:
//    - Track check success/failure rates
//    - Monitor check duration trends
//    - Alert on sustained failures
//    - Dashboard visualization
//
// 2. Incident Response:
//    - Automated failover procedures
//    - Escalation based on check types
//    - Runbook integration
//    - Dependencies impact analysis

// Test Commands:
//
// # Start the application
// go run main.go
//
// # Test health endpoints
// curl http://localhost:8080/health
// curl http://localhost:8080/health/ready
// curl http://localhost:8080/health/live
// curl http://localhost:8080/health/detailed
// curl http://localhost:8080/health/summary
//
// # Test individual checks
// curl http://localhost:8080/health/check/database
// curl http://localhost:8080/health/check/redis
// curl http://localhost:8080/health/check/external_api
//
// # Monitor health over time
// watch -n 5 'curl -s http://localhost:8080/health | jq .status'
//
// # Test failure scenarios (will show random failures)
// for i in {1..20}; do curl -s http://localhost:8080/health | jq .status; sleep 1; done
