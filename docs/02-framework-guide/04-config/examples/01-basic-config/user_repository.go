package main

import (
	"errors"
	"sync"
)

// User entity
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserRepository interface
type UserRepository interface {
	GetByID(id int) (*User, error)
	List() ([]*User, error)
	Create(user *User) (*User, error)
}

// UserRepositoryImpl implements UserRepository with in-memory storage
type UserRepositoryImpl struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

// NewUserRepository creates a new user repository (factory function)
func NewUserRepository(deps map[string]any, config map[string]any) any {
	repo := &UserRepositoryImpl{
		users:  make(map[int]*User),
		nextID: 1,
	}

	// Seed data
	repo.Create(&User{Name: "John Doe", Email: "john@example.com"})
	repo.Create(&User{Name: "Jane Smith", Email: "jane@example.com"})

	return repo
}

func (r *UserRepositoryImpl) GetByID(id int) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *UserRepositoryImpl) List() ([]*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepositoryImpl) Create(user *User) (*User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	return user, nil
}
