package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/contract"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/model"
)

// ========================================
// Order Service Remote (Proxy)
// ========================================

// OrderServiceRemote implements contract.OrderService with HTTP proxy
type OrderServiceRemote struct {
	service.RemoteServiceMetaAdapter
}

// Ensure implementation
var _ contract.OrderService = (*OrderServiceRemote)(nil)

// NewOrderServiceRemote creates a new remote order service proxy
func NewOrderServiceRemote(proxyService *proxy.Service) *OrderServiceRemote {
	return &OrderServiceRemote{
		RemoteServiceMetaAdapter: service.RemoteServiceMetaAdapter{
			Resource:     "order",
			Plural:       "orders",
			Convention:   "rest",
			ProxyService: proxyService,
			// Custom route override for nested resource
			Override: autogen.RouteOverride{
				Custom: map[string]autogen.Route{
					"GetByUserID": {Method: "GET", Path: "/users/{user_id}/orders"},
				},
			},
		},
	}
}

// GetByID retrieves an order with user information via HTTP
func (s *OrderServiceRemote) GetByID(p *contract.GetOrderParams) (*contract.OrderWithUser, error) {
	return proxy.CallWithData[*contract.OrderWithUser](s.GetProxyService(), "GetByID", p)
}

// GetByUserID retrieves all orders for a user via HTTP
func (s *OrderServiceRemote) GetByUserID(p *contract.GetUserOrdersParams) ([]*model.Order, error) {
	return proxy.CallWithData[[]*model.Order](s.GetProxyService(), "GetByUserID", p)
}

// ========================================
// Remote Factory
// ========================================

// OrderServiceRemoteFactory creates a new OrderServiceRemote instance
// Framework passes proxy.Service via config["remote"]
func OrderServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	return NewOrderServiceRemote(
		service.CastProxyService(config["remote"]),
	)
}
