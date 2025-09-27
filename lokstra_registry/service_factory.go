package lokstra_registry

type ServiceFactory = func(config map[string]any) any

var serviceFactoryRegistry = make(map[string]ServiceFactory)

// Registers a service factory function for a given service type.
// If the service type is already registered
// and the RegisterOption allowOverride is not set to true, it will panic.
func RegisterServiceFactory(serviceType string, factory ServiceFactory,
	opts ...RegisterOption) {
	var options registerOptions
	for _, opt := range opts {
		opt.apply(&options)
	}
	if !options.allowOverride {
		if _, exists := serviceFactoryRegistry[serviceType]; exists {
			panic("service factory for type " + serviceType + " already registered")
		}
	}
	serviceFactoryRegistry[serviceType] = factory
}

// Retrieves a registered service factory function by service type.
func GetServiceFactory(serviceType string) ServiceFactory {
	if factory, ok := serviceFactoryRegistry[serviceType]; ok {
		return factory
	}
	return nil
}
