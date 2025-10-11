package config

import (
	"fmt"
	"strings"
)

// ValidateServices validates service configuration for both simple and layered modes:
// 1. All dependencies in depends-on must exist
// 2. All dependencies in depends-on must be used in config
// 3. All service names in config values must be in depends-on (if they're service references)
// 4. No circular dependencies (for layered mode)
func ValidateServices(services *ServicesConfig) error {
	if services.IsSimple() {
		return validateSimpleServices(services)
	}
	if services.IsLayered() {
		return validateLayeredServices(services)
	}
	return nil
}

// ValidateLayeredServices is deprecated, use ValidateServices instead.
// Kept for backward compatibility.
func ValidateLayeredServices(services *ServicesConfig) error {
	return ValidateServices(services)
}

// validateSimpleServices validates simple (flat array) service configuration
func validateSimpleServices(services *ServicesConfig) error {
	// Build available services map
	availableServices := make(map[string]bool)
	for _, svc := range services.Simple {
		availableServices[svc.Name] = true
	}

	// Validate each service
	for _, svc := range services.Simple {
		// Validate depends-on: all must exist (extract actual service names)
		for _, dep := range svc.DependsOn {
			serviceName := extractServiceNameFromDep(dep)
			if !availableServices[serviceName] {
				return fmt.Errorf(
					"service '%s' depends on '%s' which does not exist. Available services: %v",
					svc.Name, serviceName, getServiceNames(services.Simple),
				)
			}
		}

		// Validate all depends-on are used in config
		if err := validateDependenciesUsed(svc); err != nil {
			return fmt.Errorf("service '%s': %w", svc.Name, err)
		}
	}

	return nil
}

// validateLayeredServices validates layered service configuration
func validateLayeredServices(services *ServicesConfig) error {
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
			// Validate depends-on: all must be available (extract actual service names)
			for _, dep := range svc.DependsOn {
				serviceName := extractServiceNameFromDep(dep)
				if sourceLayer, ok := availableServices[serviceName]; !ok {
					return fmt.Errorf(
						"service '%s' in layer '%s' depends on '%s' which is not available yet",
						svc.Name, layerName, serviceName,
					)
				} else if sourceLayer == layerName {
					return fmt.Errorf(
						"service '%s' in layer '%s' depends on '%s' which is in the same layer (circular dependency)",
						svc.Name, layerName, serviceName,
					)
				}
			}

			// Validate all depends-on are used in config
			if err := validateDependenciesUsed(svc); err != nil {
				return fmt.Errorf("service '%s' in layer '%s': %w", svc.Name, layerName, err)
			}

			// NOTE: Rule 3 (validateConfigReferences) is REMOVED
			// Dependencies are now explicitly declared with local_key:service-name format
			// This prevents false positives from literal strings matching service names

			// Add this service to available services
			availableServices[svc.Name] = layerName
		}
	}

	return nil
}

// validateDependenciesUsed checks that all dependencies in depends-on are used in config
// With auto-injection enabled, this validation is relaxed - dependencies will be auto-injected
func validateDependenciesUsed(svc *Service) error {
	if len(svc.DependsOn) == 0 {
		return nil
	}

	// Parse depends-on entries (support both "service-name" and "local_key:service-name" formats)
	dependencyMap := parseDependsOn(svc.DependsOn)

	// Collect all service names referenced in config values
	usedServices := make(map[string]bool)
	for _, value := range svc.Config {
		if strValue, ok := value.(string); ok {
			usedServices[strValue] = true
		}
	}

	// NOTE: With auto-injection, we DON'T require dependencies to be in config
	// They will be auto-injected by injectDependencies()
	// This validation is now only for documentation/visibility purposes

	// Optional: Log unused dependencies for debugging
	for _, dep := range svc.DependsOn {
		depInfo := dependencyMap[dep]
		if !usedServices[depInfo.ServiceName] {
			// Not an error - will be auto-injected
			// log.Printf("Dependency '%s' will be auto-injected for service '%s'", dep, svc.Name)
		}
	}

	return nil
}

// getServiceNames returns a list of service names for error messages
func getServiceNames(services []*Service) []string {
	names := make([]string, len(services))
	for i, svc := range services {
		names[i] = svc.Name
	}
	return names
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

// DependencyInfo holds parsed dependency information
type DependencyInfo struct {
	LocalKey    string // Local config key (e.g., "user_service")
	ServiceName string // Actual service name (e.g., "user-service")
	Original    string // Original depends-on entry
}

// parseDependsOn parses depends-on entries supporting both formats:
// - "service-name" - simple format (local key = service name)
// - "local_key:service-name" - explicit format (local key != service name)
//
// Examples:
//
//	"user-service" -> {LocalKey: "user-service", ServiceName: "user-service"}
//	"user_service:user-service" -> {LocalKey: "user_service", ServiceName: "user-service"}
func parseDependsOn(dependsOn []string) map[string]*DependencyInfo {
	result := make(map[string]*DependencyInfo)

	for _, dep := range dependsOn {
		info := &DependencyInfo{Original: dep}

		// Check if it contains colon (explicit local_key:service-name format)
		if idx := strings.Index(dep, ":"); idx > 0 {
			info.LocalKey = dep[:idx]
			info.ServiceName = dep[idx+1:]
		} else {
			// Simple format: local key same as service name
			info.LocalKey = dep
			info.ServiceName = dep
		}

		result[dep] = info
	}

	return result
}

// extractServiceNameFromDep extracts the actual service name from depends-on entry
// Supports both "service-name" and "local_key:service-name" formats
func extractServiceNameFromDep(dep string) string {
	if idx := strings.Index(dep, ":"); idx > 0 {
		return dep[idx+1:]
	}
	return dep
}
