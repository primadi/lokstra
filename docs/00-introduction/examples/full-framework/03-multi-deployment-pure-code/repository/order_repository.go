package repository

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/model"
)

// ========================================
// Order Repository Interface
// ========================================

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	GetByID(id int) (*model.Order, error)
	GetByUserID(userID int) ([]*model.Order, error)
}

// ========================================
// In-Memory Implementation
// ========================================

// OrderRepositoryMemory implements OrderRepository using in-memory storage
type OrderRepositoryMemory struct {
	orders map[int]*model.Order
}

// Ensure implementation
var _ OrderRepository = (*OrderRepositoryMemory)(nil)

// NewOrderRepositoryMemory creates a new in-memory order repository with seed data
func NewOrderRepositoryMemory(config map[string]any) *OrderRepositoryMemory {
	dsn := utils.GetValueFromMap(config, "dsn", "")
	log.Printf("⚙️  Initializing OrderRepositoryMemory with DSN: %s\n", dsn)
	repo := &OrderRepositoryMemory{
		orders: make(map[int]*model.Order),
	}

	// Seed orders
	repo.orders[1] = &model.Order{ID: 1, UserID: 1, Product: "Laptop", Amount: 1200.00}
	repo.orders[2] = &model.Order{ID: 2, UserID: 1, Product: "Mouse", Amount: 25.00}
	repo.orders[3] = &model.Order{ID: 3, UserID: 2, Product: "Keyboard", Amount: 75.00}

	return repo
}

// GetByID retrieves an order by ID
func (r *OrderRepositoryMemory) GetByID(id int) (*model.Order, error) {
	order, exists := r.orders[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

// GetByUserID retrieves all orders for a specific user
func (r *OrderRepositoryMemory) GetByUserID(userID int) ([]*model.Order, error) {
	orders := make([]*model.Order, 0)
	for _, order := range r.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
