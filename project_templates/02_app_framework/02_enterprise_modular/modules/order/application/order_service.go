package application

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order/domain"
	userDomain "github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/domain"
)

// OrderServiceImpl implements domain.OrderService
type OrderServiceImpl struct {
	OrderRepo   *service.Cached[domain.OrderRepository]
	UserService *service.Cached[userDomain.UserService] // Cross-service dependency
}

// Ensure implementation
var _ domain.OrderService = (*OrderServiceImpl)(nil)

// GetByID retrieves an order by ID
func (s *OrderServiceImpl) GetByID(p *domain.GetOrderRequest) (*domain.Order, error) {
	return s.OrderRepo.MustGet().GetByID(p.ID)
}

// List retrieves all orders
func (s *OrderServiceImpl) List(p *domain.ListOrdersRequest) ([]*domain.Order, error) {
	if p.UserID > 0 {
		// Cross-service call: Validate user exists before listing their orders
		userService := s.UserService.MustGet()
		_, err := userService.GetByID(&userDomain.GetUserRequest{ID: p.UserID})
		if err != nil {
			return nil, fmt.Errorf("user validation failed: %w", err)
		}
		return s.OrderRepo.MustGet().GetByUserID(p.UserID)
	}
	return s.OrderRepo.MustGet().List()
}

// Create creates a new order
func (s *OrderServiceImpl) Create(p *domain.CreateOrderRequest) (*domain.Order, error) {
	// Cross-service call: Validate user exists before creating order
	userService := s.UserService.MustGet()
	u, err := userService.GetByID(&userDomain.GetUserRequest{ID: p.UserID})
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	o := &domain.Order{
		UserID:     p.UserID,
		UserName:   u.Name,
		TotalPrice: p.TotalPrice,
		Status:     "pending",
	}
	return s.OrderRepo.MustGet().Create(o)
}

// UpdateStatus updates an order's status
func (s *OrderServiceImpl) UpdateStatus(p *domain.UpdateOrderStatusRequest) (*domain.Order, error) {
	order, err := s.OrderRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return nil, err
	}
	order.Status = p.Status
	return s.OrderRepo.MustGet().Update(order)
}

// Cancel cancels an order
func (s *OrderServiceImpl) Cancel(p *domain.CancelOrderRequest) error {
	order, err := s.OrderRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return err
	}
	order.Status = "cancelled"
	_, err = s.OrderRepo.MustGet().Update(order)
	return err
}

// Delete deletes an order
func (s *OrderServiceImpl) Delete(p *domain.DeleteOrderRequest) error {
	return s.OrderRepo.MustGet().Delete(p.ID)
}

// OrderServiceFactory creates a new OrderServiceImpl instance
func OrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderServiceImpl{
		OrderRepo:   service.Cast[domain.OrderRepository](deps["order-repository"]),
		UserService: service.Cast[userDomain.UserService](deps["user-service"]), // Cross-service dependency
	}
}
