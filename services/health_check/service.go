package health_check

import (
	"context"
	"sync"
	"time"

	"github.com/primadi/lokstra/serviceapi"
)

const MODULE_NAME = "health_check"

// Service implements serviceapi.HealthService
type Service struct {
	checkers map[string]serviceapi.HealthChecker
	mutex    sync.RWMutex
}

// NewService creates a new health check service
func NewService() *Service {
	return &Service{
		checkers: make(map[string]serviceapi.HealthChecker),
	}
}

// Name implements service.Service
func (s *Service) Name() string {
	return MODULE_NAME
}

// Start implements service.Service
func (s *Service) Start(ctx context.Context) error {
	// Health service doesn't need startup logic
	return nil
}

// Stop implements service.Service
func (s *Service) Stop(ctx context.Context) error {
	// Clear all registered checkers
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.checkers = make(map[string]serviceapi.HealthChecker)
	return nil
}

// RegisterCheck implements serviceapi.HealthService
func (s *Service) RegisterCheck(name string, checker serviceapi.HealthChecker) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.checkers[name] = checker
}

// UnregisterCheck implements serviceapi.HealthService
func (s *Service) UnregisterCheck(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.checkers, name)
}

// CheckHealth implements serviceapi.HealthService
func (s *Service) CheckHealth(ctx context.Context) serviceapi.HealthResult {
	return s.CheckHealthWithTimeout(30 * time.Second)
}

// CheckHealthWithTimeout implements serviceapi.HealthService
func (s *Service) CheckHealthWithTimeout(timeout time.Duration) serviceapi.HealthResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	result := serviceapi.HealthResult{
		Checks:    make(map[string]serviceapi.HealthCheck),
		CheckedAt: start,
		Status:    serviceapi.HealthStatusHealthy,
	}

	s.mutex.RLock()
	checkers := make(map[string]serviceapi.HealthChecker)
	for name, checker := range s.checkers {
		checkers[name] = checker
	}
	s.mutex.RUnlock()

	// Run all checks concurrently
	type checkResult struct {
		name  string
		check serviceapi.HealthCheck
	}

	checkChan := make(chan checkResult, len(checkers))
	var wg sync.WaitGroup

	for name, checker := range checkers {
		wg.Add(1)
		go func(n string, c serviceapi.HealthChecker) {
			defer wg.Done()
			check := c(ctx)
			check.Name = n
			checkChan <- checkResult{name: n, check: check}
		}(name, checker)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(checkChan)
	}()

	// Collect results
	for checkRes := range checkChan {
		result.Checks[checkRes.name] = checkRes.check

		// If any check is unhealthy, mark the overall status as unhealthy
		if checkRes.check.Status == serviceapi.HealthStatusUnhealthy {
			result.Status = serviceapi.HealthStatusUnhealthy
		} else if checkRes.check.Status == serviceapi.HealthStatusDegraded && result.Status == serviceapi.HealthStatusHealthy {
			result.Status = serviceapi.HealthStatusDegraded
		}
	}

	result.Duration = time.Since(start)
	return result
}

// IsHealthy implements serviceapi.HealthService
func (s *Service) IsHealthy(ctx context.Context) bool {
	result := s.CheckHealth(ctx)
	return result.Status == serviceapi.HealthStatusHealthy
}

// GetCheck implements serviceapi.HealthService
func (s *Service) GetCheck(ctx context.Context, name string) (serviceapi.HealthCheck, bool) {
	s.mutex.RLock()
	checker, exists := s.checkers[name]
	s.mutex.RUnlock()

	if !exists {
		return serviceapi.HealthCheck{}, false
	}

	check := checker(ctx)
	check.Name = name
	return check, true
}

// ListChecks implements serviceapi.HealthService
func (s *Service) ListChecks() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	names := make([]string, 0, len(s.checkers))
	for name := range s.checkers {
		names = append(names, name)
	}
	return names
}
