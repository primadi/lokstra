package config

import (
	"fmt"
	"strings"
)

// ValidateLayeredServices validates layered service configuration:
// 1. All dependencies in depends-on must exist in previous layers or simple services
// 2. All dependencies in depends-on must be used in config
// 3. All service names in config values must be in depends-on (if they're service references)
func ValidateLayeredServices(services *ServicesConfig) error {
	if !services.IsLayered() {
		// Simple mode doesn't need layer validation
		return nil
	}

	// Track available services (includes simple services if any)
	availableServices := make(map[string]string) // serviceName -> layerName (or "simple")

	// Add simple services to available list
	if services.Simple != nil {
		for _, svc := range services.Simple {
			availableServices[svc.Name] = "simple"
		}
	}

	// Validate each layer in order
	for _, layerName := range services.Order {
		layerServices := services.Layered[layerName]

		for _, svc := range layerServices {
			// Validate depends-on: all must be available
			for _, dep := range svc.DependsOn {
				if sourceLayer, ok := availableServices[dep]; !ok {
					return fmt.Errorf(
						"service '%s' in layer '%s' depends on '%s' which is not available yet",
						svc.Name, layerName, dep,
					)
				} else if sourceLayer == layerName {
					return fmt.Errorf(
						"service '%s' in layer '%s' depends on '%s' which is in the same layer (circular dependency)",
						svc.Name, layerName, dep,
					)
				}
			}

			// Validate all depends-on are used in config
			if err := validateDependenciesUsed(svc); err != nil {
				return fmt.Errorf("service '%s' in layer '%s': %w", svc.Name, layerName, err)
			}

			// Validate all service references in config are in depends-on
			if err := validateConfigReferences(svc, availableServices); err != nil {
				return fmt.Errorf("service '%s' in layer '%s': %w", svc.Name, layerName, err)
			}

			// Add this service to available services
			availableServices[svc.Name] = layerName
		}
	}

	return nil
}

// validateDependenciesUsed checks that all dependencies in depends-on are used in config
func validateDependenciesUsed(svc *Service) error {
	if len(svc.DependsOn) == 0 {
		return nil
	}

	// Collect all service names referenced in config values
	usedServices := make(map[string]bool)
	for _, value := range svc.Config {
		if strValue, ok := value.(string); ok {
			usedServices[strValue] = true
		}
	}

	// Check each dependency is used
	var unused []string
	for _, dep := range svc.DependsOn {
		if !usedServices[dep] {
			unused = append(unused, dep)
		}
	}

	if len(unused) > 0 {
		return fmt.Errorf(
			"dependency '%s' in depends-on but not used in config",
			strings.Join(unused, "', '"),
		)
	}

	return nil
}

// validateConfigReferences checks that all service references in config are in depends-on
func validateConfigReferences(svc *Service, availableServices map[string]string) error {
	if len(svc.Config) == 0 {
		return nil
	}

	dependsOnSet := make(map[string]bool)
	for _, dep := range svc.DependsOn {
		dependsOnSet[dep] = true
	}

	// Check each config value that looks like a service reference
	for key, value := range svc.Config {
		strValue, ok := value.(string)
		if !ok {
			continue
		}

		// If the value matches an available service name, it should be in depends-on
		if _, isService := availableServices[strValue]; isService {
			if !dependsOnSet[strValue] {
				return fmt.Errorf(
					"config key '%s' references service '%s' which is not in depends-on. Add it to depends-on: [%s]",
					key, strValue, strValue,
				)
			}
		}
	}

	return nil
}

// GetServiceLayer returns the layer name for a given service, or empty string if not found
func (sc *ServicesConfig) GetServiceLayer(serviceName string) string {
	if sc.IsSimple() {
		for _, svc := range sc.Simple {
			if svc.Name == serviceName {
				return "simple"
			}
		}
		return ""
	}

	if sc.IsLayered() {
		for _, layerName := range sc.Order {
			for _, svc := range sc.Layered[layerName] {
				if svc.Name == serviceName {
					return layerName
				}
			}
		}
	}

	return ""
}

// GetServiceDependencies returns a map of service dependencies for visualization/debugging
func (sc *ServicesConfig) GetServiceDependencies() map[string][]string {
	deps := make(map[string][]string)

	for _, svc := range sc.GetAllServices() {
		if len(svc.DependsOn) > 0 {
			deps[svc.Name] = svc.DependsOn
		}
	}

	return deps
}
