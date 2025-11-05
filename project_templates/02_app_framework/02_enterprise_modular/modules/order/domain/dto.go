package domain

// Request DTOs for Order module

type GetOrderRequest struct {
	ID int `path:"id" validate:"required"`
}

type ListOrdersRequest struct {
	UserID int    `query:"user_id"`
	Status string `query:"status"`
}

type CreateOrderRequest struct {
	UserID     int     `json:"user_id" validate:"required"`
	TotalPrice float64 `json:"total_price" validate:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	ID     int    `path:"id" validate:"required"`
	Status string `json:"status" validate:"required,oneof=pending processing completed cancelled"`
}

type CancelOrderRequest struct {
	ID     int    `path:"id" validate:"required"`
	Reason string `json:"reason"`
}

type DeleteOrderRequest struct {
	ID int `path:"id" validate:"required"`
}
