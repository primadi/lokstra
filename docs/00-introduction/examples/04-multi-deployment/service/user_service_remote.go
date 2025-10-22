package service

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/contract"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/model"
)

// ========================================
// User Service Remote (Proxy)
// ========================================

// UserServiceRemote implements contract.UserService with HTTP proxy
type UserServiceRemote struct {
	service.RemoteServiceMetaAdapter
}

// Ensure implementation
var _ contract.UserService = (*UserServiceRemote)(nil)

// NewUserServiceRemote creates a new remote user service proxy
func NewUserServiceRemote(proxyService *proxy.Service) *UserServiceRemote {
	return &UserServiceRemote{
		RemoteServiceMetaAdapter: service.RemoteServiceMetaAdapter{
			Resource:     "user",
			Plural:       "users",
			Convention:   "rest",
			ProxyService: proxyService,
		},
	}
}

// GetByID retrieves a user by ID via HTTP
func (s *UserServiceRemote) GetByID(p *contract.GetUserParams) (*model.User, error) {
	return proxy.CallWithData[*model.User](s.GetProxyService(), "GetByID", p)
}

// List retrieves all users via HTTP
func (s *UserServiceRemote) List(p *contract.ListUsersParams) ([]*model.User, error) {
	return proxy.CallWithData[[]*model.User](s.GetProxyService(), "List", p)
}

// ========================================
// Remote Factory
// ========================================

// UserServiceRemoteFactory creates a new UserServiceRemote instance
// Framework passes proxy.Service via config["remote"]
func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	return NewUserServiceRemote(
		service.CastProxyService(config["remote"]),
	)
}
