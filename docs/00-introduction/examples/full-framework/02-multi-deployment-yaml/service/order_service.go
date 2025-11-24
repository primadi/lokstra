package service

import (
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/contract"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/model"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/repository"
)

// ========================================
// Order Service Implementation (Local)
// ========================================

// OrderServiceImpl implements contract.OrderService with local repository
type OrderServiceImpl struct {
	OrderRepo repository.OrderRepository
	UserSvc   contract.UserService // Can be local OR remote!
}

// Ensure implementation
var _ contract.OrderService = (*OrderServiceImpl)(nil)

// GetByID retrieves an order by ID with associated user information
func (s *OrderServiceImpl) GetByID(p *contract.GetOrderParams) (*contract.OrderWithUser, error) {
	// Get order from repository
	order, err := s.OrderRepo.GetByID(p.ID)
	if err != nil {
		return nil, err
	}

	// Get associated user (cross-service call - can be local or remote!)
	user, err := s.UserSvc.GetByID(&contract.GetUserParams{ID: order.UserID})
	if err != nil {
		return nil, err
	}

	return &contract.OrderWithUser{
		Order: order,
		User:  user,
	}, nil
}

// GetByUserID retrieves all orders for a specific user
func (s *OrderServiceImpl) GetByUserID(p *contract.GetUserOrdersParams) ([]*model.Order, error) {
	// Verify user exists first (cross-service call - can be local or remote!)
	_, err := s.UserSvc.GetByID(&contract.GetUserParams{ID: p.UserID})
	if err != nil {
		return nil, err
	}

	// Get orders from repository
	return s.OrderRepo.GetByUserID(p.UserID)
}

// ========================================
// Factory
// ========================================

// OrderServiceFactory creates a new OrderServiceImpl instance
func OrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderServiceImpl{
		OrderRepo: deps["order-repository"].(repository.OrderRepository),
		UserSvc:   deps["user-service"].(contract.UserService),
	}
}
