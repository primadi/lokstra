package router

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/config"
	"github.com/primadi/lokstra/core/route"
)

// NewFromService creates a Router by registering service methods with explicit route definitions.
// Each method MUST have a route override in opts.RouteOverrides to be registered.
// Methods without explicit routes will be skipped.
//
// Example:
//
//	type UserService struct{}
//	func (s *UserService) GetUser(ctx *request.Context, id string) (*User, error) { ... }
//	func (s *UserService) ListUsers(ctx *request.Context) ([]*User, error) { ... }
//
//	opts := &router.ServiceRouterOptions{
//	    RouteOverrides: map[string]router.RouteMeta{
//	        "GetUser":   {HTTPMethod: "GET", Path: "/users/{id}"},
//	        "ListUsers": {HTTPMethod: "GET", Path: "/users"},
//	    },
//	}
//	router := router.NewFromService(&UserService{}, opts)
func NewFromService(service any, opts *ServiceRouterOptions) Router {
	return NewFromServiceWithEngine(service, "default", opts)
}

// NewFromServiceWithEngine creates a Router with a custom engine type.
// Allows using specific engine types like "default", "servemux", etc.
func NewFromServiceWithEngine(service any, engineType string, opts *ServiceRouterOptions) Router {
	// Extract resource name for router name
	serviceType := reflect.TypeOf(service)

	// Get the struct type for name extraction, but keep pointer type for method scanning
	structType := serviceType
	if serviceType.Kind() == reflect.Pointer {
		structType = serviceType.Elem()
	}
	routerName := structType.Name()

	// Create router
	r := NewWithEngine(routerName, engineType)

	// Scan all methods and register routes
	// IMPORTANT: Use the original serviceType (pointer) to get all methods
	// because methods with pointer receivers are only visible on the pointer type
	numMethods := serviceType.NumMethod()
	for i := range numMethods {
		method := serviceType.Method(i)

		// Skip non-exported methods
		if !method.IsExported() {
			continue
		}

		// Only process methods with explicit route overrides
		override, exists := opts.RouteOverrides[method.Name]
		if !exists {
			// Skip methods without explicit route definition
			continue
		}

		// Path is required
		if override.Path == "" {
			continue
		}

		// Create handler
		handler := createServiceMethodHandler(service, method)

		// Build path (apply prefix first, then resolve variables)
		path := override.Path
		if opts.Prefix != "" {
			path = strings.TrimSuffix(opts.Prefix, "/") + "/" + strings.TrimPrefix(path, "/")
		}

		// Expand variables in final path at runtime
		// Resolves ${key} or ${key:default} via lokstra_registry.GetConfig()
		path = config.SimpleResolver(path)

		// HTTP method is required
		httpMethod := override.HTTPMethod
		if httpMethod == "" {
			// Default to GET if not specified
			httpMethod = "GET"
		}

		// Ensure MethodName is set for route naming
		if override.MethodName == "" {
			override.MethodName = method.Name
		}

		// Register route with override meta
		registerRouteByMethod(r, httpMethod, path, handler, override)
	}

	return r
}

// registers a route based on HTTP method with route options
// Converts RouteMeta to route.RouteHandlerOption(s) and applies them along with middleware
func registerRouteByMethod(r Router, httpMethod, path string, handler any, meta RouteMeta) {
	// Build route options from RouteMeta
	var options []any

	// Add name option (use RouteMeta.Name if provided, otherwise use MethodName)
	if meta.Name != "" {
		options = append(options, route.WithNameOption(meta.Name))
	} else if meta.MethodName != "" {
		options = append(options, route.WithNameOption(meta.MethodName))
	}

	// Add description option if provided
	if meta.Description != "" {
		options = append(options, route.WithDescriptionOption(meta.Description))
	}

	// Add override parent middleware option if set
	if meta.OverrideParentMw {
		options = append(options, route.WithOverrideParentMwOption(true))
	}

	// Add middlewares
	options = append(options, meta.Middlewares...)

	switch strings.ToUpper(httpMethod) {
	case "GET":
		r.GET(path, handler, options...)
	case "POST":
		r.POST(path, handler, options...)
	case "PUT":
		r.PUT(path, handler, options...)
	case "PATCH":
		r.PATCH(path, handler, options...)
	case "DELETE":
		r.DELETE(path, handler, options...)
	default:
		// Fallback to ANY
		r.ANY(path, handler, options...)
	}
} // creates a handler function that calls the service method
func createServiceMethodHandler(service any, method reflect.Method) any {
	serviceValue := reflect.ValueOf(service)
	methodValue := serviceValue.MethodByName(method.Name)

	if !methodValue.IsValid() {
		panic(fmt.Sprintf("method %s not found on service", method.Name))
	}

	// Return the method as-is! Router's adaptSmart will detect and adapt it
	return methodValue.Interface()
}
