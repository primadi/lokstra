package lokstra_registry

import (
	"fmt"
	"reflect"
)

// ServiceDefinition defines a service interface and its implementations
type ServiceDefinition struct {
	Name                string
	Interface           reflect.Type
	LocalImplementation interface{}
	RemoteClientFactory func(baseURL string) interface{}
}

var serviceDefinitions = make(map[string]*ServiceDefinition)

// DefineService registers a service definition with local and remote implementations
func DefineService(name string, localImpl interface{}, remoteClientFactory func(string) interface{}) {
	definition := &ServiceDefinition{
		Name:                name,
		Interface:           reflect.TypeOf(localImpl),
		LocalImplementation: localImpl,
		RemoteClientFactory: remoteClientFactory,
	}

	serviceDefinitions[name] = definition
	fmt.Printf("📋 Defined service: %s\n", name)
}

// AutoRegisterServices automatically registers services based on current integration mode
func AutoRegisterServices() {
	fmt.Println("🔄 Auto-registering services based on integration mode...")

	for serviceName, definition := range serviceDefinitions {
		registerServiceFromDefinition(serviceName, definition)
	}
}

func registerServiceFromDefinition(serviceName string, def *ServiceDefinition) {
	switch serviceIntegrationConfig.Mode {
	case ServiceModeMonolith:
		fmt.Printf("🏢 Registering %s: Local implementation\n", serviceName)
		RegisterService(serviceName, def.LocalImplementation, AllowOverride(true))

	case ServiceModeMicroservices:
		baseURL := GetServiceURL(serviceName)
		if def.RemoteClientFactory != nil {
			remoteClient := def.RemoteClientFactory(baseURL)
			fmt.Printf("🔄 Registering %s: HTTP client to %s\n", serviceName, baseURL)
			RegisterService(serviceName, remoteClient, AllowOverride(true))
		} else {
			fmt.Printf("⚠️  %s: No remote client factory, using local fallback\n", serviceName)
			RegisterService(serviceName, def.LocalImplementation, AllowOverride(true))
		}

	case ServiceModeHybrid:
		// TODO: Implement hybrid logic
		fmt.Printf("🔀 Registering %s: Hybrid mode (using local for now)\n", serviceName)
		RegisterService(serviceName, def.LocalImplementation, AllowOverride(true))
	}
}

func registerServiceViaReflection(serviceName string, definition any) {
	// This handles the generic ServiceDefinition[T] case via reflection
	v := reflect.ValueOf(definition)
	if v.Kind() != reflect.Struct {
		return
	}

	localImplField := v.FieldByName("LocalImplementation")
	if !localImplField.IsValid() {
		return
	}

	switch serviceIntegrationConfig.Mode {
	case ServiceModeMonolith:
		fmt.Printf("🏢 Registering %s: Local implementation (reflection)\n", serviceName)
		RegisterService(serviceName, localImplField.Interface(), AllowOverride(true))

	case ServiceModeMicroservices:
		// Check if RemoteClientFactory exists
		factoryField := v.FieldByName("RemoteClientFactory")
		if factoryField.IsValid() && !factoryField.IsNil() {
			// Call the factory with base URL
			baseURL := GetServiceURL(serviceName)
			factory := factoryField.Interface()

			// Call factory function via reflection
			factoryValue := reflect.ValueOf(factory)
			if factoryValue.Kind() == reflect.Func {
				results := factoryValue.Call([]reflect.Value{reflect.ValueOf(baseURL)})
				if len(results) > 0 {
					remoteClient := results[0].Interface()
					fmt.Printf("🔄 Registering %s: HTTP client to %s (reflection)\n", serviceName, baseURL)
					RegisterService(serviceName, remoteClient, AllowOverride(true))
					return
				}
			}
		}

		// Fallback to local implementation
		fmt.Printf("⚠️  %s: No remote client factory, using local fallback (reflection)\n", serviceName)
		RegisterService(serviceName, localImplField.Interface(), AllowOverride(true))

	case ServiceModeHybrid:
		fmt.Printf("🔀 Registering %s: Hybrid mode (using local for now, reflection)\n", serviceName)
		RegisterService(serviceName, localImplField.Interface(), AllowOverride(true))
	}
}

// SmartGetService gets service with automatic integration handling
func SmartGetService(serviceName string, fallback interface{}) interface{} {
	// Get the registered service (which was auto-registered based on integration mode)
	return GetService(serviceName, fallback)
}

// ServiceIntegrationMiddleware can be used to configure integration at server start
func ServiceIntegrationMiddleware() {
	fmt.Println("🚀 Initializing Service Integration...")

	// Auto-configure based on current config
	AutoConfigureServiceIntegration()

	// Auto-register all defined services
	AutoRegisterServices()

	fmt.Printf("✅ Service Integration initialized in %s mode\n", serviceIntegrationConfig.Mode)
}
