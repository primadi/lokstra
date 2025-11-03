package service

import (
	"fmt"

	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/04-service-as-router/contract"
	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/04-service-as-router/model"
)

// UserService handles user-related business logic
type UserService struct {
	users []model.User
}

// NewUserService creates a new UserService instance
func NewUserService() *UserService {
	return &UserService{
		users: []model.User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
			{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
		},
	}
}

// List returns all users (optionally filtered by role)
func (s *UserService) List(p *contract.ListUsersParams) ([]model.User, error) {
	// In real app, would filter by p.Role
	return s.users, nil
}

// GetByID returns a user by ID
func (s *UserService) GetByID(p *contract.GetUserParams) (*model.User, error) {
	for _, user := range s.users {
		if user.ID == p.ID {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user with ID %d not found", p.ID)
}

// Additional methods (Create, Update, Delete) would be auto-mapped to:
// POST   /users       → Create()
// PUT    /users/{id}  → Update()
// DELETE /users/{id}  → Delete()
