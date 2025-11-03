package repository

import (
	"errors"
	"fmt"

	"github.com/primadi/lokstra/docs/01-essentials/02-service/examples/03-service-dependencies/model"
)

type UserRepository struct {
	users []model.User
}

// NewUserRepository - Mode 1: No params (simplest!)
func NewUserRepository() *UserRepository {
	fmt.Println("✅ UserRepository created (Mode 1: no params)")
	return &UserRepository{
		users: []model.User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
			{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
		},
	}
}

func (r *UserRepository) FindByID(id int) (*model.User, error) {
	for _, user := range r.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *UserRepository) FindAll() ([]model.User, error) {
	return r.users, nil
}
