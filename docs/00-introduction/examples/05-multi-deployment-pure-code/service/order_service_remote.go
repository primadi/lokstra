package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/contract"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/model"
)

// ========================================
// Order Service Remote (Proxy)
// ========================================

// OrderServiceRemote implements contract.OrderService with HTTP proxy
type OrderServiceRemote struct {
	proxyService *proxy.Service
}

// Ensure implementation
var _ contract.OrderService = (*OrderServiceRemote)(nil)

// NewOrderServiceRemote creates a new remote order service proxy
func NewOrderServiceRemote(proxyService *proxy.Service) *OrderServiceRemote {
	return &OrderServiceRemote{
		proxyService: proxyService,
	}
}

// GetByID retrieves an order with user information via HTTP
func (s *OrderServiceRemote) GetByID(p *contract.GetOrderParams) (*contract.OrderWithUser, error) {
	return proxy.CallWithData[*contract.OrderWithUser](s.proxyService, "GetByID", p)
}

// GetByUserID retrieves all orders for a user via HTTP
func (s *OrderServiceRemote) GetByUserID(p *contract.GetUserOrdersParams) ([]*model.Order, error) {
	return proxy.CallWithData[[]*model.Order](s.proxyService, "GetByUserID", p)
}

// ========================================
// Remote Factory
// ========================================

// OrderServiceRemoteFactory creates a new OrderServiceRemote instance
// Framework passes proxy.Service via config["remote"]
func OrderServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	proxyService, _ := config["remote"].(*proxy.Service)
	return NewOrderServiceRemote(proxyService)
}
