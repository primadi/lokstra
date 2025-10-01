package lokstra_registry

import (
	"fmt"
	"reflect"
)

// ServiceIntegrationMode defines how services communicate
type ServiceIntegrationMode string

const (
	ServiceModeMonolith      ServiceIntegrationMode = "monolith"      // Direct function calls
	ServiceModeMicroservices ServiceIntegrationMode = "microservices" // HTTP client calls
	ServiceModeHybrid        ServiceIntegrationMode = "hybrid"        // Mixed mode
)

// ServiceIntegrationConfig configures how services interact
type ServiceIntegrationConfig struct {
	Mode           ServiceIntegrationMode
	ServiceURLs    map[string]string // service-name -> base-url mapping
	Timeout        int               // HTTP timeout in seconds
	RetryCount     int               // Number of retries for HTTP calls
	CircuitBreaker bool              // Enable circuit breaker pattern
}

var serviceIntegrationConfig = ServiceIntegrationConfig{
	Mode:           ServiceModeMonolith, // Default to monolith
	ServiceURLs:    make(map[string]string),
	Timeout:        30,
	RetryCount:     3,
	CircuitBreaker: true,
}

// SetServiceIntegrationMode configures how services communicate globally
func SetServiceIntegrationMode(mode ServiceIntegrationMode, config ...map[string]any) {
	serviceIntegrationConfig.Mode = mode

	if len(config) > 0 {
		cfg := config[0]

		// Set service URLs for microservices mode
		if urls, ok := cfg["service_urls"].(map[string]string); ok {
			serviceIntegrationConfig.ServiceURLs = urls
		}

		// Set HTTP timeout
		if timeout, ok := cfg["timeout"].(int); ok {
			serviceIntegrationConfig.Timeout = timeout
		}

		// Set retry count
		if retryCount, ok := cfg["retry_count"].(int); ok {
			serviceIntegrationConfig.RetryCount = retryCount
		}

		// Set circuit breaker
		if circuitBreaker, ok := cfg["circuit_breaker"].(bool); ok {
			serviceIntegrationConfig.CircuitBreaker = circuitBreaker
		}
	}

	fmt.Printf("🔧 Service Integration Mode: %s\n", mode)
	if mode == ServiceModeMicroservices {
		fmt.Printf("🌐 Service URLs: %v\n", serviceIntegrationConfig.ServiceURLs)
	}
}

// AutoConfigureServiceIntegration configures integration mode from config registry
// Applications should register their own service URLs before calling this
func AutoConfigureServiceIntegration() {
	// Get deployment type from config
	deploymentType := GetConfigString("deployment-type", "monolith")

	switch deploymentType {
	case "monolith":
		SetServiceIntegrationMode(ServiceModeMonolith)
	case "microservices":
		SetServiceIntegrationMode(ServiceModeMicroservices, map[string]any{
			"timeout":         GetConfigInt("service-timeout", 30),
			"retry_count":     GetConfigInt("service-retry-count", 3),
			"circuit_breaker": GetConfigBool("service-circuit-breaker", true),
		})
	case "hybrid":
		SetServiceIntegrationMode(ServiceModeHybrid)
	}
}

// GetServiceIntegrationMode returns current service integration mode
func GetServiceIntegrationMode() ServiceIntegrationMode {
	return serviceIntegrationConfig.Mode
}

// GetServiceURL returns the URL for a service (for microservices mode)
func GetServiceURL(serviceName string) string {
	if url, exists := serviceIntegrationConfig.ServiceURLs[serviceName]; exists {
		return url
	}
	// Generic fallback - applications should register their service URLs
	return fmt.Sprintf("http://localhost:8080")
}

// ServiceProxy creates a service implementation that automatically adapts to integration mode
type ServiceProxy struct {
	serviceName   string
	localService  any
	remoteBaseURL string
	serviceType   reflect.Type
}

// CreateServiceProxy creates a proxy that adapts to current integration mode
func CreateServiceProxy[T any](serviceName string, localImplementation T) T {
	serviceType := reflect.TypeOf((*T)(nil)).Elem()

	proxy := &ServiceProxy{
		serviceName:   serviceName,
		localService:  localImplementation,
		remoteBaseURL: GetServiceURL(serviceName),
		serviceType:   serviceType,
	}

	// Create dynamic proxy based on current mode
	switch serviceIntegrationConfig.Mode {
	case ServiceModeMonolith:
		fmt.Printf("🏢 %s: Local implementation (monolith)\n", serviceName)
		return localImplementation
	case ServiceModeMicroservices:
		fmt.Printf("🔄 %s: HTTP client to %s (microservices)\n", serviceName, proxy.remoteBaseURL)
		// For now, return local implementation (HTTP client would be implemented here)
		// TODO: Create dynamic HTTP client proxy
		return localImplementation
	case ServiceModeHybrid:
		fmt.Printf("🔀 %s: Hybrid mode (smart routing)\n", serviceName)
		// TODO: Implement smart routing logic
		return localImplementation
	default:
		return localImplementation
	}
}

// RegisterServiceInterface registers a service interface with automatic integration
func RegisterServiceInterface[T any](serviceName string, localImpl T) {
	// Register the service normally
	RegisterService(serviceName, localImpl)

	// Also register it as a proxy-enabled service
	proxy := CreateServiceProxy[T](serviceName, localImpl)
	RegisterService(serviceName+"-proxy", proxy, AllowOverride(true))
}

// GetServiceInterface gets a service with automatic integration handling
func GetServiceInterface[T comparable](serviceName string, defaultImpl T) T {
	// Try to get proxy-enabled version first
	if proxyService, ok := TryGetService[T](serviceName+"-proxy", defaultImpl); ok {
		return proxyService
	}

	// Fallback to regular service
	return GetService[T](serviceName, defaultImpl)
}
