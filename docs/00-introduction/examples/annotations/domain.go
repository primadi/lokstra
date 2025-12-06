package application

import "time"

// Shared domain types for annotation examples

type User struct {
	ID    string
	Email string
	Name  string
}

type UserRepository interface {
	GetByID(id string) (*User, error)
	GetByEmail(email string) (*User, error)
	List() ([]*User, error)
	Create(user *User) (*User, error)
	Update(user *User) (*User, error)
	Delete(id string) error
}

type CacheService interface {
	Get(key string) (any, error)
	Set(key string, value any, ttl time.Duration) error
	Delete(key string) error
}
