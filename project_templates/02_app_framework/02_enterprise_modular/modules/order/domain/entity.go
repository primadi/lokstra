package domain

import "time"

// Order represents an order entity in the order bounded context
type Order struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	UserName   string    `json:"user_name"`
	TotalPrice float64   `json:"total_price"`
	Status     string    `json:"status"` // pending, processing, completed, cancelled
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// OrderItem represents a line item in an order
type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Subtotal  float64 `json:"subtotal"`
}
