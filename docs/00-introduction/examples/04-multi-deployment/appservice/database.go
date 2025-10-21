package appservice

import (
	"fmt"
	"sync"

	"github.com/primadi/lokstra/api_client"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Order struct {
	ID      int     `json:"id"`
	UserID  int     `json:"user_id"`
	Product string  `json:"product"`
	Amount  float64 `json:"amount"`
}

// ========================================
// Database (In-Memory)
// ========================================

type Database struct {
	users  map[int]*User
	orders map[int]*Order
	mu     sync.RWMutex
}

func NewDatabase() *Database {
	db := &Database{
		users:  make(map[int]*User),
		orders: make(map[int]*Order),
	}

	// Seed users
	db.users[1] = &User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	db.users[2] = &User{ID: 2, Name: "Bob", Email: "bob@example.com"}

	// Seed orders
	db.orders[1] = &Order{ID: 1, UserID: 1, Product: "Laptop", Amount: 1200.00}
	db.orders[2] = &Order{ID: 2, UserID: 1, Product: "Mouse", Amount: 25.00}
	db.orders[3] = &Order{ID: 3, UserID: 2, Product: "Keyboard", Amount: 75.00}

	return db
}

func (db *Database) GetUser(id int) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, api_client.NewApiError(404, "NOT_FOUND",
			"user not found")
	}
	return user, nil
}

func (db *Database) GetAllUsers() ([]*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}
	return users, nil
}

func (db *Database) GetOrder(id int) (*Order, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	order, exists := db.orders[id]
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (db *Database) GetOrdersByUser(userID int) ([]*Order, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	orders := make([]*Order, 0)
	for _, order := range db.orders {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}
	return orders, nil
}
