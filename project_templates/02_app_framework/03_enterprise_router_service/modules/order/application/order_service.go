package application

import (
	"fmt"

	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/order/domain"
	userDomain "github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user/domain"
)

// @RouterService name="order-service", prefix="/api", middlewares=["recovery", "request-logger"]
type OrderServiceImpl struct {
	// @Inject "@store.order-repository"
	OrderRepo domain.OrderRepository
	// @Inject "user-service"
	UserService userDomain.UserService
}

// Ensure implementation
var _ domain.OrderService = (*OrderServiceImpl)(nil)

// @Route "GET /orders/{id}"
func (s *OrderServiceImpl) GetByID(p *domain.GetOrderRequest) (*domain.Order, error) {
	return s.OrderRepo.GetByID(p.ID)
}

// @Route "GET /orders"
func (s *OrderServiceImpl) List(p *domain.ListOrdersRequest) ([]*domain.Order, error) {
	if p.UserID > 0 {
		// Cross-service call: Validate user exists before listing their orders
		_, err := s.UserService.GetByID(&userDomain.GetUserRequest{ID: p.UserID})
		if err != nil {
			return nil, fmt.Errorf("user validation failed: %w", err)
		}
		return s.OrderRepo.GetByUserID(p.UserID)
	}
	return s.OrderRepo.List()
}

// @Route "POST /orders"
func (s *OrderServiceImpl) Create(p *domain.CreateOrderRequest) (*domain.Order, error) {
	// Cross-service call: Validate user exists before creating order
	u, err := s.UserService.GetByID(&userDomain.GetUserRequest{ID: p.UserID})
	if err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	o := &domain.Order{
		UserID:     p.UserID,
		UserName:   u.Name,
		TotalPrice: p.TotalPrice,
		Status:     "pending",
	}
	return s.OrderRepo.Create(o)
}

// @Route "PUT /orders/{id}/status"
func (s *OrderServiceImpl) UpdateStatus(p *domain.UpdateOrderStatusRequest) (*domain.Order, error) {
	order, err := s.OrderRepo.GetByID(p.ID)
	if err != nil {
		return nil, err
	}
	order.Status = p.Status
	return s.OrderRepo.Update(order)
}

// @Route "POST /orders/{id}/cancel"
func (s *OrderServiceImpl) Cancel(p *domain.CancelOrderRequest) error {
	order, err := s.OrderRepo.GetByID(p.ID)
	if err != nil {
		return err
	}
	order.Status = "cancelled"
	_, err = s.OrderRepo.Update(order)
	return err
}

// @Route "DELETE /orders/{id}"
func (s *OrderServiceImpl) Delete(p *domain.DeleteOrderRequest) error {
	return s.OrderRepo.Delete(p.ID)
}

func Register() {
	// do nothing, just to ensure package is loaded
}
