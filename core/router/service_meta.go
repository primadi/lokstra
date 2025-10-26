package router

import "reflect"

// ServiceMeta contains metadata for service-to-router mapping
type ServiceMeta struct {
	// Prefix is prepended to all routes from this service
	// Example: "/api/v1"
	Prefix string

	// Name is the service name (optional, defaults to struct name)
	Name string

	// Tags for additional metadata
	Tags map[string]string
}

// RouteMeta contains metadata for a specific service method route
type RouteMeta struct {
	// Method name from service (e.g., "GetUser")
	MethodName string

	// HTTP method (GET, POST, PUT, DELETE, PATCH)
	// Auto-detected from method name if empty
	HTTPMethod string

	// Path for this route (e.g., "/users/{id}")
	// Auto-generated from method name if empty
	Path string

	// Name of the route (for identification and introspection)
	// If empty, defaults to method name
	Name string

	// Description of the route
	Description string

	// OverrideParentMw determines if this route should override parent middleware
	// Default is false (parent middleware will be applied)
	OverrideParentMw bool

	// Middlewares for this specific route
	// If not empty, these middlewares will be applied to this route only
	// Can be middleware names (string) or middleware functions
	Middlewares []any

	// Query parameters expected
	QueryParams []string

	// Headers expected
	Headers []string

	// Auth requirement
	AuthRequired bool

	// Custom tags
	Tags map[string]string
}

// ServiceRouterOptions configures service-to-router conversion
type ServiceRouterOptions struct {
	// ConventionName specifies which convention to use (e.g., "rest", "rpc")
	// If empty, uses the default convention (typically "rest")
	// Convention must be registered in lokstra_registry
	ConventionName string

	// Prefix for all routes (e.g., "/api/v1")
	Prefix string

	// ResourceName override (auto-detected from service name if empty)
	// Example: "UserService" → "user"
	ResourceName string

	// PluralResourceName for list endpoints (auto-pluralized if empty)
	// Example: "user" → "users"
	PluralResourceName string

	// DisableConventions uses only explicit RouteMeta
	DisableConventions bool

	// RouteOverrides allows overriding specific method routes
	// Key: method name, Value: RouteMeta
	RouteOverrides map[string]RouteMeta

	// Middlewares to apply to all routes
	Middlewares []string
}

// ServiceMethodInfo holds reflection info about a service method
type ServiceMethodInfo struct {
	Method      reflect.Method
	Name        string
	NumIn       int  // Number of input params
	NumOut      int  // Number of return values
	HasContext  bool // First param is *RequestContext
	HasError    bool // Last return is error
	ParamType   reflect.Type
	ReturnType  reflect.Type
	ErrorReturn int // Index of error return (-1 if none)
}

// DefaultServiceRouterOptions returns default options
func DefaultServiceRouterOptions() *ServiceRouterOptions {
	return &ServiceRouterOptions{
		ConventionName:     "", // Empty means use default convention
		Prefix:             "",
		DisableConventions: false,
		RouteOverrides:     make(map[string]RouteMeta),
	}
}

// WithConvention sets the convention name
func (o *ServiceRouterOptions) WithConvention(conventionName string) *ServiceRouterOptions {
	o.ConventionName = conventionName
	return o
}

// WithPrefix sets the route prefix
func (o *ServiceRouterOptions) WithPrefix(prefix string) *ServiceRouterOptions {
	o.Prefix = prefix
	return o
}

// WithResourceName sets the resource name
func (o *ServiceRouterOptions) WithResourceName(name string) *ServiceRouterOptions {
	o.ResourceName = name
	return o
}

// WithPluralResourceName sets the plural resource name
func (o *ServiceRouterOptions) WithPluralResourceName(name string) *ServiceRouterOptions {
	o.PluralResourceName = name
	return o
}

// WithRouteOverride adds a route override for a specific method
func (o *ServiceRouterOptions) WithRouteOverride(methodName string, meta RouteMeta) *ServiceRouterOptions {
	if o.RouteOverrides == nil {
		o.RouteOverrides = make(map[string]RouteMeta)
	}
	o.RouteOverrides[methodName] = meta
	return o
}

// WithMethodMiddleware adds middleware for a specific method without changing route
// This is a convenience method that creates a RouteOverride with only middleware
func (o *ServiceRouterOptions) WithMethodMiddleware(methodName string, middleware ...any) *ServiceRouterOptions {
	if o.RouteOverrides == nil {
		o.RouteOverrides = make(map[string]RouteMeta)
	}
	existing, exists := o.RouteOverrides[methodName]
	if exists {
		// Append to existing middleware
		existing.Middlewares = append(existing.Middlewares, middleware...)
	} else {
		// Create new override with only middleware (Path and HTTPMethod are empty, will use convention)
		existing = RouteMeta{
			MethodName:  methodName,
			Middlewares: middleware,
		}
	}
	o.RouteOverrides[methodName] = existing
	return o
}

// WithMiddlewares sets middlewares for all routes
func (o *ServiceRouterOptions) WithMiddlewares(mws ...string) *ServiceRouterOptions {
	o.Middlewares = mws
	return o
}

// WithoutConventions disables convention-based route generation
func (o *ServiceRouterOptions) WithoutConventions() *ServiceRouterOptions {
	o.DisableConventions = true
	return o
}
