package domain

// OrderService defines business operations for orders
type OrderService interface {
	GetByID(p *GetOrderRequest) (*Order, error)
	List(p *ListOrdersRequest) ([]*Order, error)
	Create(p *CreateOrderRequest) (*Order, error)
	UpdateStatus(p *UpdateOrderStatusRequest) (*Order, error)
	Cancel(p *CancelOrderRequest) error
	Delete(p *DeleteOrderRequest) error
}

// OrderRepository defines data access for orders
type OrderRepository interface {
	GetByID(id int) (*Order, error)
	GetByUserID(userID int) ([]*Order, error)
	List() ([]*Order, error)
	Create(order *Order) (*Order, error)
	Update(order *Order) (*Order, error)
	Delete(id int) error
}
