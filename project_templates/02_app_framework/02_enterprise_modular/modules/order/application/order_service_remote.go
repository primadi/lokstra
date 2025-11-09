package application

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order/domain"
)

// OrderServiceRemote implements domain.OrderService with HTTP proxy
type OrderServiceRemote struct {
	proxyService *proxy.Service
}

// Ensure implementation
var _ domain.OrderService = (*OrderServiceRemote)(nil)

// NewOrderServiceRemote creates a new remote order service proxy
func NewOrderServiceRemote(proxyService *proxy.Service) *OrderServiceRemote {
	return &OrderServiceRemote{
		proxyService: proxyService,
	}
}

// GetByID retrieves an order by ID via HTTP
func (s *OrderServiceRemote) GetByID(p *domain.GetOrderRequest) (*domain.Order, error) {
	return proxy.CallWithData[*domain.Order](s.proxyService, "GetByID", p)
}

// List retrieves all orders via HTTP
func (s *OrderServiceRemote) List(p *domain.ListOrdersRequest) ([]*domain.Order, error) {
	return proxy.CallWithData[[]*domain.Order](s.proxyService, "List", p)
}

// Create creates a new order via HTTP
func (s *OrderServiceRemote) Create(p *domain.CreateOrderRequest) (*domain.Order, error) {
	return proxy.CallWithData[*domain.Order](s.proxyService, "Create", p)
}

// UpdateStatus updates the status of an order via HTTP
func (s *OrderServiceRemote) UpdateStatus(p *domain.UpdateOrderStatusRequest) (*domain.Order, error) {
	return proxy.CallWithData[*domain.Order](s.proxyService, "UpdateStatus", p)
}

// Cancel cancels an order via HTTP
func (s *OrderServiceRemote) Cancel(p *domain.CancelOrderRequest) error {
	return proxy.Call(s.proxyService, "Cancel", p)
}

// Delete deletes an order via HTTP
func (s *OrderServiceRemote) Delete(p *domain.DeleteOrderRequest) error {
	return proxy.Call(s.proxyService, "Delete", p)
}
