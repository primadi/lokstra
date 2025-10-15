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
// Invoice Service Interface
// ==============================================================================

type InvoiceService interface {
	GenerateInvoice(ctx *request.Context, req *GenerateInvoiceRequest) (*GenerateInvoiceResponse, error)
	GetInvoice(ctx *request.Context, req *GetInvoiceRequest) (*GetInvoiceResponse, error)
}

// ==============================================================================
// Request/Response Structs
// ==============================================================================

type GenerateInvoiceRequest struct {
	PaymentID string `json:"payment_id" validate:"required"`
}

type GenerateInvoiceResponse struct {
	InvoiceID string  `json:"invoice_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
}

type GetInvoiceRequest struct {
	InvoiceID string `path:"id" json:"invoice_id" validate:"required"`
}

type GetInvoiceResponse struct {
	InvoiceID string  `json:"invoice_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
}

// ==============================================================================
// Local Implementation
// ==============================================================================

type invoiceServiceLocal struct {
	storage        string
	paymentService *service.Cached[PaymentService] // Cross-server dependency
	invoices       map[string]*GetInvoiceResponse
}

func (s *invoiceServiceLocal) GenerateInvoice(ctx *request.Context, req *GenerateInvoiceRequest) (*GenerateInvoiceResponse, error) {
	// Get payment details (cross-service call - may be LOCAL or REMOTE!)
	var amount float64 = 0
	if s.paymentService != nil {
		paymentSvc := s.paymentService.Get()
		payment, err := paymentSvc.GetPayment(ctx, &GetPaymentRequest{PaymentID: req.PaymentID})
		if err != nil {
			return nil, err
		}
		amount = payment.Amount
	}

	invoiceID := fmt.Sprintf("invoice_%d", len(s.invoices)+1)
	invoice := &GetInvoiceResponse{
		InvoiceID: invoiceID,
		PaymentID: req.PaymentID,
		Amount:    amount,
	}
	s.invoices[invoiceID] = invoice

	return &GenerateInvoiceResponse{
		InvoiceID: invoiceID,
		PaymentID: req.PaymentID,
		Amount:    amount,
	}, nil
}

func (s *invoiceServiceLocal) GetInvoice(ctx *request.Context, req *GetInvoiceRequest) (*GetInvoiceResponse, error) {
	invoice, ok := s.invoices[req.InvoiceID]
	if !ok {
		return nil, fmt.Errorf("invoice not found")
	}
	return invoice, nil
}

// ==============================================================================
// Remote Implementation
// ==============================================================================

type invoiceServiceRemote struct {
	client *api_client.RemoteService
}

func (s *invoiceServiceRemote) GenerateInvoice(ctx *request.Context, req *GenerateInvoiceRequest) (*GenerateInvoiceResponse, error) {
	return api_client.CallRemoteService[*GenerateInvoiceResponse](s.client, "GenerateInvoice", ctx, req)
}

func (s *invoiceServiceRemote) GetInvoice(ctx *request.Context, req *GetInvoiceRequest) (*GetInvoiceResponse, error) {
	return api_client.CallRemoteService[*GetInvoiceResponse](s.client, "GetInvoice", ctx, req)
}

// ==============================================================================
// Service Factories
// ==============================================================================

// CreateInvoiceServiceLocal creates local implementation
func CreateInvoiceServiceLocal(cfg map[string]any) any {
	storage := utils.GetValueFromMap(cfg, "storage", "memory")

	fmt.Printf("[invoice-service] Creating LOCAL with storage: %s\n", storage)

	// Get lazy dependency using Dep[T] with underscore key (matches YAML)
	return &invoiceServiceLocal{
		storage:        storage,
		paymentService: service.MustLazyLoadFromConfig[PaymentService](cfg, "payment_service"),
		invoices:       make(map[string]*GetInvoiceResponse),
	}
}

// CreateInvoiceServiceRemote creates HTTP client wrapper
func CreateInvoiceServiceRemote(cfg map[string]any) any {
	routerName := utils.GetValueFromMap(cfg, "router", "invoice-service")
	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/invoices")

	fmt.Printf("[invoice-service] Creating REMOTE client for router: %s, prefix: %s\n", routerName, pathPrefix)

	return &invoiceServiceRemote{
		client: lokstra_registry.GetRemoteService(cfg),
	}
}

func RegisterInvoiceService() {
	lokstra_registry.RegisterServiceFactoryLocalAndRemote("invoice_service",
		CreateInvoiceServiceLocal, CreateInvoiceServiceRemote)
}
