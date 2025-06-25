package core

import (
	"fmt"
	"lokstra/internal"
)

type CreateInstanceServiceFactory func(instanceName string, config map[string]any) (Service, error)

type RunningService struct {
	Service  Service
	Settings *ServiceConfig
}

var (
	registeredFactories = map[string]CreateInstanceServiceFactory{}
	activeServices      = map[string]RunningService{} // instanceName -> service

	responseHooks = []OnResponseHook{}
	contextHooks  = []OnContextHook{}
)

func RegisterNamedService(serviceType string, factory CreateInstanceServiceFactory) {
	if _, exists := registeredFactories[serviceType]; exists {
		panic("duplicate service type registered: " + serviceType)
	}
	registeredFactories[serviceType] = factory
}

func StartServiceFromConfig(cfg *ServiceConfig) error {
	if !internal.IsEnabled(cfg.Enabled, true) {
		return nil // skip disabled
	}

	factory, ok := registeredFactories[cfg.Type]
	if !ok {
		return fmt.Errorf("no factory for service type %s", cfg.Type)
	}

	svc, err := factory(cfg.Name, cfg.Config)
	if err != nil {
		return err
	}

	registerActiveService(svc, cfg)
	return nil
}

func StartNamedService(serviceType, instanceName string, config map[string]any) error {
	return StartServiceFromConfig(&ServiceConfig{
		Type:    serviceType,
		Name:    instanceName,
		Enabled: "true",
		Config:  config,
	})
}

func registerActiveService(svc Service, cfg *ServiceConfig) {
	nameKey := svc.InstanceName()
	activeServices[nameKey] = RunningService{Service: svc,
		Settings: cfg}

	if h, ok := svc.(OnResponseHook); ok {
		responseHooks = append(responseHooks, h)
	}
	if h, ok := svc.(OnContextHook); ok {
		contextHooks = append(contextHooks, h)
	}
}

func GetService(instanceName string) Service {
	return activeServices[instanceName].Service
}

func GetServiceTyped[T any](instanceName string) T {
	return activeServices[instanceName].Service.(T)
}

func GetAllServices() []RunningService {
	result := make([]RunningService, 0, len(activeServices))
	for _, s := range activeServices {
		result = append(result, s)
	}

	return result
}

func ReloadAllServices() error {
	for name, entry := range activeServices {
		if r, ok := entry.Service.(ReloadableService); ok {
			err := r.Reload(entry.Settings.Config)
			if err != nil {
				return fmt.Errorf("failed to reload service %s: %w", name, err)
			}
		}
	}
	return nil
}

func GetResponseHooks() []OnResponseHook {
	return responseHooks
}

func GetContextHooks() []OnContextHook {
	return contextHooks
}
