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
// Payment Service Interface
// ==============================================================================

type PaymentService interface {
	ProcessPayment(ctx *request.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error)
	GetPayment(ctx *request.Context, req *GetPaymentRequest) (*GetPaymentResponse, error)
}

// ==============================================================================
// Request/Response Structs
// ==============================================================================

type ProcessPaymentRequest struct {
	OrderID string  `json:"order_id" validate:"required"`
	Amount  float64 `json:"amount" validate:"required,gt=0"`
}

type ProcessPaymentResponse struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

type GetPaymentRequest struct {
	PaymentID string `path:"id" json:"payment_id" validate:"required"`
}

type GetPaymentResponse struct {
	PaymentID string  `json:"payment_id"`
	OrderID   string  `json:"order_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// ==============================================================================
// Local Implementation
// ==============================================================================

type paymentServiceLocal struct {
	storage     string
	currency    string
	userService *service.Lazy[UserService] // Cross-server dependency
	payments    map[string]*GetPaymentResponse
}

func (s *paymentServiceLocal) ProcessPayment(ctx *request.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
	paymentID := fmt.Sprintf("payment_%d", len(s.payments)+1)
	payment := &GetPaymentResponse{
		PaymentID: paymentID,
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Status:    "completed",
	}
	s.payments[paymentID] = payment

	return &ProcessPaymentResponse{
		PaymentID: paymentID,
		OrderID:   req.OrderID,
		Amount:    req.Amount,
		Status:    "completed",
	}, nil
}

func (s *paymentServiceLocal) GetPayment(ctx *request.Context, req *GetPaymentRequest) (*GetPaymentResponse, error) {
	payment, ok := s.payments[req.PaymentID]
	if !ok {
		return nil, fmt.Errorf("payment not found")
	}
	return payment, nil
}

// ==============================================================================
// Remote Implementation
// ==============================================================================

type paymentServiceRemote struct {
	client *api_client.RemoteService
}

func (s *paymentServiceRemote) ProcessPayment(ctx *request.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error) {
	return api_client.CallRemoteService[*ProcessPaymentResponse](s.client, "ProcessPayment", ctx, req)
}

func (s *paymentServiceRemote) GetPayment(ctx *request.Context, req *GetPaymentRequest) (*GetPaymentResponse, error) {
	return api_client.CallRemoteService[*GetPaymentResponse](s.client, "GetPayment", ctx, req)
}

// ==============================================================================
// Service Factories
// ==============================================================================

// CreatePaymentServiceLocal creates local implementation
func CreatePaymentServiceLocal(cfg map[string]any) any {
	storage := utils.GetValueFromMap(cfg, "storage", "memory")
	currency := utils.GetValueFromMap(cfg, "currency", "USD")

	fmt.Printf("[payment-service] Creating LOCAL with storage: %s, currency: %s\n", storage, currency)

	// Get lazy dependency using Dep[T] with underscore key (matches YAML)
	return &paymentServiceLocal{
		storage:     storage,
		currency:    currency,
		userService: service.MustLazyLoadFromConfig[UserService](cfg, "user_service"),
		payments:    make(map[string]*GetPaymentResponse),
	}
}

// CreatePaymentServiceRemote creates HTTP client wrapper
func CreatePaymentServiceRemote(cfg map[string]any) any {
	routerName := utils.GetValueFromMap(cfg, "router", "payment-service")
	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/payments")

	fmt.Printf("[payment-service] Creating REMOTE client for router: %s, prefix: %s\n", routerName, pathPrefix)

	return &paymentServiceRemote{
		client: lokstra_registry.GetRemoteService(cfg),
	}
}

func RegisterPaymentService() {
	lokstra_registry.RegisterServiceFactoryLocalAndRemote("payment_service",
		CreatePaymentServiceLocal, CreatePaymentServiceRemote)
}
