package appservice

import (
	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/router/autogen"
)

// ========================================
// OrderServiceRemote - Convention-based Proxy
// ========================================
//
// This uses proxy.Service with convention-based method mapping.
// The framework automatically maps methods to HTTP endpoints:
//   - GetByID(params) → GET /orders/{id}
//   - GetByUserID(params) → POST /actions/get_by_user_id (custom action)
//
// Convention patterns:
//   1. Standard CRUD: List, GetByID, Create, Update, Delete
//   2. Custom actions: Any other method → POST /actions/{method_name_snake_case}
//

type OrderServiceRemote struct {
	service *proxy.Service
}

// Ensure OrderServiceRemote implements OrderService
var _ OrderService = (*OrderServiceRemote)(nil)

func NewOrderServiceRemote(baseURL string) *OrderServiceRemote {
	// Create proxy service with REST convention (empty string defaults to "rest")
	service := proxy.NewService(
		baseURL,
		autogen.ConversionRule{
			Convention:     "", // Empty = default REST convention
			Resource:       "order",
			ResourcePlural: "orders",
		},
		autogen.RouteOverride{}, // No code-level overrides, use config if needed
	)

	return &OrderServiceRemote{
		service: service,
	}
}

// GetByID is automatically handled by proxy.Service convention
// Maps to: GET /orders/{id}
func (s *OrderServiceRemote) GetByID(params *GetOrderParams) (*OrderWithUser, error) {
	return proxy.CallWithData[*OrderWithUser](s.service, "GetByID", params)
}

// GetByUserID is a custom action (non-standard CRUD)
// Maps to: POST /actions/get_by_user_id
func (s *OrderServiceRemote) GetByUserID(params *GetUserOrdersParams) ([]*Order, error) {
	return proxy.CallWithData[[]*Order](s.service, "GetByUserID", params)
}
