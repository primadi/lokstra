package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/service"
)

// ========================================
// Service with Lazy DI
// ========================================

type UserService struct {
	DB *service.Cached[*Database]
}

// Request types for service methods
type GetByIDParams struct {
	ID int `path:"id"`
}

type CreateParams struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateParams struct {
	ID    int    `path:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type DeleteParams struct {
	ID int `path:"id"`
}

// Service methods
func (s *UserService) GetAll() ([]*User, error) {
	return s.DB.MustGet().GetAll()
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
	user, err := s.DB.MustGet().GetByID(p.ID)
	if err != nil {
		return nil, fmt.Errorf("user with ID %d not found", p.ID)
	}
	return user, nil
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
	user, err := s.DB.MustGet().Create(p.Name, p.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Update(p *UpdateParams) (*User, error) {
	user, err := s.DB.MustGet().Update(p.ID, p.Name, p.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Delete(p *DeleteParams) error {
	err := s.DB.MustGet().Delete(p.ID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// ========================================
// CHOOSE YOUR APPROACH!
// ========================================
// Run with: go run main.go --mode=code     (default - manual approach)
// Run with: go run main.go --mode=config   (new paradigm with YAML)

func UserServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserService{
		DB: service.Cast[*Database](deps["database"]),
	}
}
