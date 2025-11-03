package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/primadi/lokstra/docs/01-router-guide/02-service/examples/03-service-dependencies/model"
)

type OrderRepository struct {
	orders       []model.Order
	nextID       int
	pricePerItem float64
}

// NewOrderRepository - Mode 2: Config only
func NewOrderRepository(cfg map[string]any) *OrderRepository {
	pricePerItem := 10.0 // default
	if price, ok := cfg["price_per_item"].(float64); ok {
		pricePerItem = price
	}

	fmt.Printf("✅ OrderRepository created (Mode 2: config only, price=%.2f)\n", pricePerItem)

	return &OrderRepository{
		orders:       []model.Order{},
		nextID:       1,
		pricePerItem: pricePerItem,
	}
}

func (r *OrderRepository) Create(userID int, items []string) (*model.Order, error) {
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	order := model.Order{
		ID:        r.nextID,
		UserID:    userID,
		Items:     items,
		Total:     float64(len(items)) * r.pricePerItem,
		CreatedAt: time.Now(),
	}

	r.orders = append(r.orders, order)
	r.nextID++

	return &order, nil
}

func (r *OrderRepository) FindByID(id int) (*model.Order, error) {
	for _, order := range r.orders {
		if order.ID == id {
			return &order, nil
		}
	}
	return nil, errors.New("order not found")
}

func (r *OrderRepository) FindAll() ([]model.Order, error) {
	return r.orders, nil
}
