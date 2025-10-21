package appservice

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
)

type OrderService interface {
	GetByID(p *GetOrderParams) (*OrderWithUser, error)
	GetByUserID(p *GetUserOrdersParams) ([]*Order, error)
}

type OrderServiceImpl struct {
	DB    *service.Cached[*Database]
	Users *service.Cached[UserService] // Cross-service dependency (can be local OR remote)
}

var _ OrderService = (*OrderServiceImpl)(nil) // Ensure implementation

type GetOrderParams struct {
	ID int `path:"id"`
}

type GetUserOrdersParams struct {
	UserID int `path:"user_id"`
}

type OrderWithUser struct {
	Order *Order `json:"order"`
	User  *User  `json:"user"`
}

func (s *OrderServiceImpl) GetByID(p *GetOrderParams) (*OrderWithUser, error) {
	// Get order
	order, err := s.DB.MustGet().GetOrder(p.ID)
	if err != nil {
		return nil, err
	}

	// Get associated user (cross-service call)
	// In monolith: Direct method call
	// In microservices: HTTP call to user-service
	user, err := s.Users.MustGet().GetByID(&GetUserParams{ID: order.UserID})
	if err != nil {
		return nil, fmt.Errorf("order found but user not found: %v", err)
	}

	return &OrderWithUser{
		Order: order,
		User:  user,
	}, nil
}

func (s *OrderServiceImpl) GetByUserID(p *GetUserOrdersParams) ([]*Order, error) {
	// Verify user exists (cross-service call)
	_, err := s.Users.MustGet().GetByID(&GetUserParams{ID: p.UserID})
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	return s.DB.MustGet().GetOrdersByUser(p.UserID)
}
