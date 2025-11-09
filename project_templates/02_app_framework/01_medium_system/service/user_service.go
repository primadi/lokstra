package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain/user"
)

// UserServiceImpl implements user.UserService
type UserServiceImpl struct {
	UserRepo *service.Cached[user.UserRepository]
}

// Ensure implementation
var _ user.UserService = (*UserServiceImpl)(nil)

// GetByID retrieves a user by ID
func (s *UserServiceImpl) GetByID(p *user.GetUserParams) (*user.User, error) {
	return s.UserRepo.MustGet().GetByID(p.ID)
}

// List retrieves all users
func (s *UserServiceImpl) List(p *user.ListUsersParams) ([]*user.User, error) {
	return s.UserRepo.MustGet().List()
}

// Create creates a new user
func (s *UserServiceImpl) Create(p *user.CreateUserParams) (*user.User, error) {
	u := &user.User{
		Name:  p.Name,
		Email: p.Email,
	}
	return s.UserRepo.MustGet().Create(u)
}

// Update updates an existing user
func (s *UserServiceImpl) Update(p *user.UpdateUserParams) (*user.User, error) {
	u := &user.User{
		ID:    p.ID,
		Name:  p.Name,
		Email: p.Email,
	}
	return s.UserRepo.MustGet().Update(u)
}

// Delete deletes a user
func (s *UserServiceImpl) Delete(p *user.DeleteUserParams) error {
	return s.UserRepo.MustGet().Delete(p.ID)
}

// UserServiceFactory creates a new UserServiceImpl instance
func UserServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserServiceImpl{
		UserRepo: service.Cast[user.UserRepository](deps["user-repository"]),
	}
}

// UserServiceRemoteFactory creates a remote client for UserService
// This is used when the service is deployed as a separate microservice
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	// The framework provides proxy.Service via config["remote"]
	proxyService, _ := config["remote"].(*proxy.Service)
	return NewUserServiceRemote(proxyService)
}
