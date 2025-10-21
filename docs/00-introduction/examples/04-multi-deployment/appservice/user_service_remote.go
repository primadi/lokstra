package appservice

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/router/autogen"
)

// ========================================
// UserServiceRemote - Convention-based Proxy
// ========================================
//
// This uses proxy.Service with convention-based method mapping.
// The framework automatically maps methods to HTTP endpoints:
//   - List() → GET /users
//   - GetByID(params) → GET /users/{id}
//
// This is the RECOMMENDED approach instead of manual proxy.Router:
//   ✅ Automatic method-to-HTTP mapping via conventions
//   ✅ Type-safe response handling with generics
//   ✅ Automatic JSON wrapper extraction
//   ✅ Less code, fewer errors
//

type UserServiceRemote struct {
	service *proxy.Service
}

// Ensure UserServiceRemote implements UserService
var _ UserService = (*UserServiceRemote)(nil)

func NewUserServiceRemote(baseURL string) *UserServiceRemote {
	// Create proxy service with REST convention (empty string defaults to "rest")
	service := proxy.NewService(
		baseURL,
		autogen.ConversionRule{
			Convention:     "", // Empty = default REST convention
			Resource:       "user",
			ResourcePlural: "users",
		},
		autogen.RouteOverride{}, // No code-level overrides, use config if needed
	)

	return &UserServiceRemote{
		service: service,
	}
}

// GetByID maps to: GET /users/{id}
func (u *UserServiceRemote) GetByID(params *GetUserParams) (*User, error) {
	return proxy.CallWithData[*User](u.service, "GetByID", params)
}

// List maps to: GET /users
func (u *UserServiceRemote) List(params *ListUsersParams) ([]*User, error) {
	return proxy.CallWithData[[]*User](u.service, "List", params)
}
