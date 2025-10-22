package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/router/autogen"
)

// RemoteServiceMeta defines metadata for a remote service.
// This interface allows services to provide routing information
// for proxy generation without tight coupling.
//
// Services should implement this interface to provide:
//   - Resource naming (singular and plural forms)
//   - Convention to use for route generation
//   - Custom route overrides
//
// This is typically implemented by remote service wrappers,
// not by the service interface itself.
type RemoteServiceMeta interface {
	// GetResourceName returns (singular, plural) resource names
	// Example: ("user", "users"), ("order", "orders")
	// The plural form is used for collection endpoints (List, etc.)
	GetResourceName() (string, string)

	// GetConventionName returns the convention name to use
	// Example: "rest", "jsonapi", "custom"
	// This determines how methods are mapped to HTTP routes
	GetConventionName() string

	// GetRouteOverride returns custom route overrides
	// This allows per-service customization of routes,
	// including path prefix and custom method routes
	GetRouteOverride() autogen.RouteOverride
}

// RemoteServiceMetaAdapter provides default implementation of RemoteServiceMeta.
// Services can embed this struct and only override what they need.
//
// Example usage:
//
//	type UserServiceRemote struct {
//	    service.RemoteServiceMetaAdapter
//	}
//
//	func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
//	    return &UserServiceRemote{
//	        RemoteServiceMetaAdapter: service.RemoteServiceMetaAdapter{
//	            Resource:   "user",
//	            Plural:     "users",
//	            Convention: "rest",
//	            ProxyService: proxyService,
//	        },
//	    }
//	}
type RemoteServiceMetaAdapter struct {
	// Resource is the singular resource name (e.g., "user", "order")
	Resource string

	// Plural is the plural resource name (e.g., "users", "orders")
	// If empty, defaults to Resource + "s"
	Plural string

	// Convention is the convention name (e.g., "rest", "jsonapi")
	// If empty, defaults to "rest"
	Convention string

	// Override contains custom route overrides
	Override autogen.RouteOverride

	// ProxyService is the proxy.Service instance for making HTTP calls
	// This is set by the framework when creating remote service instances
	ProxyService *proxy.Service
}

// GetProxyService returns the proxy.Service instance
func (a *RemoteServiceMetaAdapter) GetProxyService() *proxy.Service {
	return a.ProxyService
}

// GetResourceName implements RemoteServiceMeta
func (a *RemoteServiceMetaAdapter) GetResourceName() (string, string) {
	plural := a.Plural
	if plural == "" {
		plural = a.Resource + "s"
	}
	return a.Resource, plural
}

// GetConventionName implements RemoteServiceMeta
func (a *RemoteServiceMetaAdapter) GetConventionName() string {
	if a.Convention == "" {
		return "rest" // default convention
	}
	return a.Convention
}

// GetRouteOverride implements RemoteServiceMeta
func (a *RemoteServiceMetaAdapter) GetRouteOverride() autogen.RouteOverride {
	return a.Override
}

// Ensure RemoteServiceMetaAdapter implements RemoteServiceMeta
var _ RemoteServiceMeta = (*RemoteServiceMetaAdapter)(nil)
