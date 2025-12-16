package repository

import (
	"github.com/primadi/lokstra/common/api_client"
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/model"
)

// ========================================
// User Repository Interface
// ========================================

// UserRepository defines the interface for user data access
type UserRepository interface {
	GetByID(id int) (*model.User, error)
	List() ([]*model.User, error)
}

// ========================================
// In-Memory Implementation
// ========================================

// UserRepositoryMemory implements UserRepository using in-memory storage
type UserRepositoryMemory struct {
	users map[int]*model.User
}

// Ensure implementation
var _ UserRepository = (*UserRepositoryMemory)(nil)

// NewUserRepositoryMemory creates a new in-memory user repository with seed data
func NewUserRepositoryMemory(config map[string]any) *UserRepositoryMemory {
	dsn := utils.GetValueFromMap(config, "dsn", "")
	logger.LogInfo("⚙️  Initializing UserRepositoryMemory with DSN: %s", dsn)

	repo := &UserRepositoryMemory{
		users: make(map[int]*model.User),
	}

	// Seed users
	repo.users[1] = &model.User{ID: 1, Name: "Alice", Email: "alice@example.com"}
	repo.users[2] = &model.User{ID: 2, Name: "Bob", Email: "bob@example.com"}

	return repo
}

// GetByID retrieves a user by ID
func (r *UserRepositoryMemory) GetByID(id int) (*model.User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, api_client.NewApiError(404, "NOT_FOUND", "user not found")
	}
	return user, nil
}

// List retrieves all users
func (r *UserRepositoryMemory) List() ([]*model.User, error) {
	users := make([]*model.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}
