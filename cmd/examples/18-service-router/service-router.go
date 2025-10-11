package main

import (
	"fmt"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

// ============================================================================
// User Models
// ============================================================================

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// ============================================================================
// User Service - Basic CRUD Operations
// ============================================================================

type UserService struct {
	users map[string]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: map[string]*User{
			"1": {ID: "1", Name: "John Doe", Email: "john@example.com"},
			"2": {ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
		},
	}
}

// GET /users - List all users
func (s *UserService) ListUsers(ctx *request.Context) ([]*User, error) {
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

// GET /users/{id} - Get user by ID
func (s *UserService) GetUser(ctx *request.Context, id string) (*User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// POST /users - Create new user
func (s *UserService) CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error) {
	id := fmt.Sprintf("%d", len(s.users)+1)
	user := &User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[id] = user
	return user, nil
}

// PUT /users/{id} - Update user
func (s *UserService) UpdateUser(ctx *request.Context, id string, req *UpdateUserRequest) (*User, error) {
	user, exists := s.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	user.Name = req.Name
	user.Email = req.Email
	return user, nil
}

// DELETE /users/{id} - Delete user
func (s *UserService) DeleteUser(ctx *request.Context, id string) error {
	if _, exists := s.users[id]; !exists {
		return fmt.Errorf("user not found")
	}
	delete(s.users, id)
	return nil
}

// GET /users/search?q=xxx - Search users
func (s *UserService) SearchUsers(ctx *request.Context) ([]*User, error) {
	query := ctx.Req.QueryParam("q", "")
	results := make([]*User, 0)
	for _, user := range s.users {
		if query == "" || containsIgnoreCase(user.Name, query) || containsIgnoreCase(user.Email, query) {
			results = append(results, user)
		}
	}
	return results, nil
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}
