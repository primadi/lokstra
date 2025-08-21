package health_check

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

// HandlerConfig contains configuration for health check handlers
type HandlerConfig struct {
	HealthService serviceapi.HealthService
	Timeout       time.Duration
}

// NewHandlerConfig creates a new handler configuration
func NewHandlerConfig(healthService serviceapi.HealthService) *HandlerConfig {
	return &HandlerConfig{
		HealthService: healthService,
		Timeout:       30 * time.Second, // Default timeout
	}
}

// HealthCheckHandler returns a handler for the main health check endpoint
// Used by Kubernetes liveness and readiness probes
func (h *HandlerConfig) HealthCheckHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		result := h.HealthService.CheckHealthWithTimeout(h.Timeout)

		// Set appropriate HTTP status based on health status
		switch result.Status {
		case serviceapi.HealthStatusUnhealthy:
			ctx.StatusCode = http.StatusServiceUnavailable
		case serviceapi.HealthStatusDegraded:
			ctx.StatusCode = http.StatusOK // Still return 200 for degraded
		default:
			ctx.StatusCode = http.StatusOK
		}

		return ctx.Ok(result)
	}
}

// LivenessHandler returns a simplified handler for Kubernetes liveness probe
// This should only check if the service is running, not external dependencies
func (h *HandlerConfig) LivenessHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		response := map[string]any{
			"status":     "healthy",
			"service":    "lokstra",
			"checked_at": time.Now(),
		}
		return ctx.Ok(response)
	}
}

// ReadinessHandler returns a handler for Kubernetes readiness probe
// This checks if the service is ready to accept traffic
func (h *HandlerConfig) ReadinessHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		isReady := h.HealthService.IsHealthy(context.Background())

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
	}
}

// DetailedHealthHandler returns a detailed health check with all checks
func (h *HandlerConfig) DetailedHealthHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		result := h.HealthService.CheckHealthWithTimeout(h.Timeout)

		// Create detailed response with additional metadata
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
	}
}

// SingleCheckHandler returns a handler for checking a specific health check
func (h *HandlerConfig) SingleCheckHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		checkName := ctx.GetPathParam("name")
		if checkName == "" {
			return ctx.ErrorBadRequest("check name is required")
		}

		check, exists := h.HealthService.GetCheck(context.Background(), checkName)
		if !exists {
			return ctx.ErrorNotFound(fmt.Sprintf("health check '%s' not found", checkName))
		}

		if check.Status == serviceapi.HealthStatusUnhealthy {
			ctx.StatusCode = http.StatusServiceUnavailable
		}

		return ctx.Ok(check)
	}
}

// ListChecksHandler returns a handler that lists all available health checks
func (h *HandlerConfig) ListChecksHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		checks := h.HealthService.ListChecks()

		response := map[string]any{
			"checks":     checks,
			"count":      len(checks),
			"checked_at": time.Now(),
		}

		return ctx.Ok(response)
	}
}

// PrometheusMetricsHandler returns health check metrics in Prometheus format
func (h *HandlerConfig) PrometheusMetricsHandler() request.HandlerFunc {
	return func(ctx *request.Context) error {
		result := h.HealthService.CheckHealthWithTimeout(h.Timeout)

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
	}
}

// RegisterHealthRoutes registers all health check routes with the provided registration context
func RegisterHealthRoutes(regCtx interface{}, healthService serviceapi.HealthService) {
	config := NewHandlerConfig(healthService)

	// Standard health check endpoints for Kubernetes
	if registerHandler, ok := regCtx.(interface {
		RegisterHandler(string, request.HandlerFunc)
	}); ok {
		// Main health endpoint
		registerHandler.RegisterHandler("health", config.HealthCheckHandler())

		// Kubernetes probe endpoints
		registerHandler.RegisterHandler("health.liveness", config.LivenessHandler())
		registerHandler.RegisterHandler("health.readiness", config.ReadinessHandler())

		// Detailed health information
		registerHandler.RegisterHandler("health.detailed", config.DetailedHealthHandler())

		// Individual check endpoint
		registerHandler.RegisterHandler("health.check", config.SingleCheckHandler())

		// List all checks
		registerHandler.RegisterHandler("health.list", config.ListChecksHandler())

		// Prometheus metrics
		registerHandler.RegisterHandler("health.metrics", config.PrometheusMetricsHandler())
	}
}
