package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain/order"
)

// OrderServiceRemote implements order.OrderService with HTTP proxy
type OrderServiceRemote struct {
	proxyService *proxy.Service
}

// Ensure implementation
var _ order.OrderService = (*OrderServiceRemote)(nil)

// NewOrderServiceRemote creates a new remote order service proxy
func NewOrderServiceRemote(proxyService *proxy.Service) *OrderServiceRemote {
	return &OrderServiceRemote{
		proxyService: proxyService,
	}
}

// GetByID retrieves an order by ID via HTTP
func (s *OrderServiceRemote) GetByID(p *order.GetOrderParams) (*order.Order, error) {
	return proxy.CallWithData[*order.Order](s.proxyService, "GetByID", p)
}

// GetByUserID retrieves all orders for a specific user via HTTP
func (s *OrderServiceRemote) GetByUserID(p *order.GetOrdersByUserParams) ([]*order.Order, error) {
	return proxy.CallWithData[[]*order.Order](s.proxyService, "GetByUserID", p)
}

// List retrieves all orders via HTTP
func (s *OrderServiceRemote) List(p *order.ListOrdersParams) ([]*order.Order, error) {
	return proxy.CallWithData[[]*order.Order](s.proxyService, "List", p)
}

// Create creates a new order via HTTP
func (s *OrderServiceRemote) Create(p *order.CreateOrderParams) (*order.Order, error) {
	return proxy.CallWithData[*order.Order](s.proxyService, "Create", p)
}

// UpdateStatus updates the status of an order via HTTP
func (s *OrderServiceRemote) UpdateStatus(p *order.UpdateOrderStatusParams) (*order.Order, error) {
	return proxy.CallWithData[*order.Order](s.proxyService, "UpdateStatus", p)
}

// Delete deletes an order via HTTP
func (s *OrderServiceRemote) Delete(p *order.DeleteOrderParams) error {
	return proxy.Call(s.proxyService, "Delete", p)
}
