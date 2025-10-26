package router

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
)

var (
	// typeOfContext is used to detect *request.Context parameters
	// Works with type aliases like: type RequestContext = request.Context
	typeOfContextPtr = reflect.TypeOf((*request.Context)(nil))
)

// NewFromService creates a Router by auto-generating routes from service methods using conventions.
// It uses the default engine and applies convention-based method name parsing.
//
// Convention rules:
//   - Get{Resource} -> GET /{resources}/{id}
//   - List{Resources} -> GET /{resources}
//   - Create{Resource} -> POST /{resources}
//   - Update{Resource} -> PUT /{resources}/{id}
//   - Delete{Resource} -> DELETE /{resources}/{id}
//
// Example:
//
//	type UserService struct{}
//	func (s *UserService) GetUser(ctx *request.Context, id string) (*User, error) { ... }
//	func (s *UserService) ListUsers(ctx *request.Context) ([]*User, error) { ... }
//
//	router := router.NewFromService(&UserService{}, router.DefaultServiceRouterOptions())
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

	// Determine resource names
	resourceName := opts.ResourceName
	if resourceName == "" {
		resourceName = extractResourceNameFromType(structType.Name())
	}

	pluralResourceName := opts.PluralResourceName
	if pluralResourceName == "" {
		pluralResourceName = pluralizeName(resourceName)
	}

	// Create convention parser
	parser := NewConventionParser(resourceName, pluralResourceName)

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

		// Check for route override
		if override, exists := opts.RouteOverrides[method.Name]; exists {
			handler := createServiceMethodHandler(service, method)
			path := override.Path
			httpMethod := override.HTTPMethod

			if path == "" {
				// Use convention to generate path
				genMethod, genPath, err := parser.ParseMethodName(method.Name)
				if err != nil {
					continue
				}
				path = genPath
				if httpMethod == "" {
					httpMethod = genMethod
				}
			}

			// Apply prefix
			if opts.Prefix != "" {
				path = strings.TrimSuffix(opts.Prefix, "/") + "/" + strings.TrimPrefix(path, "/")
			}

			// Ensure MethodName is set for route naming
			if override.MethodName == "" {
				override.MethodName = method.Name
			}

			// Register route with override meta
			registerRouteByMethod(r, httpMethod, path, handler, override)
			continue
		}

		// Parse method name using conventions
		if opts.DisableConventions {
			continue
		}

		// Extract action from method name
		action := parser.extractAction(method.Name)
		if action == "" {
			continue
		}

		// Detect struct parameter with path tags
		structType := detectStructParameter(method.Type)
		var path string
		var httpMethod string

		if structType != nil {
			// Generate path from struct tags
			httpMethod = parser.actionToHTTPMethod(action)
			path = parser.GeneratePathFromStruct(action, structType)
		} else {
			// No struct parameter - use default convention (simple case)
			var err error
			httpMethod, path, err = parser.ParseMethodName(method.Name)
			if err != nil {
				continue
			}
		}

		// Apply prefix
		if opts.Prefix != "" {
			path = strings.TrimSuffix(opts.Prefix, "/") + "/" + strings.TrimPrefix(path, "/")
		}

		// Create handler
		handler := createServiceMethodHandler(service, method)

		// Create basic RouteMeta for convention-based routes
		meta := RouteMeta{
			MethodName: method.Name,
		}

		// Register route with method name
		registerRouteByMethod(r, httpMethod, path, handler, meta)
	}

	return r
}

// extractResourceNameFromType extracts resource name from service type name
// Example: "UserService" -> "user", "ProductService" -> "product"
func extractResourceNameFromType(typeName string) string {
	// Remove "Service" suffix
	name := strings.TrimSuffix(typeName, "Service")
	if name == "" {
		name = typeName
	}

	// Convert to lowercase
	return strings.ToLower(name)
}

// pluralizeName simple pluralization
func pluralizeName(name string) string {
	if strings.HasSuffix(name, "s") || strings.HasSuffix(name, "x") ||
		strings.HasSuffix(name, "z") || strings.HasSuffix(name, "ch") ||
		strings.HasSuffix(name, "sh") {
		return name + "es"
	}
	if strings.HasSuffix(name, "y") {
		return name[:len(name)-1] + "ies"
	}
	return name + "s"
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

// detectStructParameter checks if method has a struct parameter (excluding context)
// Returns the struct type if found, nil otherwise
//
// Uses type comparison to detect *request.Context, which works correctly with type aliases:
//   - type RequestContext = request.Context
//   - type MyContext = request.Context
//
// Supported signatures:
//   - func(req *Struct) error
//   - func(req *Struct) (data, error)
//   - func(ctx *Context, req *Struct) error
//   - func(ctx *Context, req *Struct) (data, error)
func detectStructParameter(methodType reflect.Type) reflect.Type {
	numIn := methodType.NumIn()

	// Skip receiver (index 0)
	for i := 1; i < numIn; i++ {
		paramType := methodType.In(i)

		// Skip *request.Context using type comparison (not name comparison!)
		// This works with type aliases: type MyContext = request.Context
		if paramType == typeOfContextPtr {
			continue
		}

		// Check if it's a struct pointer (our request struct)
		if paramType.Kind() == reflect.Pointer {
			elemType := paramType.Elem()
			if elemType.Kind() == reflect.Struct {
				return paramType
			}
		}

		// Check for non-pointer struct
		if paramType.Kind() == reflect.Struct {
			return paramType
		}
	}

	return nil
}
