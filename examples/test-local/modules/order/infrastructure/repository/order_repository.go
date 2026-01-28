package repository

import (
	"errors"
	"sync"
	"time"

	"examples/test-local/modules/order/domain"
)

// OrderRepositoryImpl implements domain.OrderRepository with in-memory storage
// @Service "order-repository"
type OrderRepositoryImpl struct {
	mu       sync.RWMutex
	orders   map[int]*domain.Order
	nextID   int
	byUserID map[int][]*domain.Order
}

// Ensure implementation
var _ domain.OrderRepository = (*OrderRepositoryImpl)(nil)

// NewOrderRepository creates a new in-memory order repository with seed data
func (r *OrderRepositoryImpl) Init() error {
	r.orders = make(map[int]*domain.Order)
	r.byUserID = make(map[int][]*domain.Order)
	r.nextID = 1

	// Seed data
	seedOrders := []*domain.Order{
		{ID: 0, UserID: 1, TotalPrice: 150.00, Status: "completed", CreatedAt: time.Now().AddDate(0, 0, -5)},
		{ID: 0, UserID: 2, TotalPrice: 89.99, Status: "processing", CreatedAt: time.Now().AddDate(0, 0, -2)},
		{ID: 0, UserID: 2, TotalPrice: 249.50, Status: "pending", CreatedAt: time.Now().AddDate(0, 0, -1)},
	}

	for _, o := range seedOrders {
		r.Create(o)
	}

	return nil
}

// GetByID retrieves an order by ID
func (r *OrderRepositoryImpl) GetByID(id int) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

// GetByUserID retrieves all orders for a user
func (r *OrderRepositoryImpl) GetByUserID(userID int) ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders, exists := r.byUserID[userID]
	if !exists {
		return []*domain.Order{}, nil
	}
	return orders, nil
}

// List retrieves all orders
func (r *OrderRepositoryImpl) List() ([]*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	orders := make([]*domain.Order, 0, len(r.orders))
	for _, order := range r.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

// Create creates a new order
func (r *OrderRepositoryImpl) Create(order *domain.Order) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	order.ID = r.nextID
	r.nextID++
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()

	r.orders[order.ID] = order
	r.byUserID[order.UserID] = append(r.byUserID[order.UserID], order)
	return order, nil
}

// Update updates an existing order
func (r *OrderRepositoryImpl) Update(order *domain.Order) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.orders[order.ID]
	if !exists {
		return nil, errors.New("order not found")
	}

	order.UpdatedAt = time.Now()
	order.CreatedAt = existing.CreatedAt // Preserve creation time

	// Update in main map
	r.orders[order.ID] = order

	// Update in user index
	userOrders := r.byUserID[order.UserID]
	for i, o := range userOrders {
		if o.ID == order.ID {
			userOrders[i] = order
			break
		}
	}

	return order, nil
}

// Delete deletes an order
func (r *OrderRepositoryImpl) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	order, exists := r.orders[id]
	if !exists {
		return errors.New("order not found")
	}

	delete(r.orders, id)

	// Remove from user index
	userOrders := r.byUserID[order.UserID]
	for i, o := range userOrders {
		if o.ID == id {
			r.byUserID[order.UserID] = append(userOrders[:i], userOrders[i+1:]...)
			break
		}
	}

	return nil
}
