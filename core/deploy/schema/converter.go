package schema

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
)

// MiddlewareRegistry defines the interface for resolving middleware names to instances
type MiddlewareRegistry interface {
	GetMiddleware(name string) (request.HandlerFunc, bool)
}

// ToConversionRule converts RouterDef to autogen.ConversionRule
func (r *RouterDef) ToConversionRule() autogen.ConversionRule {
	// Default resource to service name if not specified
	resource := r.Resource
	if resource == "" {
		resource = r.Service
	}

	return autogen.ConversionRule{
		Convention:     convention.ConventionType(r.Convention),
		Resource:       resource,
		ResourcePlural: r.ResourcePlural,
	}
}

// ToRouteOverride converts RouterOverrideDef to autogen.RouteOverride
func (r *RouterOverrideDef) ToRouteOverride(registry MiddlewareRegistry) autogen.RouteOverride {
	override := autogen.RouteOverride{
		PathPrefix:  r.PathPrefix,
		Hidden:      r.Hidden,
		Custom:      make(map[string]autogen.Route),
		Middlewares: resolveMiddlewares(r.Middlewares, registry),
	}

	// Convert Custom array to map
	for _, routeDef := range r.Custom {
		override.Custom[routeDef.Name] = autogen.Route{
			Method:      routeDef.Method,
			Path:        routeDef.Path,
			Middlewares: resolveMiddlewares(routeDef.Middlewares, registry),
		}
	}

	return override
}

// resolveMiddlewares resolves middleware names to instances using the registry
func resolveMiddlewares(names []string, registry MiddlewareRegistry) []any {
	if len(names) == 0 {
		return nil
	}

	middlewares := make([]any, 0, len(names))
	for _, name := range names {
		if mw, found := registry.GetMiddleware(name); found {
			middlewares = append(middlewares, mw)
		} else {
			// If middleware not found, you might want to log or panic
			// For now, we skip it
			// log.Printf("Warning: middleware '%s' not found in registry", name)
		}
	}

	return middlewares
}

// ToHandlerFunc is a helper to ensure middleware is converted to request.HandlerFunc
// This is useful when working with middleware chains
func ToHandlerFunc(middleware any) request.HandlerFunc {
	if mw, ok := middleware.(request.HandlerFunc); ok {
		return mw
	}
	if mw, ok := middleware.(func(*request.Context) error); ok {
		return request.HandlerFunc(mw)
	}
	return nil
}
