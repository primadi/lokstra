package application

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/order/domain"
	userDomain "github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user/domain"
)

// @RouterService name="order-service", prefix="/api", middlewares=["recovery", "request-logger"]
type OrderServiceImpl struct {
	// @Inject "order-repository"
	OrderRepo *service.Cached[domain.OrderRepository]
	// @Inject "user-service"
	UserService *service.Cached[userDomain.UserService]
}

// Ensure implementation
var _ domain.OrderService = (*OrderServiceImpl)(nil)

// @Route "GET /orders/{id}"
func (s *OrderServiceImpl) GetByID(p *domain.GetOrderRequest) (*domain.Order, error) {
	return s.OrderRepo.MustGet().GetByID(p.ID)
}

// @Route "GET /orders"
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

// @Route "POST /orders"
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

// @Route "PUT /orders/{id}/status"
func (s *OrderServiceImpl) UpdateStatus(p *domain.UpdateOrderStatusRequest) (*domain.Order, error) {
	order, err := s.OrderRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return nil, err
	}
	order.Status = p.Status
	return s.OrderRepo.MustGet().Update(order)
}

// @Route "POST /orders/{id}/cancel"
func (s *OrderServiceImpl) Cancel(p *domain.CancelOrderRequest) error {
	order, err := s.OrderRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return err
	}
	order.Status = "cancelled"
	_, err = s.OrderRepo.MustGet().Update(order)
	return err
}

// @Route "DELETE /orders/{id}"
func (s *OrderServiceImpl) Delete(p *domain.DeleteOrderRequest) error {
	return s.OrderRepo.MustGet().Delete(p.ID)
}

func Register() {
	// do nothing, just to ensure package is loaded
}
