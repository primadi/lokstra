package application

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/domain"
)

// UserServiceRemote implements domain.UserService with HTTP proxy
type UserServiceRemote struct {
	proxyService *proxy.Service
}

// Ensure implementation
var _ domain.UserService = (*UserServiceRemote)(nil)

// NewUserServiceRemote creates a new remote user service proxy
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
	return &UserServiceRemote{
		proxyService: proxyService,
	}
}

// GetByID retrieves a user by ID via HTTP
func (s *UserServiceRemote) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
	return proxy.CallWithData[*domain.User](s.proxyService, "GetByID", p)
}

// List retrieves all users via HTTP
func (s *UserServiceRemote) List(p *domain.ListUsersRequest) ([]*domain.User, error) {
	return proxy.CallWithData[[]*domain.User](s.proxyService, "List", p)
}

// Create creates a new user via HTTP
func (s *UserServiceRemote) Create(p *domain.CreateUserRequest) (*domain.User, error) {
	return proxy.CallWithData[*domain.User](s.proxyService, "Create", p)
}

// Update updates an existing user via HTTP
func (s *UserServiceRemote) Update(p *domain.UpdateUserRequest) (*domain.User, error) {
	return proxy.CallWithData[*domain.User](s.proxyService, "Update", p)
}

// Suspend suspends a user account via HTTP
func (s *UserServiceRemote) Suspend(p *domain.SuspendUserRequest) error {
	return proxy.Call(s.proxyService, "Suspend", p)
}

// Activate activates a user account via HTTP
func (s *UserServiceRemote) Activate(p *domain.ActivateUserRequest) error {
	return proxy.Call(s.proxyService, "Activate", p)
}

// Delete deletes a user via HTTP
func (s *UserServiceRemote) Delete(p *domain.DeleteUserRequest) error {
	return proxy.Call(s.proxyService, "Delete", p)
}
