package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain/user"
)

// UserServiceRemote implements user.UserService with HTTP proxy
type UserServiceRemote struct {
	proxyService *proxy.Service
}

// Ensure implementation
var _ user.UserService = (*UserServiceRemote)(nil)

// NewUserServiceRemote creates a new remote user service proxy
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
	return &UserServiceRemote{
		proxyService: proxyService,
	}
}

// GetByID retrieves a user by ID via HTTP
func (s *UserServiceRemote) GetByID(p *user.GetUserParams) (*user.User, error) {
	return proxy.CallWithData[*user.User](s.proxyService, "GetByID", p)
}

// List retrieves all users via HTTP
func (s *UserServiceRemote) List(p *user.ListUsersParams) ([]*user.User, error) {
	return proxy.CallWithData[[]*user.User](s.proxyService, "List", p)
}

// Create creates a new user via HTTP
func (s *UserServiceRemote) Create(p *user.CreateUserParams) (*user.User, error) {
	return proxy.CallWithData[*user.User](s.proxyService, "Create", p)
}

// Update updates an existing user via HTTP
func (s *UserServiceRemote) Update(p *user.UpdateUserParams) (*user.User, error) {
	return proxy.CallWithData[*user.User](s.proxyService, "Update", p)
}

// Delete deletes a user via HTTP
func (s *UserServiceRemote) Delete(p *user.DeleteUserParams) error {
	return proxy.Call(s.proxyService, "Delete", p)
}
