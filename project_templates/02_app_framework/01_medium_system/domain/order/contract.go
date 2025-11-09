package order

// OrderService defines the interface for order-related operations
type OrderService interface {
	GetByID(p *GetOrderParams) (*Order, error)
	GetByUserID(p *GetOrdersByUserParams) ([]*Order, error)
	List(p *ListOrdersParams) ([]*Order, error)
	Create(p *CreateOrderParams) (*Order, error)
	UpdateStatus(p *UpdateOrderStatusParams) (*Order, error)
	Delete(p *DeleteOrderParams) error
}

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	GetByID(id int) (*Order, error)
	GetByUserID(userID int) ([]*Order, error)
	List() ([]*Order, error)
	Create(order *Order) (*Order, error)
	Update(order *Order) (*Order, error)
	Delete(id int) error
}
