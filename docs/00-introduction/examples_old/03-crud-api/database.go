package main

import (
	"fmt"
	"sync"
	"time"
)

// ========================================
// Database (In-Memory)
// ========================================

type Database struct {
	users  map[int]*User
	nextID int
	mu     sync.RWMutex
}

func NewDatabase() *Database {
	db := &Database{
		users:  make(map[int]*User),
		nextID: 1,
	}

	// Seed data
	db.users[1] = &User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-24 * time.Hour),
	}
	db.users[2] = &User{
		ID:        2,
		Name:      "Bob",
		Email:     "bob@example.com",
		CreatedAt: time.Now().Add(-12 * time.Hour),
		UpdatedAt: time.Now().Add(-12 * time.Hour),
	}
	db.nextID = 3

	return db
}

func (db *Database) GetAll() ([]*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}
	return users, nil
}

func (db *Database) GetByID(id int) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (db *Database) Create(name, email string) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Check duplicate email
	for _, u := range db.users {
		if u.Email == email {
			return nil, fmt.Errorf("email already exists")
		}
	}

	user := &User{
		ID:        db.nextID,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	db.users[db.nextID] = user
	db.nextID++

	return user, nil
}

func (db *Database) Update(id int, name, email string) (*User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	user, exists := db.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Check duplicate email (excluding current user)
	for _, u := range db.users {
		if u.ID != id && u.Email == email {
			return nil, fmt.Errorf("email already exists")
		}
	}

	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()

	return user, nil
}

func (db *Database) Delete(id int) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.users[id]; !exists {
		return fmt.Errorf("user not found")
	}

	delete(db.users, id)
	return nil
}
