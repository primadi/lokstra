package router

import (
	"fmt"
	"reflect"
	"sync"
)

// ServiceConvention defines how a service is converted to router routes and client router methods
type ServiceConvention interface {
	// Name returns the convention name (e.g., "rest", "rpc", "graphql")
	Name() string

	// GenerateRoutes generates route metadata from service methods
	// Returns: map[methodName]RouteMeta
	GenerateRoutes(serviceType reflect.Type, options ServiceRouterOptions) (map[string]RouteMeta, error)

	// GenerateClientMethod generates client method signature from service method
	// This is used for auto-generating client router methods
	GenerateClientMethod(method ServiceMethodInfo, options ServiceRouterOptions) (ClientMethodMeta, error)
}

// ClientMethodMeta contains metadata for generating client router methods
type ClientMethodMeta struct {
	// MethodName is the original service method name
	MethodName string

	// HTTPMethod is the HTTP verb (GET, POST, PUT, DELETE, PATCH)
	HTTPMethod string

	// Path is the URL path (e.g., "/users/{id}")
	Path string

	// HasBody indicates if the request has a body (POST, PUT, PATCH)
	HasBody bool

	// BodyParam is the parameter to use as request body
	BodyParam string

	// PathParams are parameters from the URL path
	PathParams []string

	// QueryParams are parameters from the query string
	QueryParams []string

	// Headers are expected headers
	Headers []string
}

var (
	// conventionRegistry uses sync.Map for better concurrent read performance
	// Conventions are registered once at startup, read many times - optimal for sync.Map
	conventionRegistry sync.Map
	defaultConvention  = "rest" // Default convention name
)

// RegisterConvention registers a new service convention
func RegisterConvention(convention ServiceConvention) error {
	if convention == nil {
		return fmt.Errorf("convention cannot be nil")
	}

	name := convention.Name()
	if name == "" {
		return fmt.Errorf("convention name cannot be empty")
	}

	if _, exists := conventionRegistry.Load(name); exists {
		return fmt.Errorf("convention '%s' already registered", name)
	}

	conventionRegistry.Store(name, convention)
	return nil
}

// GetConvention retrieves a registered convention by name
func GetConvention(name string) (ServiceConvention, error) {
	if v, ok := conventionRegistry.Load(name); ok {
		return v.(ServiceConvention), nil
	}
	return nil, fmt.Errorf("convention '%s' not found", name)
}

// GetDefaultConvention returns the default convention (REST)
func GetDefaultConvention() (ServiceConvention, error) {
	return GetConvention(defaultConvention)
}

// SetDefaultConvention sets the default convention name
func SetDefaultConvention(name string) error {
	if _, ok := conventionRegistry.Load(name); !ok {
		return fmt.Errorf("convention '%s' not registered", name)
	}
	defaultConvention = name
	return nil
}

// ListConventions returns names of all registered conventions
func ListConventions() []string {
	names := make([]string, 0)
	conventionRegistry.Range(func(key, value any) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

// MustRegisterConvention registers a convention and panics on error
// Use this for built-in conventions during init()
func MustRegisterConvention(convention ServiceConvention) {
	if err := RegisterConvention(convention); err != nil {
		panic(fmt.Sprintf("failed to register convention: %v", err))
	}
}
