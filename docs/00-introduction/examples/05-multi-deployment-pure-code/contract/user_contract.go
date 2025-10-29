package contract

import "github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/model"

// ========================================
// User Service Contract
// ========================================

// UserService defines the interface for user-related operations
type UserService interface {
	GetByID(p *GetUserParams) (*model.User, error)
	List(p *ListUsersParams) ([]*model.User, error)
}

// ========================================
// Request/Response DTOs
// ========================================

// GetUserParams contains parameters for getting a single user
type GetUserParams struct {
	ID int `path:"id"`
}

// ListUsersParams contains parameters for listing users
type ListUsersParams struct{}
