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
// Cart Service Interface
// ==============================================================================

type CartService interface {
	AddToCart(ctx *request.Context, req *AddToCartRequest) (*AddToCartResponse, error)
	GetCart(ctx *request.Context, req *GetCartRequest) (*GetCartResponse, error)
}

// ==============================================================================
// Request/Response Structs
// ==============================================================================

type AddToCartRequest struct {
	UserID   string  `json:"user_id" validate:"required"`
	ItemID   string  `json:"item_id" validate:"required"`
	Quantity int     `json:"quantity" validate:"required,min=1"`
	Price    float64 `json:"price" validate:"required,min=0"`
}

type ItemCart struct {
	ItemID   string  `json:"item_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type AddToCartResponse struct {
	CartID string      `json:"cart_id"`
	UserID string      `json:"user_id"`
	Items  []*ItemCart `json:"items"`
}

type GetCartRequest struct {
	UserID string `path:"userId" json:"user_id" validate:"required"`
}

type GetCartResponse struct {
	CartID string      `json:"cart_id"`
	UserID string      `json:"user_id"`
	Items  []*ItemCart `json:"items"`
}

// ==============================================================================
// Local Implementation
// ==============================================================================

type cartServiceLocal struct {
	storage        string
	sessionTimeout int
	userService    *service.Lazy[UserService] // Cross-server dependency
	carts          map[string]*GetCartResponse
}

func (s *cartServiceLocal) AddToCart(ctx *request.Context, req *AddToCartRequest) (*AddToCartResponse, error) {
	// Verify user exists (cross-service call - may be LOCAL or REMOTE!)
	if s.userService != nil {
		fmt.Printf("[cart-service] Getting user service (lazy load)...\n")
		userSvc := s.userService.Get()
		fmt.Printf("[cart-service] Got user service: %T\n", userSvc)
		fmt.Printf("[cart-service] Calling GetUser for UserID: %s\n", req.UserID)
		user, err := userSvc.GetUser(ctx, &GetUserRequest{UserID: req.UserID})
		if err != nil {
			fmt.Printf("[cart-service] GetUser failed: %v\n", err)
			return nil, fmt.Errorf("user verification failed: %w", err)
		}
		fmt.Printf("[cart-service] User verified successfully: %v\n", user)
	}

	cartID := fmt.Sprintf("cart_%s", req.UserID)
	cart, ok := s.carts[cartID]
	if !ok {
		cart = &GetCartResponse{
			CartID: cartID,
			UserID: req.UserID,
			Items:  []*ItemCart{},
		}
		s.carts[cartID] = cart
	}

	cart.Items = append(cart.Items, &ItemCart{
		ItemID:   req.ItemID,
		Quantity: req.Quantity,
		Price:    req.Price,
	})

	return &AddToCartResponse{
		CartID: cartID,
		UserID: req.UserID,
		Items:  cart.Items,
	}, nil
}

func (s *cartServiceLocal) GetCart(ctx *request.Context, req *GetCartRequest) (*GetCartResponse, error) {
	cartID := fmt.Sprintf("cart_%s", req.UserID)
	cart, ok := s.carts[cartID]
	if !ok {
		return &GetCartResponse{
			CartID: cartID,
			UserID: req.UserID,
			Items:  []*ItemCart{},
		}, nil
	}
	return cart, nil
}

// ==============================================================================
// Remote Implementation
// ==============================================================================

type cartServiceRemote struct {
	client *api_client.RemoteService
}

func (s *cartServiceRemote) AddToCart(ctx *request.Context, req *AddToCartRequest) (*AddToCartResponse, error) {
	return api_client.CallRemoteService[*AddToCartResponse](s.client, "AddToCart", ctx, req)
}

func (s *cartServiceRemote) GetCart(ctx *request.Context, req *GetCartRequest) (*GetCartResponse, error) {
	return api_client.CallRemoteService[*GetCartResponse](s.client, "GetCart", ctx, req)
}

// ==============================================================================
// Service Factories
// ==============================================================================

// CreateCartServiceLocal creates local implementation
func CreateCartServiceLocal(cfg map[string]any) any {
	storage := utils.GetValueFromMap(cfg, "storage", "memory")
	sessionTimeout := utils.GetValueFromMap(cfg, "session_timeout", 1800)

	fmt.Printf("[cart-service] Creating LOCAL with storage: %s, session_timeout: %d\n", storage, sessionTimeout)

	// Get lazy dependency using Dep[T] with underscore key (matches YAML)
	return &cartServiceLocal{
		storage:        storage,
		sessionTimeout: sessionTimeout,
		userService:    service.MustLazyLoadFromConfig[UserService](cfg, "user_service"),
		carts:          make(map[string]*GetCartResponse),
	}
}

// CreateCartServiceRemote creates HTTP client wrapper
func CreateCartServiceRemote(cfg map[string]any) any {
	routerName := utils.GetValueFromMap(cfg, "router", "cart-service")
	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/cart")

	fmt.Printf("[cart-service] Creating REMOTE client for router: %s, prefix: %s\n", routerName, pathPrefix)

	return &cartServiceRemote{
		client: lokstra_registry.GetRemoteService(cfg),
	}
}

func RegisterCartService() {
	lokstra_registry.RegisterServiceFactoryLocalAndRemote("cart_service",
		CreateCartServiceLocal, CreateCartServiceRemote)
}
