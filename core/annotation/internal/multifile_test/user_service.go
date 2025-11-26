package main

import (
	"github.com/primadi/lokstra/core/annotation/internal/multifile_test/userdomain"
	"github.com/primadi/lokstra/core/service"
)

// @RouterService name="user-service", prefix="/api/v1/users"
type UserService struct {
	// @Inject "user-repo"
	UserRepo *service.Cached[any]
}

// @Route "GET /users/{id}"
func (s *UserService) GetByID(id string) (*userdomain.UserDTO, error) {
	// Get user by ID from repository with validation
	return &userdomain.UserDTO{
		ID:   id,
		Name: "user",
	}, nil
}

// @Route "POST /users"
func (s *UserService) Create(req *userdomain.CreateUserRequest) (*userdomain.UserDTO, error) {
	return &userdomain.UserDTO{
		ID:    "new-id",
		Name:  req.Name,
		Email: req.Email,
	}, nil
}
