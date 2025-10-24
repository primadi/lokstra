package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/router/autogen"
)

// ServiceMeta defines metadata for any service (local or remote).
// This interface allows services to provide routing information
// for auto-router generation without tight coupling.
//
// Services should implement this interface to provide:
//   - Resource naming (singular and plural forms)
//   - Convention to use for route generation
//   - Custom route overrides
//
// This can be implemented by:
//   - Local services (for auto-router generation)
//   - Remote service wrappers (for proxy generation)
//
// Example (Local Service):
//
//	type OrderService struct {
//	    service.ServiceMetaAdapter
//	}
//
//	func NewOrderService() *OrderService {
//	    return &OrderService{
//	        ServiceMetaAdapter: service.ServiceMetaAdapter{
//	            Resource:   "order",
//	            Plural:     "orders",
//	            Convention: "rest",
//	            Override: autogen.RouteOverride{
//	                Custom: map[string]autogen.Route{
//	                    "Refund": {Method: "POST", Path: "/orders/{id}/refund"},
//	                },
//	            },
//	        },
//	    }
//	}
//
// Example (Remote Service):
//
//	type PaymentServiceRemote struct {
//	    service.ServiceMetaAdapter
//	}
//
//	func NewPaymentServiceRemote(proxyService *proxy.Service) *PaymentServiceRemote {
//	    return &PaymentServiceRemote{
//	        ServiceMetaAdapter: service.ServiceMetaAdapter{
//	            Resource:     "payment",
//	            Plural:       "payments",
//	            Convention:   "rest",
//	            ProxyService: proxyService,
//	        },
//	    }
//	}
type ServiceMeta interface {
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

// RemoteServiceMeta is an alias for ServiceMeta for backward compatibility.
// New code should use ServiceMeta directly.
type RemoteServiceMeta = ServiceMeta

// ServiceMetaAdapter provides default implementation of ServiceMeta.
// Can be used by both local and remote services.
//
// Example (Local Service with Custom Routes):
//
//	type OrderService struct {
//	    service.ServiceMetaAdapter
//	    // ... other fields
//	}
//
//	func NewOrderService() *OrderService {
//	    return &OrderService{
//	        ServiceMetaAdapter: service.ServiceMetaAdapter{
//	            Resource:   "order",
//	            Plural:     "orders",
//	            Convention: "rest",
//	            Override: autogen.RouteOverride{
//	                Custom: map[string]autogen.Route{
//	                    "Refund": {Method: "POST", Path: "/orders/{id}/refund"},
//	                },
//	            },
//	        },
//	    }
//	}
//
// Example (Remote Service):
//
//	type UserServiceRemote struct {
//	    service.ServiceMetaAdapter
//	}
//
//	func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
//	    return &UserServiceRemote{
//	        ServiceMetaAdapter: service.ServiceMetaAdapter{
//	            Resource:     "user",
//	            Plural:       "users",
//	            Convention:   "rest",
//	            ProxyService: proxyService,
//	        },
//	    }
//	}
type ServiceMetaAdapter struct {
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

// RemoteServiceMetaAdapter is an alias for ServiceMetaAdapter for backward compatibility.
// New code should use ServiceMetaAdapter directly.
type RemoteServiceMetaAdapter = ServiceMetaAdapter

// GetProxyService returns the proxy.Service instance
func (a *ServiceMetaAdapter) GetProxyService() *proxy.Service {
	return a.ProxyService
}

// GetResourceName implements ServiceMeta
func (a *ServiceMetaAdapter) GetResourceName() (string, string) {
	plural := a.Plural
	if plural == "" {
		plural = a.Resource + "s"
	}
	return a.Resource, plural
}

// GetConventionName implements ServiceMeta
func (a *ServiceMetaAdapter) GetConventionName() string {
	if a.Convention == "" {
		return "rest" // default convention
	}
	return a.Convention
}

// GetRouteOverride implements ServiceMeta
func (a *ServiceMetaAdapter) GetRouteOverride() autogen.RouteOverride {
	return a.Override
}

// Ensure ServiceMetaAdapter implements ServiceMeta
var _ ServiceMeta = (*ServiceMetaAdapter)(nil)
