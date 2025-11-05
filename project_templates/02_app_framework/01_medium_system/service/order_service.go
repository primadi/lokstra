package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain/order"
)

// OrderServiceImpl implements order.OrderService
type OrderServiceImpl struct {
	OrderRepo *service.Cached[order.OrderRepository]
}

// Ensure implementation
var _ order.OrderService = (*OrderServiceImpl)(nil)

// GetByID retrieves an order by ID
func (s *OrderServiceImpl) GetByID(p *order.GetOrderParams) (*order.Order, error) {
	return s.OrderRepo.MustGet().GetByID(p.ID)
}

// GetByUserID retrieves all orders for a specific user
func (s *OrderServiceImpl) GetByUserID(p *order.GetOrdersByUserParams) ([]*order.Order, error) {
	return s.OrderRepo.MustGet().GetByUserID(p.UserID)
}

// List retrieves all orders
func (s *OrderServiceImpl) List(p *order.ListOrdersParams) ([]*order.Order, error) {
	return s.OrderRepo.MustGet().List()
}

// Create creates a new order
func (s *OrderServiceImpl) Create(p *order.CreateOrderParams) (*order.Order, error) {
	o := &order.Order{
		UserID:   p.UserID,
		Product:  p.Product,
		Quantity: p.Quantity,
		Total:    p.Total,
		Status:   "pending",
	}
	return s.OrderRepo.MustGet().Create(o)
}

// UpdateStatus updates the status of an order
func (s *OrderServiceImpl) UpdateStatus(p *order.UpdateOrderStatusParams) (*order.Order, error) {
	// Get existing order
	o, err := s.OrderRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return nil, err
	}

	// Update status
	o.Status = p.Status
	return s.OrderRepo.MustGet().Update(o)
}

// Delete deletes an order
func (s *OrderServiceImpl) Delete(p *order.DeleteOrderParams) error {
	return s.OrderRepo.MustGet().Delete(p.ID)
}

// OrderServiceFactory creates a new OrderServiceImpl instance
func OrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderServiceImpl{
		OrderRepo: service.Cast[order.OrderRepository](deps["order-repository"]),
	}
}

// OrderServiceRemoteFactory creates a remote client for OrderService
// This is used when the service is deployed as a separate microservice
func OrderServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	// The framework provides proxy.Service via config["remote"]
	proxyService, _ := config["remote"].(*proxy.Service)
	return NewOrderServiceRemote(proxyService)
}
