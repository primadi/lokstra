package services

import (
	"fmt"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ==============================================================================
// Order Service Interface
// ==============================================================================

type OrderService interface {
	CreateOrder(ctx *request.Context, req *CreateOrderRequest) (*CreateOrderResponse, error)
	GetOrder(ctx *request.Context, req *GetOrderRequest) (*GetOrderResponse, error)
}

// ==============================================================================
// Request/Response Structs
// ==============================================================================

type ItemOrder struct {
	ProductID string  `json:"product_id" validate:"required"`
	Quantity  int     `json:"quantity" validate:"required,min=1"`
	Price     float64 `json:"price" validate:"required,min=0"`
}

type CreateOrderRequest struct {
	UserID string       `json:"user_id" validate:"required"`
	Items  []*ItemOrder `json:"items" validate:"required"`
}

type CreateOrderResponse struct {
	OrderID string       `json:"order_id"`
	UserID  string       `json:"user_id"`
	Items   []*ItemOrder `json:"items"`
	Status  string       `json:"status"`
}

type GetOrderRequest struct {
	OrderID string `path:"id" json:"order_id" validate:"required"`
}

type GetOrderResponse struct {
	OrderID string       `json:"order_id"`
	UserID  string       `json:"user_id"`
	Items   []*ItemOrder `json:"items"`
	Status  string       `json:"status"`
}

// ==============================================================================
// Local Implementation
// ==============================================================================

type orderServiceLocal struct {
	storage        string
	maxItems       int
	userService    *service.Lazy[UserService]    // Cross-server dependency
	paymentService *service.Lazy[PaymentService] // Cross-server dependency
	orders         map[string]*GetOrderResponse
}

func (s *orderServiceLocal) CreateOrder(ctx *request.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	if len(req.Items) > s.maxItems {
		return nil, fmt.Errorf("too many items (max: %d)", s.maxItems)
	}

	// Verify user exists (cross-service call - may be LOCAL or REMOTE!)
	if s.userService != nil {
		userSvc := s.userService.Get() // Dereference pointer to interface
		_, err := userSvc.GetUser(ctx, &GetUserRequest{UserID: req.UserID})
		if err != nil {
			return nil, fmt.Errorf("user verification failed: %w", err)
		}
	}

	orderID := fmt.Sprintf("order_%d", len(s.orders)+1)
	order := &GetOrderResponse{
		OrderID: orderID,
		UserID:  req.UserID,
		Items:   req.Items,
		Status:  "pending",
	}
	s.orders[orderID] = order

	return &CreateOrderResponse{
		OrderID: orderID,
		UserID:  req.UserID,
		Items:   req.Items,
		Status:  "pending",
	}, nil
}

func (s *orderServiceLocal) GetOrder(ctx *request.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	order, ok := s.orders[req.OrderID]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

// ==============================================================================
// Remote Implementation
// ==============================================================================

type orderServiceRemote struct {
	client *api_client.RemoteService
}

func (s *orderServiceRemote) CreateOrder(ctx *request.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	return api_client.CallRemoteService[*CreateOrderResponse](s.client, "CreateOrder", ctx, req)
}

func (s *orderServiceRemote) GetOrder(ctx *request.Context, req *GetOrderRequest) (*GetOrderResponse, error) {
	return api_client.CallRemoteService[*GetOrderResponse](s.client, "GetOrder", ctx, req)
}

// ==============================================================================
// Service Factories
// ==============================================================================

// CreateOrderServiceLocal creates local implementation
func CreateOrderServiceLocal(cfg map[string]any) any {
	storage := utils.GetValueFromMap(cfg, "storage", "memory")
	maxItems := utils.GetValueFromMap(cfg, "max_items_per_order", 50)

	fmt.Printf("[order-service] Creating LOCAL with storage: %s, max_items: %d\n", storage, maxItems)

	// Get lazy dependencies using Dep[T] (shorter syntax)
	// Config keys use underscore (user_service) to match YAML convention
	return &orderServiceLocal{
		storage:        storage,
		maxItems:       maxItems,
		userService:    service.MustLazyLoadFromConfig[UserService](cfg, "user_service"),
		paymentService: service.MustLazyLoadFromConfig[PaymentService](cfg, "payment_service"),
		orders:         make(map[string]*GetOrderResponse),
	}
}

// CreateOrderServiceRemote creates HTTP client wrapper
func CreateOrderServiceRemote(cfg map[string]any) any {
	routerName := utils.GetValueFromMap(cfg, "router", "order-service")
	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/orders")

	fmt.Printf("[order-service] Creating REMOTE client for router: %s, prefix: %s\n", routerName, pathPrefix)

	return &orderServiceRemote{
		client: lokstra_registry.GetRemoteService(cfg),
	}
}

func RegisterOrderService() {
	lokstra_registry.RegisterServiceFactoryLocalAndRemote("order_service",
		CreateOrderServiceLocal, CreateOrderServiceRemote)
}
