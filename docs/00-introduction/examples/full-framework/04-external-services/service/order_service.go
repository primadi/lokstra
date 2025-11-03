package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra/core/service"
)

// ========================================
// Order Models
// ========================================

type Order struct {
	ID          string    `json:"id"`
	UserID      int       `json:"user_id"`
	Items       []string  `json:"items"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`
	PaymentID   string    `json:"payment_id,omitempty"`
	Status      string    `json:"status"` // pending, paid, failed, refunded
	CreatedAt   time.Time `json:"created_at"`
}

type OrderCreateParams struct {
	UserID      int      `json:"user_id"`
	Items       []string `json:"items"`
	TotalAmount float64  `json:"total_amount"`
	Currency    string   `json:"currency"`
}

type OrderGetParams struct {
	ID string `path:"id"`
}

type OrderRefundParams struct {
	ID string `path:"id"`
}

// ========================================
// Order Service (Business Logic)
// ========================================

// OrderService handles order processing and payment integration
type OrderService struct {
	Payment *service.Cached[*PaymentServiceRemote]
}

var (
	orders   = make(map[string]*Order)
	ordersMu sync.RWMutex
	orderID  = 1
)

// Create creates a new order and processes payment via external gateway
func (s *OrderService) Create(p *OrderCreateParams) (*Order, error) {
	// Validate
	if len(p.Items) == 0 {
		return nil, fmt.Errorf("order must have at least one item")
	}
	if p.TotalAmount <= 0 {
		return nil, fmt.Errorf("total amount must be greater than 0")
	}
	if p.Currency == "" {
		p.Currency = "USD"
	}

	// Create order (pending payment)
	ordersMu.Lock()
	id := fmt.Sprintf("order_%d", orderID)
	orderID++

	order := &Order{
		ID:          id,
		UserID:      p.UserID,
		Items:       p.Items,
		TotalAmount: p.TotalAmount,
		Currency:    p.Currency,
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	orders[id] = order
	ordersMu.Unlock()

	// Process payment via external gateway
	payment, err := s.Payment.MustGet().CreatePayment(&CreatePaymentParams{
		Amount:      p.TotalAmount,
		Currency:    p.Currency,
		Description: fmt.Sprintf("Payment for order %s", id),
	})

	if err != nil {
		// Payment failed - update order status
		ordersMu.Lock()
		order.Status = "failed"
		ordersMu.Unlock()
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	// Payment succeeded - update order
	ordersMu.Lock()
	order.PaymentID = payment.ID
	order.Status = "paid"
	ordersMu.Unlock()

	return order, nil
}

// Get retrieves an order by ID
func (s *OrderService) Get(p *OrderGetParams) (*Order, error) {
	ordersMu.RLock()
	order, exists := orders[p.ID]
	ordersMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("order not found: %s", p.ID)
	}

	return order, nil
}

// Refund refunds an order via external payment gateway
func (s *OrderService) Refund(p *OrderRefundParams) (*Order, error) {
	ordersMu.Lock()
	defer ordersMu.Unlock()

	order, exists := orders[p.ID]
	if !exists {
		return nil, fmt.Errorf("order not found: %s", p.ID)
	}

	if order.Status != "paid" {
		return nil, fmt.Errorf("only paid orders can be refunded")
	}

	if order.PaymentID == "" {
		return nil, fmt.Errorf("order has no payment ID")
	}

	// Process refund via external gateway
	_, err := s.Payment.MustGet().Refund(&RefundParams{
		ID: order.PaymentID,
	})

	if err != nil {
		return nil, fmt.Errorf("refund failed: %w", err)
	}

	// Update order status
	order.Status = "refunded"

	return order, nil
}

// ========================================
// Factory
// ========================================

func OrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &OrderService{
		Payment: service.Cast[*PaymentServiceRemote](deps["payment-gateway"]),
	}
}
