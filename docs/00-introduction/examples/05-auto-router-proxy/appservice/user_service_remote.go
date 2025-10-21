package appservice

import (
	"fmt"

	"github.com/primadi/lokstra/core/proxy"
)

// ========================================
// UserServiceRemote - Manual Proxy Approach
// ========================================
//
// NOTE: This implementation uses MANUAL proxy.Router with DoJSON() calls.
//
// In production, you would use automated patterns:
//   - proxy.NewService() with convention-based method mapping
//   - Same convention as router.NewFromService()
//   - Auto-mapping of methods to endpoints
//   - See EVOLUTION.md for automated patterns
//
// Manual approach shown here for educational purposes:
//   - Understand how HTTP proxy works
//   - Learn manual path construction
//   - See JSON wrapper handling
//   - Foundation before automation
//

type UserServiceRemote struct {
	proxy *proxy.Router
}

// GetByID implements UserService.
func (u *UserServiceRemote) GetByID(p *GetUserParams) (*User, error) {
	var JsonWrapper struct {
		Status string `json:"status"`
		Data   *User  `json:"data"`
	}

	err := u.proxy.DoJSON("GET", fmt.Sprintf("/users/%d", p.ID), nil, nil, &JsonWrapper)
	if err != nil {
		return nil, proxy.ParseRouterError(err)
	}
	return JsonWrapper.Data, nil
}

// List implements UserService.
func (u *UserServiceRemote) List(p *ListUsersParams) ([]*User, error) {
	var JsonWrapper struct {
		Status string  `json:"status"`
		Data   []*User `json:"data"`
	}
	err := u.proxy.DoJSON("GET", "/users", nil, nil, &JsonWrapper)
	if err != nil {
		return nil, proxy.ParseRouterError(err)
	}
	return JsonWrapper.Data, nil
}

var _ UserService = (*UserServiceRemote)(nil) // Ensure implementation

func NewUserServiceRemote() *UserServiceRemote {
	return &UserServiceRemote{
		proxy: proxy.NewRemoteRouter("http://localhost:3004"),
	}
}
