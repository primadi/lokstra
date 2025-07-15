package health

import (
	"context"
	"lokstra/common/iface"
	"sync"
	"time"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusUnhealthy HealthStatus = "unhealthy"
	StatusUnknown   HealthStatus = "unknown"
)

type HealthCheck struct {
	Name        string
	CheckFunc   func(ctx context.Context) error
	Timeout     time.Duration
	LastChecked time.Time
	LastStatus  HealthStatus
	LastError   error
}

type HealthService struct {
	instanceName string
	config       map[string]any
	checks       map[string]*HealthCheck
	mutex        sync.RWMutex
}

func (h *HealthService) InstanceName() string {
	return h.instanceName
}

func (h *HealthService) GetConfig(key string) any {
	return h.config[key]
}

func (h *HealthService) RegisterCheck(name string, checkFunc func(ctx context.Context) error, timeout time.Duration) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.checks[name] = &HealthCheck{
		Name:      name,
		CheckFunc: checkFunc,
		Timeout:   timeout,
		LastStatus: StatusUnknown,
	}
}

func (h *HealthService) RemoveCheck(name string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	delete(h.checks, name)
}

func (h *HealthService) CheckHealth(ctx context.Context) map[string]HealthStatus {
	h.mutex.RLock()
	checks := make(map[string]*HealthCheck, len(h.checks))
	for name, check := range h.checks {
		checks[name] = check
	}
	h.mutex.RUnlock()

	results := make(map[string]HealthStatus)
	var wg sync.WaitGroup

	for name, check := range checks {
		wg.Add(1)
		go func(name string, check *HealthCheck) {
			defer wg.Done()
			status := h.runCheck(ctx, check)
			results[name] = status
		}(name, check)
	}

	wg.Wait()
	return results
}

func (h *HealthService) IsHealthy(ctx context.Context) bool {
	results := h.CheckHealth(ctx)
	for _, status := range results {
		if status != StatusHealthy {
			return false
		}
	}
	return true
}

func (h *HealthService) GetCheckStatus(name string) (HealthStatus, error, time.Time) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if check, exists := h.checks[name]; exists {
		return check.LastStatus, check.LastError, check.LastChecked
	}
	return StatusUnknown, nil, time.Time{}
}

func (h *HealthService) runCheck(ctx context.Context, check *HealthCheck) HealthStatus {
	checkCtx, cancel := context.WithTimeout(ctx, check.Timeout)
	defer cancel()

	err := check.CheckFunc(checkCtx)
	
	h.mutex.Lock()
	check.LastChecked = time.Now()
	if err != nil {
		check.LastStatus = StatusUnhealthy
		check.LastError = err
	} else {
		check.LastStatus = StatusHealthy
		check.LastError = nil
	}
	h.mutex.Unlock()

	return check.LastStatus
}

func newHealthService(instanceName string, config map[string]any) (*HealthService, error) {
	return &HealthService{
		instanceName: instanceName,
		config:       config,
		checks:       make(map[string]*HealthCheck),
	}, nil
}
