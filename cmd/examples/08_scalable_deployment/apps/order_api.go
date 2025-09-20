package apps

import (
	"github.com/primadi/lokstra/core/request"
)

// Order API handlers
func GetOrders(ctx *request.Context) error {
	orders := []map[string]any{
		{"id": 1, "user_id": 1, "product": "Laptop", "amount": 999.99, "status": "pending"},
		{"id": 2, "user_id": 2, "product": "Phone", "amount": 599.99, "status": "completed"},
	}

	return ctx.Ok(orders)
}

func GetOrder(ctx *request.Context) error {
	id := ctx.GetPathParam("id")

	order := map[string]any{
		"id":      id,
		"user_id": 1,
		"product": "Laptop",
		"amount":  999.99,
		"status":  "pending",
	}

	return ctx.Ok(order)
}

type createOrderRequest struct {
	Product string  `body:"product"`
	Amount  float64 `body:"amount"`
}

// CreateOrder handles the creation of a new order
// It expects a JSON body with "product" and "amount" fields
func CreateOrder(ctx *request.Context, req *createOrderRequest) error {
	// Simulate order creation
	order := map[string]any{
		"id":      456,
		"product": req.Product,
		"amount":  req.Amount,
		"status":  "pending",
	}

	return ctx.OkCreated(order)
}
