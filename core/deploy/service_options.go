package deploy

import "strings"

// ServiceTypeConfig is a structured configuration for service type registration
// This provides a cleaner, more maintainable alternative to functional options
type ServiceTypeConfig struct {
	// Basic metadata
	Resource       string // Singular resource name (e.g., "user")
	ResourcePlural string // Plural resource name (e.g., "users")
	Convention     string // Convention type (e.g., "rest", "rpc", "graphql")

	// Router-level configuration
	PathPrefix  string   // Path prefix for all routes (e.g., "/api/v1")
	Middlewares []string // Middleware names to apply to all routes

	// Route filtering
	Hidden []string // Method names to hide from auto-generated router

	// Custom route overrides with full metadata support
	RouteOverrides map[string]RouteConfig
}

// RouteConfig defines custom configuration for a specific route
// Supports both path override and route-level middlewares
type RouteConfig struct {
	Method      string   // HTTP method (e.g., "POST", "GET") - auto-detected if empty
	Path        string   // Custom path (e.g., "/auth/login", "/users/{id}/orders")
	Middlewares []string // Route-specific middleware names (in addition to router-level)
}

// RegisterServiceTypeOption configures service type registration (legacy functional options)
// Deprecated: Use ServiceTypeConfig struct for better readability
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
// Deprecated: Use ServiceTypeConfig.RouteOverrides for better control including route-level middlewares
func WithRouteOverride(methodName, pathSpec string) RegisterServiceTypeOption {
	return func(m *ServiceMetadata) {
		if m.RouteOverrides == nil {
			m.RouteOverrides = make(map[string]RouteMetadata)
		}

		// Parse pathSpec: "POST /path" or "/path"
		parts := strings.SplitN(strings.TrimSpace(pathSpec), " ", 2)
		method := ""
		path := pathSpec

		if len(parts) == 2 {
			possibleMethod := strings.ToUpper(parts[0])
			switch possibleMethod {
			case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
				method = possibleMethod
				path = strings.TrimSpace(parts[1])
			}
		}

		m.RouteOverrides[methodName] = RouteMetadata{
			Method: method,
			Path:   path,
		}
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
