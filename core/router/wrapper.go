package router

// ApplyMiddlewares applies additional middlewares to an existing router.
// This is useful for applying deployment-specific middlewares to manually registered routers.
//
// The middlewares are prepended (run before) any existing middlewares in the router.
// Middleware execution order: new middlewares → existing middlewares → handler
//
// Supported middleware types:
//   - func(*lokstra.RequestContext) error
//   - request.HandlerFunc
//   - func(*lokstra.RequestContext, any) error
//   - string (middleware name from registry)
//
// Example:
//
//	// Manual router with its own middlewares
//	r := router.New("/api/v1")
//	r.Use("logger", "recovery")
//	r.GET("/users", handler.List)
//	lokstra_registry.RegisterRouter("my-router", r)
//
//	// Later, apply additional middlewares from YAML config
//	router.ApplyMiddlewares(r, "auth", "rate-limiter")
//	// Execution order: auth → rate-limiter → logger → recovery → handler
//
// Use cases:
//   - Apply environment-specific middlewares (auth only in prod)
//   - Add monitoring/logging per deployment
//   - Inject rate limiting from configuration
//
// Note: This modifies the router in-place. Call this before the router is used.
func ApplyMiddlewares(r Router, middlewares ...any) {
	if len(middlewares) == 0 {
		return
	}
	r.Use(middlewares...)
}
