package application

import (
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/domain"
)

// UserServiceImpl implements domain.UserService
type UserServiceImpl struct {
	UserRepo *service.Cached[domain.UserRepository]
}

// Ensure implementation
var _ domain.UserService = (*UserServiceImpl)(nil)

// GetByID retrieves a user by ID
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
	return s.UserRepo.MustGet().GetByID(p.ID)
}

// List retrieves all users
func (s *UserServiceImpl) List(p *domain.ListUsersRequest) ([]*domain.User, error) {
	return s.UserRepo.MustGet().List()
}

// Create creates a new user
func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
	u := &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		RoleID: p.RoleID,
		Status: "active",
	}
	return s.UserRepo.MustGet().Create(u)
}

// Update updates an existing user
func (s *UserServiceImpl) Update(p *domain.UpdateUserRequest) (*domain.User, error) {
	u := &domain.User{
		ID:     p.ID,
		Name:   p.Name,
		Email:  p.Email,
		RoleID: p.RoleID,
	}
	return s.UserRepo.MustGet().Update(u)
}

// Suspend suspends a user account
func (s *UserServiceImpl) Suspend(p *domain.SuspendUserRequest) error {
	user, err := s.UserRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return err
	}
	user.Status = "suspended"
	_, err = s.UserRepo.MustGet().Update(user)
	return err
}

// Activate activates a user account
func (s *UserServiceImpl) Activate(p *domain.ActivateUserRequest) error {
	user, err := s.UserRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return err
	}
	user.Status = "active"
	_, err = s.UserRepo.MustGet().Update(user)
	return err
}

// Delete deletes a user
func (s *UserServiceImpl) Delete(p *domain.DeleteUserRequest) error {
	return s.UserRepo.MustGet().Delete(p.ID)
}

// UserServiceFactory creates a new UserServiceImpl instance
func UserServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserServiceImpl{
		UserRepo: service.Cast[domain.UserRepository](deps["user-repository"]),
	}
}
