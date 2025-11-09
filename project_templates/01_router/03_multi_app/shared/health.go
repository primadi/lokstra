package shared

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

// SetupHealthRouter creates a health check router for any app
// The appName parameter helps identify which app the health check is for
func SetupHealthRouter(appName string) lokstra.Router {
	r := lokstra.NewRouter(fmt.Sprintf("%s_health_router", appName))

	// Health check endpoints - no middleware for performance
	r.GET("/health", func() (*HealthStatus, error) {
		return handleHealth(appName)
	})

	r.GET("/ready", func() (*ReadyStatus, error) {
		return handleReady(appName)
	})

	return r
}

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string    `json:"status"`
	App       string    `json:"app"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// handleHealth returns the health status
func handleHealth(appName string) (*HealthStatus, error) {
	return &HealthStatus{
		Status:    "healthy",
		App:       appName,
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}, nil
}

// ReadyStatus represents the readiness check response
type ReadyStatus struct {
	Ready     bool              `json:"ready"`
	App       string            `json:"app"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// handleReady checks if the application is ready
func handleReady(appName string) (*ReadyStatus, error) {
	// In production: check database, cache, external services, etc.
	return &ReadyStatus{
		Ready:     true,
		App:       appName,
		Timestamp: time.Now(),
		Services: map[string]string{
			"database": "ok",
			"cache":    "ok",
		},
	}, nil
}
