package order

// Request/Response DTOs (Data Transfer Objects)

// GetOrderParams contains parameters for getting a single order
type GetOrderParams struct {
	ID int `path:"id" validate:"required"`
}

// GetOrdersByUserParams contains parameters for getting orders by user
type GetOrdersByUserParams struct {
	UserID int `path:"user_id" validate:"required"`
}

// ListOrdersParams contains parameters for listing orders
type ListOrdersParams struct{}

// CreateOrderParams contains parameters for creating an order
type CreateOrderParams struct {
	UserID   int     `json:"user_id" validate:"required"`
	Product  string  `json:"product" validate:"required"`
	Quantity int     `json:"quantity" validate:"required,min=1"`
	Total    float64 `json:"total" validate:"required,min=0"`
}

// UpdateOrderStatusParams contains parameters for updating order status
type UpdateOrderStatusParams struct {
	ID     int    `path:"id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=pending processing shipped delivered cancelled"`
}

// DeleteOrderParams contains parameters for deleting an order
type DeleteOrderParams struct {
	ID int `path:"id" validate:"required"`
}
