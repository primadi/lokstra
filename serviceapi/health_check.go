package serviceapi

import (
	"context"
	"time"

	"github.com/primadi/lokstra/core/service"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck represents a single health check result
type HealthCheck struct {
	Name      string         `json:"name"`
	Status    HealthStatus   `json:"status"`
	Message   string         `json:"message,omitempty"`
	Details   map[string]any `json:"details,omitempty"`
	Duration  time.Duration  `json:"duration"`
	CheckedAt time.Time      `json:"checked_at"`
	Error     string         `json:"error,omitempty"`
}

// HealthResult represents the overall health check result
type HealthResult struct {
	Status    HealthStatus           `json:"status"`
	Checks    map[string]HealthCheck `json:"checks"`
	Duration  time.Duration          `json:"duration"`
	CheckedAt time.Time              `json:"checked_at"`
}

// HealthChecker defines a function that performs a health check
type HealthChecker func(ctx context.Context) HealthCheck

// HealthService provides health checking capabilities for Kubernetes and monitoring
type HealthService interface {
	service.Service

	// RegisterCheck registers a health check with a name
	RegisterCheck(name string, checker HealthChecker)

	// UnregisterCheck removes a health check
	UnregisterCheck(name string)

	// CheckHealth performs all registered health checks
	CheckHealth(ctx context.Context) HealthResult

	// CheckHealthWithTimeout performs health checks with a timeout
	CheckHealthWithTimeout(timeout time.Duration) HealthResult

	// IsHealthy returns true if all checks are healthy
	IsHealthy(ctx context.Context) bool

	// GetCheck performs a specific health check by name
	GetCheck(ctx context.Context, name string) (HealthCheck, bool)

	// ListChecks returns all registered check names
	ListChecks() []string
}
