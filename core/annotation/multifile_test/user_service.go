package main

import "github.com/primadi/lokstra/core/service"

// @RouterService name="user-service", prefix="/api/v1/users"
type UserService struct {
	// @Inject "user-repo"
	UserRepo *service.Cached[any]
}

// @Route "GET /users/{id}"
func (s *UserService) GetByID(id string) (string, error) {
	return "user", nil
}

// @Route "POST /users"
func (s *UserService) Create(name string, email string) (string, error) {
	return "created", nil
}