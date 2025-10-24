package service

import (
	"time"

	"github.com/primadi/lokstra/core/proxy"
	"github.com/primadi/lokstra/core/router/autogen"
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
	service.RemoteServiceMetaAdapter
}

// NewPaymentServiceRemote creates a new payment service proxy
func NewPaymentServiceRemote(proxyService *proxy.Service) *PaymentServiceRemote {
	return &PaymentServiceRemote{
		RemoteServiceMetaAdapter: service.RemoteServiceMetaAdapter{
			Resource:     "payment",
			Plural:       "payments",
			Convention:   "rest",
			ProxyService: proxyService,
			// Route overrides for external API endpoints
			// External APIs often have non-standard method names
			Override: autogen.RouteOverride{
				Custom: map[string]autogen.Route{
					// Map CreatePayment -> POST /payments
					"CreatePayment": {Method: "POST", Path: "/payments"},
					// Map GetPayment -> GET /payments/{id}
					"GetPayment": {Method: "GET", Path: "/payments/{id}"},
					// Map Refund -> POST /payments/{id}/refund
					"Refund": {Method: "POST", Path: "/payments/{id}/refund"},
				},
			},
		},
	}
}

// CreatePayment creates a new payment via external gateway
func (s *PaymentServiceRemote) CreatePayment(p *CreatePaymentParams) (*Payment, error) {
	// Auto-generates: POST /payments
	return proxy.CallWithData[*Payment](s.GetProxyService(), "CreatePayment", p)
}

// GetPayment retrieves payment status
func (s *PaymentServiceRemote) GetPayment(p *GetPaymentParams) (*Payment, error) {
	// Auto-generates: GET /payments/{id}
	return proxy.CallWithData[*Payment](s.GetProxyService(), "GetPayment", p)
}

// Refund processes a refund for a payment
func (s *PaymentServiceRemote) Refund(p *RefundParams) (*RefundResponse, error) {
	// Uses custom route: POST /payments/{id}/refund
	return proxy.CallWithData[*RefundResponse](s.GetProxyService(), "Refund", p)
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
