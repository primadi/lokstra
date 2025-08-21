package health_check

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

// Module for health check service
type Module struct{}

// Name implements iface.Module
func (m *Module) Name() string {
	return MODULE_NAME
}

// Description implements iface.Module
func (m *Module) Description() string {
	return "Health Check Service for Kubernetes and monitoring"
}

// Register implements iface.Module
func (m *Module) Register(regCtx iface.RegistrationContext) error {
	// Register health check service factory
	regCtx.RegisterServiceFactory(MODULE_NAME, func(config any) (service.Service, error) {
		return NewService(), nil
	})

	// Register health check HTTP handlers
	registerHealthHandlers(regCtx)

	return nil
}

// GetModule returns a new health check module instance
func GetModule() iface.Module {
	return &Module{}
}

// registerHealthHandlers registers all health check HTTP endpoints
func registerHealthHandlers(regCtx iface.RegistrationContext) {
	// Helper function to get health service
	getHealthService := func() serviceapi.HealthService {
		service, err := regCtx.GetService(MODULE_NAME + ".default")
		if err != nil {
			// Try to create if doesn't exist
			service, err = regCtx.CreateService(MODULE_NAME, MODULE_NAME+".default", nil)
			if err != nil {
				return nil
			}
		}
		if health, ok := service.(serviceapi.HealthService); ok {
			return health
		}
		return nil
	}

	// Main health check endpoint
	regCtx.RegisterHandler("health", func(ctx *request.Context) error {
		health := getHealthService()
		if health == nil {
			ctx.StatusCode = http.StatusServiceUnavailable
			return ctx.ErrorInternal("Health service not available")
		}

		result := health.CheckHealthWithTimeout(30 * time.Second)
		if result.Status == serviceapi.HealthStatusUnhealthy {
			ctx.StatusCode = http.StatusServiceUnavailable
		}
		return ctx.Ok(result)
	})

	// Kubernetes liveness probe endpoint
	regCtx.RegisterHandler("health.liveness", func(ctx *request.Context) error {
		return ctx.Ok(map[string]any{
			"status":     "healthy",
			"service":    "lokstra",
			"checked_at": time.Now(),
		})
	})

	// Kubernetes readiness probe endpoint
	regCtx.RegisterHandler("health.readiness", func(ctx *request.Context) error {
		health := getHealthService()
		if health == nil {
			ctx.StatusCode = http.StatusServiceUnavailable
			return ctx.Ok(map[string]any{
				"status":     "not_ready",
				"ready":      false,
				"checked_at": time.Now(),
				"error":      "Health service not available",
			})
		}

		isReady := health.IsHealthy(context.Background())
		response := map[string]any{
			"status":     "ready",
			"ready":      isReady,
			"checked_at": time.Now(),
		}

		if !isReady {
			ctx.StatusCode = http.StatusServiceUnavailable
			response["status"] = "not_ready"
		}

		return ctx.Ok(response)
	})

	// Detailed health information endpoint
	regCtx.RegisterHandler("health.detailed", func(ctx *request.Context) error {
		health := getHealthService()
		if health == nil {
			ctx.StatusCode = http.StatusServiceUnavailable
			return ctx.ErrorInternal("Health service not available")
		}

		result := health.CheckHealthWithTimeout(30 * time.Second)

		response := map[string]any{
			"status":     result.Status,
			"checked_at": result.CheckedAt,
			"duration":   result.Duration.String(),
			"checks":     result.Checks,
			"summary": map[string]any{
				"total":     len(result.Checks),
				"healthy":   0,
				"degraded":  0,
				"unhealthy": 0,
			},
		}

		// Count check statuses
		summary := response["summary"].(map[string]any)
		for _, check := range result.Checks {
			switch check.Status {
			case serviceapi.HealthStatusHealthy:
				summary["healthy"] = summary["healthy"].(int) + 1
			case serviceapi.HealthStatusDegraded:
				summary["degraded"] = summary["degraded"].(int) + 1
			case serviceapi.HealthStatusUnhealthy:
				summary["unhealthy"] = summary["unhealthy"].(int) + 1
			}
		}

		if result.Status == serviceapi.HealthStatusUnhealthy {
			ctx.StatusCode = http.StatusServiceUnavailable
		}

		return ctx.Ok(response)
	})

	// List all health checks endpoint
	regCtx.RegisterHandler("health.list", func(ctx *request.Context) error {
		health := getHealthService()
		if health == nil {
			ctx.StatusCode = http.StatusServiceUnavailable
			return ctx.ErrorInternal("Health service not available")
		}

		checks := health.ListChecks()
		return ctx.Ok(map[string]any{
			"checks":     checks,
			"count":      len(checks),
			"checked_at": time.Now(),
		})
	})

	// Individual health check endpoint
	regCtx.RegisterHandler("health.check", func(ctx *request.Context) error {
		health := getHealthService()
		if health == nil {
			ctx.StatusCode = http.StatusServiceUnavailable
			return ctx.ErrorInternal("Health service not available")
		}

		checkName := ctx.GetPathParam("name")
		if checkName == "" {
			return ctx.ErrorBadRequest("check name is required")
		}

		check, exists := health.GetCheck(context.Background(), checkName)
		if !exists {
			return ctx.ErrorNotFound(fmt.Sprintf("health check '%s' not found", checkName))
		}

		if check.Status == serviceapi.HealthStatusUnhealthy {
			ctx.StatusCode = http.StatusServiceUnavailable
		}

		return ctx.Ok(check)
	})

	// Prometheus metrics endpoint
	regCtx.RegisterHandler("health.metrics", func(ctx *request.Context) error {
		health := getHealthService()
		if health == nil {
			ctx.Writer.Header().Set("Content-Type", "text/plain")
			ctx.Writer.WriteHeader(http.StatusServiceUnavailable)
			_, err := ctx.Writer.Write([]byte("# Health service not available\n"))
			return err
		}

		result := health.CheckHealthWithTimeout(30 * time.Second)

		var metrics []string

		// Overall health metric
		overallHealth := 0
		if result.Status == serviceapi.HealthStatusHealthy {
			overallHealth = 1
		}

		metrics = append(metrics, fmt.Sprintf("lokstra_health_status %d", overallHealth))
		metrics = append(metrics, fmt.Sprintf("lokstra_health_check_duration_seconds %f", result.Duration.Seconds()))
		metrics = append(metrics, fmt.Sprintf("lokstra_health_checks_total %d", len(result.Checks)))

		// Individual check metrics
		for name, check := range result.Checks {
			checkHealth := 0
			if check.Status == serviceapi.HealthStatusHealthy {
				checkHealth = 1
			}

			metrics = append(metrics, fmt.Sprintf("lokstra_health_check_status{name=\"%s\"} %d", name, checkHealth))
			metrics = append(metrics, fmt.Sprintf("lokstra_health_check_duration_seconds{name=\"%s\"} %f", name, check.Duration.Seconds()))
		}

		response := ""
		for _, metric := range metrics {
			response += metric + "\n"
		}

		ctx.Writer.Header().Set("Content-Type", "text/plain")
		ctx.Writer.WriteHeader(http.StatusOK)
		_, err := ctx.Writer.Write([]byte(response))
		return err
	})
}
