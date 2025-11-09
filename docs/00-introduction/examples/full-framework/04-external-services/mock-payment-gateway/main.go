package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/primadi/lokstra"
)

// ========================================
// Models
// ========================================

type Payment struct {
	ID          string     `json:"id"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"` // completed, refunded
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	RefundedAt  *time.Time `json:"refunded_at,omitempty"`
}

type CreatePaymentRequest struct {
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Currency    string  `json:"currency"`
	Description string  `json:"description"`
}

type GetPaymentRequest struct {
	ID string `path:"id"`
}

type RefundRequest struct {
	ID string `path:"id"`
}

type RefundResponse struct {
	PaymentID  string    `json:"payment_id"`
	RefundedAt time.Time `json:"refunded_at"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
}

// ========================================
// Storage
// ========================================

var (
	payments   = make(map[string]*Payment)
	paymentsMu sync.RWMutex
	nextID     = 1
)

// ========================================
// Handlers
// ========================================

func createPayment(req *CreatePaymentRequest) (*Payment, error) {
	// Set default currency
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// Create payment
	paymentsMu.Lock()
	id := fmt.Sprintf("pay_%d", nextID)
	nextID++

	payment := &Payment{
		ID:          id,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      "completed", // Simulate instant success
		Description: req.Description,
		CreatedAt:   time.Now(),
	}
	payments[id] = payment
	paymentsMu.Unlock()

	log.Printf("✅ Payment created: %s - $%.2f %s", id, req.Amount, req.Currency)

	return payment, nil
}

func getPayment(req *GetPaymentRequest) (*Payment, error) {
	paymentsMu.RLock()
	payment, exists := payments[req.ID]
	paymentsMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("payment not found: %s", req.ID)
	}

	log.Printf("📄 Payment retrieved: %s - Status: %s", req.ID, payment.Status)

	return payment, nil
}

func refundPayment(req *RefundRequest) (*RefundResponse, error) {
	paymentsMu.Lock()
	defer paymentsMu.Unlock()

	payment, exists := payments[req.ID]
	if !exists {
		return nil, fmt.Errorf("payment not found: %s", req.ID)
	}

	if payment.Status == "refunded" {
		return nil, fmt.Errorf("payment already refunded")
	}

	if payment.Status != "completed" {
		return nil, fmt.Errorf("only completed payments can be refunded")
	}

	// Process refund
	now := time.Now()
	payment.Status = "refunded"
	payment.RefundedAt = &now

	log.Printf("💸 Payment refunded: %s - $%.2f %s", req.ID, payment.Amount, payment.Currency)

	return &RefundResponse{
		PaymentID:  req.ID,
		RefundedAt: now,
		Status:     "refunded",
		Message:    fmt.Sprintf("Payment %s has been refunded", req.ID),
	}, nil
}

// ========================================
// Main
// ========================================

func main() {
	// Create router
	r := lokstra.NewRouter("payment-api")

	// Routes
	r.POST("/payments", createPayment)
	r.GET("/payments/{id}", getPayment)
	r.POST("/payments/{id}/refund", refundPayment)

	// Start server
	app := lokstra.NewApp("payment-gateway", ":9000", r)

	fmt.Println("==========================================================")
	fmt.Println("🏦 Mock Payment Gateway (Lokstra)")
	fmt.Println("==========================================================")
	fmt.Println()
	fmt.Println("Running on: http://localhost:9000")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  POST   /payments              - Create payment")
	fmt.Println("  GET    /payments/{id}         - Get payment status")
	fmt.Println("  POST   /payments/{id}/refund  - Refund payment")
	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Println()

	if err := app.Run(30 * time.Second); err != nil {
		log.Fatalf("Failed to run app: %v", err)
	}
}
