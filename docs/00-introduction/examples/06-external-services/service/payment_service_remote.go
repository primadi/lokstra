package service

import (
	"time"

	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/service"
)

// ========================================
// Payment Models
// ========================================

type Payment struct {
	ID          string     `json:"id"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty"`
}

type CreatePaymentParams struct {
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type GetPaymentParams struct {
	ID string `path:"id"`
}

type RefundParams struct {
	ID string `path:"id"`
}

type RefundResponse struct {
	PaymentID  string    `json:"payment_id"`
	RefundedAt time.Time `json:"refunded_at"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
}

// ========================================
// Payment Service Remote (External API)
// ========================================

// PaymentServiceRemote is a proxy to external payment gateway
// This demonstrates wrapping a third-party API as a Lokstra service
type PaymentServiceRemote struct {
	proxyService *proxy.Service
}

// NewPaymentServiceRemote creates a new payment service proxy
func NewPaymentServiceRemote(proxyService *proxy.Service) *PaymentServiceRemote {
	return &PaymentServiceRemote{
		proxyService: proxyService,
	}
}

// CreatePayment creates a new payment via external gateway
func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
	// Uses custom route: POST /payments (from RegisterServiceType)
	return proxy.CallWithData[*Payment](s.proxyService, "CreatePayment", p)
}

// GetPayment retrieves payment status
func (s *PaymentServiceRemote) GetPayment(p *GetPaymentParams) (*Payment, error) {
	// Uses custom route: GET /payments/{id} (from RegisterServiceType)
	return proxy.CallWithData[*Payment](s.proxyService, "GetPayment", p)
}

// Refund processes a refund for a payment
func (s *PaymentServiceRemote) Refund(p *RefundParams) (*RefundResponse, error) {
	// Uses custom route: POST /payments/{id}/refund (from RegisterServiceType)
	return proxy.CallWithData[*RefundResponse](s.proxyService, "Refund", p)
}

// ========================================
// Factory for External Service
// ========================================

// PaymentServiceRemoteFactory creates a new PaymentServiceRemote instance
// Framework passes proxy.Service via config["remote"]
func PaymentServiceRemoteFactory(deps map[string]any, config map[string]any) any {
	return NewPaymentServiceRemote(
		service.CastProxyService(config["remote"]),
	)
}
