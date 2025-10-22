package deploy

// RegisterServiceTypeOption configures service type registration
type RegisterServiceTypeOption func(*ServiceMetadata)

// WithResource sets the resource name (singular and plural)
func WithResource(singular, plural string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		m.Resource = singular
		m.ResourcePlural = plural
	}
}

// WithConvention sets the convention type (default: "rest")
func WithConvention(convention string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		m.Convention = convention
	}
}

// WithRouteOverride adds a custom route path for a method
func WithRouteOverride(methodName, path string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		if m.RouteOverrides == nil {
			m.RouteOverrides = make(map[string]string)
		}
		m.RouteOverrides[methodName] = path
	}
}

// WithHiddenMethods hides methods from auto-generated router
func WithHiddenMethods(methods ...string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		m.HiddenMethods = append(m.HiddenMethods, methods...)
	}
}

// WithPathPrefix sets a path prefix for all routes
func WithPathPrefix(prefix string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		m.PathPrefix = prefix
	}
}

// WithMiddlewares adds middleware names to apply to all routes
func WithMiddlewares(middlewares ...string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		m.MiddlewareNames = append(m.MiddlewareNames, middlewares...)
	}
}
