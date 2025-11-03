package service

import (
	"fmt"

	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/03-service-dependencies/model"
	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/03-service-dependencies/repository"
	"github.com/primadi/lokstra/lokstra_registry"
)

type UserService struct {
	repo *repository.UserRepository
}

// NewUserService - Mode 1: No params (gets dependency from registry)
func NewUserService() *UserService {
	// Get dependency from registry
	repo := lokstra_registry.MustGetService[*repository.UserRepository]("user-repo")

	fmt.Println("✅ UserService created (Mode 1: no params, dependency from registry)")

	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetUser(id int) (*model.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) GetAllUsers() ([]model.User, error) {
	return s.repo.FindAll()
}
