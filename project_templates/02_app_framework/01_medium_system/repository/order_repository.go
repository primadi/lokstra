package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain/order"
)

// OrderRepositoryMemory implements order.OrderRepository using in-memory storage
type OrderRepositoryMemory struct {
	orders map[int]*order.Order
	nextID int
}

// Ensure implementation
var _ order.OrderRepository = (*OrderRepositoryMemory)(nil)

// NewOrderRepositoryMemory creates a new in-memory order repository with seed data
func NewOrderRepositoryMemory(config map[string]any) *OrderRepositoryMemory {
	dsn := utils.GetValueFromMap(config, "dsn", "memory://orders")
	log.Printf("⚙️  Initializing OrderRepositoryMemory with DSN: %s", dsn)

	repo := &OrderRepositoryMemory{
		orders: make(map[int]*order.Order),
		nextID: 3,
	}

	// Seed orders
	repo.orders[1] = &order.Order{
		ID:        1,
		UserID:    1,
		Product:   "Laptop",
		Quantity:  1,
		Total:     1200.00,
		Status:    "delivered",
		CreatedAt: time.Now().Add(-48 * time.Hour),
	}
	repo.orders[2] = &order.Order{
		ID:        2,
		UserID:    1,
		Product:   "Mouse",
		Quantity:  2,
		Total:     50.00,
		Status:    "shipped",
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	return repo
}

// GetByID retrieves an order by ID
func (r *OrderRepositoryMemory) GetByID(id int) (*order.Order, error) {
	o, exists := r.orders[id]
	if !exists {
		return nil, api_client.NewApiError(404, "NOT_FOUND", fmt.Sprintf("order with ID %d not found", id))
	}
	return o, nil
}

// GetByUserID retrieves all orders for a specific user
func (r *OrderRepositoryMemory) GetByUserID(userID int) ([]*order.Order, error) {
	orders := make([]*order.Order, 0)
	for _, o := range r.orders {
		if o.UserID == userID {
			orders = append(orders, o)
		}
	}
	return orders, nil
}

// List retrieves all orders
func (r *OrderRepositoryMemory) List() ([]*order.Order, error) {
	orders := make([]*order.Order, 0, len(r.orders))
	for _, o := range r.orders {
		orders = append(orders, o)
	}
	return orders, nil
}

// Create creates a new order
func (r *OrderRepositoryMemory) Create(o *order.Order) (*order.Order, error) {
	o.ID = r.nextID
	r.nextID++
	o.CreatedAt = time.Now()
	if o.Status == "" {
		o.Status = "pending"
	}
	r.orders[o.ID] = o
	return o, nil
}

// Update updates an existing order
func (r *OrderRepositoryMemory) Update(o *order.Order) (*order.Order, error) {
	if _, exists := r.orders[o.ID]; !exists {
		return nil, api_client.NewApiError(404, "NOT_FOUND", fmt.Sprintf("order with ID %d not found", o.ID))
	}
	r.orders[o.ID] = o
	return o, nil
}

// Delete deletes an order
func (r *OrderRepositoryMemory) Delete(id int) error {
	if _, exists := r.orders[id]; !exists {
		return api_client.NewApiError(404, "NOT_FOUND", fmt.Sprintf("order with ID %d not found", id))
	}
	delete(r.orders, id)
	return nil
}
