package health_check

import (
	"context"
	"testing"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

func TestRealMemoryHealthChecker(t *testing.T) {
	// Test memory health checker with real implementation
	checker := MemoryHealthChecker(1024) // 1GB limit

	ctx := context.Background()
	result := checker(ctx)

	// Verify the result structure
	if result.Name != "memory" {
		t.Errorf("Expected name 'memory', got %s", result.Name)
	}

	if result.Status == "" {
		t.Error("Expected status to be set")
	}

	if result.Message == "" {
		t.Error("Expected message to be set")
	}

	if result.Duration < 0 {
		t.Error("Expected duration to be non-negative")
	}

	// Check that details are populated
	details := result.Details
	if details == nil {
		t.Error("Expected details to be populated")
	} else {
		expectedFields := []string{"alloc_mb", "sys_mb", "max_memory_mb", "usage_percent", "total_alloc_mb", "num_gc"}
		for _, field := range expectedFields {
			if _, exists := details[field]; !exists {
				t.Errorf("Expected detail field '%s' to exist", field)
			}
		}
	}

	t.Logf("Memory Health Check Result: %+v", result)
}

func TestRealDiskHealthChecker(t *testing.T) {
	// Test disk health checker with real implementation
	// Use current directory as path since it should exist
	checker := DiskHealthChecker(".", 90.0) // 90% threshold

	ctx := context.Background()
	result := checker(ctx)

	// Verify the result structure
	if result.Name != "disk" {
		t.Errorf("Expected name 'disk', got %s", result.Name)
	}

	if result.Status == "" {
		t.Error("Expected status to be set")
	}

	if result.Message == "" {
		t.Error("Expected message to be set")
	}

	if result.Duration < 0 {
		t.Error("Expected duration to be non-negative")
	}

	// Check that details are populated
	details := result.Details
	if details == nil {
		t.Error("Expected details to be populated")
	} else {
		expectedFields := []string{"path", "total_bytes", "used_bytes", "available_bytes", "usage_percent", "max_usage_percent"}
		for _, field := range expectedFields {
			if _, exists := details[field]; !exists {
				t.Errorf("Expected detail field '%s' to exist", field)
			}
		}
	}

	t.Logf("Disk Health Check Result: %+v", result)
}

func TestRealApplicationHealthChecker(t *testing.T) {
	// Test application health checker
	appName := "test-app"
	checker := ApplicationHealthChecker(appName)

	ctx := context.Background()
	result := checker(ctx)

	// Verify the result
	if result.Name != "application" {
		t.Errorf("Expected name 'application', got %s", result.Name)
	}

	if result.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected status 'healthy', got %s", result.Status)
	}

	expectedMessage := appName + " is running normally"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Duration < 0 {
		t.Error("Expected duration to be non-negative")
	}

	t.Logf("Application Health Check Result: %+v", result)
}

func TestCustomHealthChecker(t *testing.T) {
	// Test custom health checker with different scenarios

	// Test healthy scenario
	healthyChecker := CustomHealthChecker("test_service", func(ctx context.Context) (bool, string, map[string]any) {
		return true, "Service is running well", map[string]any{
			"uptime":  "5m30s",
			"version": "1.0.0",
		}
	})

	result := healthyChecker(context.Background())
	if result.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", result.Status)
	}

	// Test unhealthy scenario
	unhealthyChecker := CustomHealthChecker("test_service", func(ctx context.Context) (bool, string, map[string]any) {
		return false, "Service has issues", map[string]any{
			"error": "Connection failed",
		}
	})

	result = unhealthyChecker(context.Background())
	if result.Status != serviceapi.HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", result.Status)
	}

	t.Logf("Custom Health Check Results tested successfully")
}

func TestHealthCheckPerformance(t *testing.T) {
	// Test that health checks complete within reasonable time
	checkers := map[string]serviceapi.HealthChecker{
		"memory": MemoryHealthChecker(1024),
		"disk":   DiskHealthChecker(".", 90.0),
		"app":    ApplicationHealthChecker("perf-test"),
	}

	ctx := context.Background()

	for name, checker := range checkers {
		start := time.Now()
		result := checker(ctx)
		elapsed := time.Since(start)

		// Health checks should complete quickly (under 1 second)
		if elapsed > time.Second {
			t.Errorf("Health check '%s' took too long: %v", name, elapsed)
		}

		// Verify duration is recorded accurately
		if result.Duration == 0 {
			t.Errorf("Health check '%s' did not record duration", name)
		}

		t.Logf("Health check '%s' completed in %v (recorded: %v)", name, elapsed, result.Duration)
	}
}
