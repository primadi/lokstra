package convention

import (
	"fmt"
)

// ConventionType represents the routing convention to use
type ConventionType = string

const (
	REST    ConventionType = "rest"
	RPC     ConventionType = "rpc"
	GraphQL ConventionType = "graphql"
)

// RouteMapping defines how a method maps to HTTP route
type RouteMapping struct {
	HTTPMethod   string // GET, POST, PUT, DELETE, etc.
	PathTemplate string // e.g., "/{resource-plural}", "/{resource-plural}/{id}"
}

// Convention defines the interface for routing conventions
type Convention interface {
	// Name returns the convention type
	Name() ConventionType

	// ResolveMethod maps a service method name to HTTP method and path template
	// Returns (httpMethod, pathTemplate, found)
	ResolveMethod(methodName string, resource string, resourcePlural string) (httpMethod string, pathTemplate string, found bool)
}

// Global registry
var registry = make(map[ConventionType]Convention)

// Register adds a convention to the registry
func Register(conv Convention) {
	registry[conv.Name()] = conv
}

// Get retrieves a convention from the registry
// Empty string defaults to REST convention
func Get(name ConventionType) (Convention, error) {
	// Default to REST if empty
	if name == "" {
		name = REST
	}

	conv, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("convention not found: %s", name)
	}
	return conv, nil
}

// MustGet retrieves a convention or panics
func MustGet(name ConventionType) Convention {
	conv, err := Get(name)
	if err != nil {
		panic(err)
	}
	return conv
}

func init() {
	// Register built-in conventions
	Register(&RESTConvention{})
	Register(&RPCConvention{})
}
