package repository

import (
	"fmt"
	"log"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/domain/user"
)

// UserRepositoryMemory implements user.UserRepository using in-memory storage
type UserRepositoryMemory struct {
	users  map[int]*user.User
	nextID int
}

// Ensure implementation
var _ user.UserRepository = (*UserRepositoryMemory)(nil)

// NewUserRepositoryMemory creates a new in-memory user repository with seed data
func NewUserRepositoryMemory(config map[string]any) *UserRepositoryMemory {
	dsn := utils.GetValueFromMap(config, "dsn", "memory://users")
	log.Printf("⚙️  Initializing UserRepositoryMemory with DSN: %s", dsn)

	repo := &UserRepositoryMemory{
		users:  make(map[int]*user.User),
		nextID: 3,
	}

	// Seed users
	repo.users[1] = &user.User{ID: 1, Name: "Alice Johnson", Email: "alice@example.com"}
	repo.users[2] = &user.User{ID: 2, Name: "Bob Smith", Email: "bob@example.com"}

	return repo
}

// GetByID retrieves a user by ID
func (r *UserRepositoryMemory) GetByID(id int) (*user.User, error) {
	u, exists := r.users[id]
	if !exists {
		return nil, api_client.NewApiError(404, "NOT_FOUND", fmt.Sprintf("user with ID %d not found", id))
	}
	return u, nil
}

// List retrieves all users
func (r *UserRepositoryMemory) List() ([]*user.User, error) {
	users := make([]*user.User, 0, len(r.users))
	for _, u := range r.users {
		users = append(users, u)
	}
	return users, nil
}

// Create creates a new user
func (r *UserRepositoryMemory) Create(u *user.User) (*user.User, error) {
	u.ID = r.nextID
	r.nextID++
	r.users[u.ID] = u
	return u, nil
}

// Update updates an existing user
func (r *UserRepositoryMemory) Update(u *user.User) (*user.User, error) {
	if _, exists := r.users[u.ID]; !exists {
		return nil, api_client.NewApiError(404, "NOT_FOUND", fmt.Sprintf("user with ID %d not found", u.ID))
	}
	r.users[u.ID] = u
	return u, nil
}

// Delete deletes a user
func (r *UserRepositoryMemory) Delete(id int) error {
	if _, exists := r.users[id]; !exists {
		return api_client.NewApiError(404, "NOT_FOUND", fmt.Sprintf("user with ID %d not found", id))
	}
	delete(r.users, id)
	return nil
}
