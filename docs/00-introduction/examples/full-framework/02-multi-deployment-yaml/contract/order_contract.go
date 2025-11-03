package contract

import "github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment-yaml/model"

// ========================================
// Order Service Contract
// ========================================

// OrderService defines the interface for order-related operations
type OrderService interface {
	GetByID(p *GetOrderParams) (*OrderWithUser, error)
	GetByUserID(p *GetUserOrdersParams) ([]*model.Order, error)
}

// ========================================
// Request/Response DTOs
// ========================================

// GetOrderParams contains parameters for getting a single order
type GetOrderParams struct {
	ID int `path:"id"`
}

// GetUserOrdersParams contains parameters for getting orders by user
type GetUserOrdersParams struct {
	UserID int `path:"user_id"`
}

// OrderWithUser is a DTO that combines order with user information
type OrderWithUser struct {
	Order *model.Order `json:"order"`
	User  *model.User  `json:"user"`
}
