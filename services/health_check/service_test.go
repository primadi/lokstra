package health_check

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

func TestHealthCheckService(t *testing.T) {
	service := NewService()

	// Test service name
	if service.Name() != MODULE_NAME {
		t.Errorf("Expected service name %s, got %s", MODULE_NAME, service.Name())
	}

	// Test start and stop
	ctx := context.Background()
	if err := service.Start(ctx); err != nil {
		t.Errorf("Failed to start service: %v", err)
	}

	if err := service.Stop(ctx); err != nil {
		t.Errorf("Failed to stop service: %v", err)
	}
}

func TestRegisterAndUnregisterCheck(t *testing.T) {
	service := NewService()

	// Create a simple health checker
	healthyChecker := func(ctx context.Context) serviceapi.HealthCheck {
		return serviceapi.HealthCheck{
			Name:      "test",
			Status:    serviceapi.HealthStatusHealthy,
			Message:   "Test is healthy",
			CheckedAt: time.Now(),
			Duration:  time.Millisecond,
		}
	}

	// Register the check
	service.RegisterCheck("test", healthyChecker)

	// Verify it's in the list
	checks := service.ListChecks()
	if len(checks) != 1 || checks[0] != "test" {
		t.Errorf("Expected 1 check named 'test', got %v", checks)
	}

	// Unregister the check
	service.UnregisterCheck("test")

	// Verify it's removed
	checks = service.ListChecks()
	if len(checks) != 0 {
		t.Errorf("Expected 0 checks, got %v", checks)
	}
}

func TestCheckHealth(t *testing.T) {
	service := NewService()

	// Create healthy and unhealthy checkers
	healthyChecker := func(ctx context.Context) serviceapi.HealthCheck {
		return serviceapi.HealthCheck{
			Name:      "healthy",
			Status:    serviceapi.HealthStatusHealthy,
			Message:   "All good",
			CheckedAt: time.Now(),
			Duration:  time.Millisecond,
		}
	}

	unhealthyChecker := func(ctx context.Context) serviceapi.HealthCheck {
		return serviceapi.HealthCheck{
			Name:      "unhealthy",
			Status:    serviceapi.HealthStatusUnhealthy,
			Message:   "Something's wrong",
			Error:     "Test error",
			CheckedAt: time.Now(),
			Duration:  time.Millisecond,
		}
	}

	degradedChecker := func(ctx context.Context) serviceapi.HealthCheck {
		return serviceapi.HealthCheck{
			Name:      "degraded",
			Status:    serviceapi.HealthStatusDegraded,
			Message:   "Performance degraded",
			CheckedAt: time.Now(),
			Duration:  time.Millisecond,
		}
	}

	// Test with only healthy check
	service.RegisterCheck("healthy", healthyChecker)
	result := service.CheckHealth(context.Background())

	if result.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected overall status to be healthy, got %s", result.Status)
	}

	if len(result.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(result.Checks))
	}

	if !service.IsHealthy(context.Background()) {
		t.Error("Expected IsHealthy to return true")
	}

	// Test with unhealthy check
	service.RegisterCheck("unhealthy", unhealthyChecker)
	result = service.CheckHealth(context.Background())

	if result.Status != serviceapi.HealthStatusUnhealthy {
		t.Errorf("Expected overall status to be unhealthy, got %s", result.Status)
	}

	if len(result.Checks) != 2 {
		t.Errorf("Expected 2 check results, got %d", len(result.Checks))
	}

	if service.IsHealthy(context.Background()) {
		t.Error("Expected IsHealthy to return false")
	}

	// Test with degraded check only
	service.UnregisterCheck("healthy")
	service.UnregisterCheck("unhealthy")
	service.RegisterCheck("degraded", degradedChecker)
	result = service.CheckHealth(context.Background())

	if result.Status != serviceapi.HealthStatusDegraded {
		t.Errorf("Expected overall status to be degraded, got %s", result.Status)
	}
}

func TestGetCheck(t *testing.T) {
	service := NewService()

	healthyChecker := func(ctx context.Context) serviceapi.HealthCheck {
		return serviceapi.HealthCheck{
			Name:      "test",
			Status:    serviceapi.HealthStatusHealthy,
			Message:   "Test is healthy",
			CheckedAt: time.Now(),
			Duration:  time.Millisecond,
		}
	}

	service.RegisterCheck("test", healthyChecker)

	// Test getting existing check
	check, exists := service.GetCheck(context.Background(), "test")
	if !exists {
		t.Error("Expected check to exist")
	}
	if check.Name != "test" {
		t.Errorf("Expected check name 'test', got %s", check.Name)
	}
	if check.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected check status healthy, got %s", check.Status)
	}

	// Test getting non-existing check
	_, exists = service.GetCheck(context.Background(), "nonexistent")
	if exists {
		t.Error("Expected check to not exist")
	}
}

