package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/contract"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/model"
)

// ========================================
// User Service Remote (Proxy)
// ========================================

// UserServiceRemote implements contract.UserService with HTTP proxy
type UserServiceRemote struct {
	proxyService *proxy.Service
}

// Ensure implementation
var _ contract.UserService = (*UserServiceRemote)(nil)

// NewUserServiceRemote creates a new remote user service proxy
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
	return &UserServiceRemote{
		proxyService: proxyService,
	}
}

// GetByID retrieves a user by ID via HTTP
func (s *UserServiceRemote) GetByID(p *contract.GetUserParams) (*model.User, error) {
	return proxy.CallWithData[*model.User](s.proxyService, "GetByID", p)
}

// List retrieves all users via HTTP
func (s *UserServiceRemote) List(p *contract.ListUsersParams) ([]*model.User, error) {
	return proxy.CallWithData[[]*model.User](s.proxyService, "List", p)
}

// ========================================
// Remote Factory
// ========================================

// UserServiceRemoteFactory creates a new UserServiceRemote instance
// Framework passes proxy.Service via config["remote"]
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	proxyService, _ := config["remote"].(*proxy.Service)
	return NewUserServiceRemote(proxyService)
}
