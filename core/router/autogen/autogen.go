package autogen

import (
	"fmt"
	"path"
	"reflect"
	"slices"
	"strings"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router/convention"
)

// ConversionRule defines how to convert service methods to routes
type ConversionRule struct {
	Convention     convention.ConventionType
	Resource       string // singular form, e.g., "user"
	ResourcePlural string // plural form, e.g., "users"
}

// RouteOverride defines custom route configuration
type RouteOverride struct {
	PathPrefix  string           // e.g., "/api/v1"
	Hidden      []string         // methods to hide
	Custom      map[string]Route // custom route definitions
	Middlewares []any            // middlewares to apply
}

// Route defines a custom route
type Route struct {
	Method      string
	Path        string
	Middlewares []any // middlewares specific to this route
}

// NewFromService creates a router by reflecting on a service interface
// and applying conversion rules and overrides
func NewFromService(service any, rule ConversionRule, override RouteOverride) lokstra.Router {
	router := lokstra.NewRouter(fmt.Sprintf("%s-auto", rule.Resource))

	// Get service type - keep as pointer type to access methods with pointer receivers
	serviceType := reflect.TypeOf(service)

	// Iterate through methods
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		methodName := method.Name

		// Skip if hidden
		if slices.Contains(override.Hidden, methodName) {
			continue
		}

		// Check for custom route
		if customRoute, ok := override.Custom[methodName]; ok {
			registerCustomRoute(router, service, methodName, customRoute, override.PathPrefix, override.Middlewares)
			continue
		}

		// Use convention registry to resolve method
		conv := convention.MustGet(rule.Convention)
		httpMethod, pathTemplate, found := conv.ResolveMethod(methodName, rule.Resource, rule.ResourcePlural)

		if !found {
			// Skip unknown methods
			continue
		}

		// Apply path prefix using path.Join to avoid double slashes
		fullPath := pathTemplate
		if override.PathPrefix != "" {
			fullPath = path.Join(override.PathPrefix, pathTemplate)
		}

		// Register the route with global middlewares
		registerRoute(router, httpMethod, fullPath, service, methodName, override.Middlewares)
	}

	return router
}

// registerCustomRoute registers a custom route
func registerCustomRoute(router lokstra.Router, service any, methodName string, route Route, pathPrefix string, globalMiddlewares []any) {
	// Apply path prefix using path.Join to avoid double slashes
	fullPath := route.Path
	if pathPrefix != "" {
		fullPath = path.Join(pathPrefix, route.Path)
	}

	// If method is empty, auto-detect from method name
	httpMethod := route.Method
	if httpMethod == "" {
		// Auto-detect: Get*, List* → GET, Create* → POST, Update* → PUT, Delete* → DELETE
		if strings.HasPrefix(methodName, "Get") || strings.HasPrefix(methodName, "List") {
			httpMethod = "GET"
		} else if strings.HasPrefix(methodName, "Create") {
			httpMethod = "POST"
		} else if strings.HasPrefix(methodName, "Update") {
			httpMethod = "PUT"
		} else if strings.HasPrefix(methodName, "Delete") {
			httpMethod = "DELETE"
		} else {
			// Default to POST for unknown patterns
			httpMethod = "POST"
		}
	}

	// Merge middlewares: global middlewares + route-specific middlewares
	var middlewares []any
	middlewares = append(middlewares, globalMiddlewares...)
	middlewares = append(middlewares, route.Middlewares...)

	registerRoute(router, httpMethod, fullPath, service, methodName, middlewares)
}

// registerRoute registers a route by calling the appropriate router method
func registerRoute(r lokstra.Router, httpMethod, path string, service any, methodName string, middlewares []any) {
	// Get the method from service
	serviceValue := reflect.ValueOf(service)
	method := serviceValue.MethodByName(methodName)

	if !method.IsValid() {
		return
	}

	// Pass the method directly to the router
	// The router's adaptHandler will automatically handle ~29 different handler signatures:
	// - func(*Context) error
	// - func(*Context) (any, error)
	// - func(*Context, *Struct) error
	// - func(*Struct) (any, error)
	// - etc.
	handler := method.Interface()
	middlewares = append(middlewares, route.WithNameOption(methodName))

	// Register based on HTTP method with middlewares
	switch httpMethod {
	case "GET":
		r.GET(path, handler, middlewares...)
	case "POST":
		r.POST(path, handler, middlewares...)
	case "PUT":
		r.PUT(path, handler, middlewares...)
	case "DELETE":
		r.DELETE(path, handler, middlewares...)
	case "PATCH":
		r.PATCH(path, handler, middlewares...)
	}
}
