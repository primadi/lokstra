package order

import "time"

// Order represents an order entity in the domain
type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Product   string    `json:"product"`
	Quantity  int       `json:"quantity"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