func TestCheckHealthWithTimeout(t *testing.T) {
	service := NewService()

	// Create a slow checker that takes longer than timeout
	slowChecker := func(ctx context.Context) serviceapi.HealthCheck {
		select {
		case <-time.After(100 * time.Millisecond):
			return serviceapi.HealthCheck{
				Name:      "slow",
				Status:    serviceapi.HealthStatusHealthy,
				Message:   "Slow but healthy",
				CheckedAt: time.Now(),
				Duration:  100 * time.Millisecond,
			}
		case <-ctx.Done():
			return serviceapi.HealthCheck{
				Name:      "slow",
				Status:    serviceapi.HealthStatusUnhealthy,
				Message:   "Timed out",
				Error:     "context timeout",
				CheckedAt: time.Now(),
				Duration:  0,
			}
		}
	}

	service.RegisterCheck("slow", slowChecker)

	// Test with very short timeout
	result := service.CheckHealthWithTimeout(10 * time.Millisecond)

	// The check should still complete even if it takes longer than timeout
	// because each checker runs in its own goroutine
	if len(result.Checks) != 1 {
		t.Errorf("Expected 1 check result, got %d", len(result.Checks))
	}
}

// Test helper functions for common components
func TestDatabaseHealthChecker(t *testing.T) {
	// Mock database pool
	mockPool := &mockDbPool{shouldFail: false}
	checker := DatabaseHealthChecker(mockPool)

	check := checker(context.Background())
	if check.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", check.Status)
	}

	// Test with failing database
	mockPool.shouldFail = true
	check = checker(context.Background())
	if check.Status != serviceapi.HealthStatusUnhealthy {
		t.Errorf("Expected unhealthy status, got %s", check.Status)
	}
}

func TestMemoryHealthChecker(t *testing.T) {
	checker := MemoryHealthChecker(1024)
	check := checker(context.Background())

	if check.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", check.Status)
	}

	if check.Details["max_memory_mb"] != int64(1024) {
		t.Errorf("Expected max_memory_mb to be 1024, got %v", check.Details["max_memory_mb"])
	}
}

func TestDiskHealthChecker(t *testing.T) {
	checker := DiskHealthChecker("/tmp", 80.0)
	check := checker(context.Background())

	if check.Status != serviceapi.HealthStatusHealthy {
		t.Errorf("Expected healthy status, got %s", check.Status)
	}

	if check.Details["path"] != "/tmp" {
		t.Errorf("Expected path to be /tmp, got %v", check.Details["path"])
	}

	if check.Details["max_usage_percent"] != 80.0 {
		t.Errorf("Expected max_usage_percent to be 80.0, got %v", check.Details["max_usage_percent"])
	}
}

// Mock implementations for testing

type mockDbPool struct {
	shouldFail bool
}

func (m *mockDbPool) Acquire(ctx context.Context, schema string) (serviceapi.DbConn, error) {
	if m.shouldFail {
		return nil, errors.New("mock database connection failed")
	}
	return &mockDbConn{}, nil
}

type mockDbConn struct{}

func (m *mockDbConn) Begin(ctx context.Context) (serviceapi.DbTx, error) { return nil, nil }
func (m *mockDbConn) Transaction(ctx context.Context, fn func(tx serviceapi.DbExecutor) error) error {
	return nil
}
func (m *mockDbConn) Release() error { return nil }
func (m *mockDbConn) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	return nil, nil
}
func (m *mockDbConn) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	return nil, nil
}
func (m *mockDbConn) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	return nil
}
func (m *mockDbConn) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	return nil
}
func (m *mockDbConn) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	return nil
}
func (m *mockDbConn) SelectOneRowMap(ctx context.Context, query string, args ...any) (serviceapi.RowMap, error) {
	return nil, nil
}
func (m *mockDbConn) SelectManyRowMap(ctx context.Context, query string, args ...any) ([]serviceapi.RowMap, error) {
	return nil, nil
}
func (m *mockDbConn) SelectManyWithMapper(ctx context.Context, fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {
	return nil, nil
}
func (m *mockDbConn) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	return false, nil
}
func (m *mockDbConn) IsErrorNoRows(err error) bool {
	return false
}
