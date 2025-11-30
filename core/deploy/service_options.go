package deploy

// ServiceTypeConfig is a structured configuration for service type registration
type ServiceTypeConfig struct {
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
