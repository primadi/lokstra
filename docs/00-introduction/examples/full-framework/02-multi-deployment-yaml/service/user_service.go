package service

import (
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/contract"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/model"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/repository"
)

// ========================================
// User Service Implementation (Local)
// ========================================

// UserServiceImpl implements contract.UserService with local repository
type UserServiceImpl struct {
	UserRepo *service.Cached[repository.UserRepository]
}

// Ensure implementation
var _ contract.UserService = (*UserServiceImpl)(nil)

// GetByID retrieves a user by ID
func (s *UserServiceImpl) GetByID(p *contract.GetUserParams) (*model.User, error) {
	return s.UserRepo.MustGet().GetByID(p.ID)
}

// List retrieves all users
func (s *UserServiceImpl) List(p *contract.ListUsersParams) ([]*model.User, error) {
	return s.UserRepo.MustGet().List()
}

// ========================================
// Factory
// ========================================

// UserServiceFactory creates a new UserServiceImpl instance
func UserServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserServiceImpl{
		UserRepo: service.Cast[repository.UserRepository](deps["user-repository"]),
	}
}
