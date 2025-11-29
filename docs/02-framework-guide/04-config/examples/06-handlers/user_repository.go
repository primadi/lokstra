package main

import (
	"errors"
	"sync"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UserRepository interface {
	GetByID(id int) (*User, error)
	List() ([]*User, error)
}

type UserRepositoryImpl struct {
	mu     sync.RWMutex
	users  map[int]*User
	nextID int
}

func NewUserRepository(deps map[string]any, config map[string]any) any {
	repo := &UserRepositoryImpl{
		users:  make(map[int]*User),
		nextID: 1,
	}
	repo.users[1] = &User{ID: 1, Name: "John Doe", Email: "john@example.com"}
	repo.users[2] = &User{ID: 2, Name: "Jane Smith", Email: "jane@example.com"}
	repo.nextID = 3
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
