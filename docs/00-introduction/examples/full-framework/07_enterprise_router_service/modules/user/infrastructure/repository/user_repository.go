package repository

import (
	"errors"
	"sync"

	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/07_enterprise_router_service/modules/user/domain"
)

// UserRepositoryImpl implements domain.UserRepository with in-memory storage
type UserRepositoryImpl struct {
	mu      sync.RWMutex
	users   map[int]*domain.User
	nextID  int
	byEmail map[string]*domain.User
}

// Ensure implementation
var _ domain.UserRepository = (*UserRepositoryImpl)(nil)

// NewUserRepository creates a new in-memory user repository with seed data
func NewUserRepository() *UserRepositoryImpl {
	repo := &UserRepositoryImpl{
		users:   make(map[int]*domain.User),
		byEmail: make(map[string]*domain.User),
		nextID:  1,
	}

	// Seed data
	seedUsers := []*domain.User{
		{ID: 0, Name: "Admin User", Email: "admin@example.com", Status: "active", RoleID: 1},
		{ID: 0, Name: "John Doe", Email: "john@example.com", Status: "active", RoleID: 2},
		{ID: 0, Name: "Jane Smith", Email: "jane@example.com", Status: "active", RoleID: 2},
	}

	for _, u := range seedUsers {
		repo.Create(u)
	}

	return repo
}

// GetByID retrieves a user by ID
func (r *UserRepositoryImpl) GetByID(id int) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.byEmail[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// List retrieves all users
func (r *UserRepositoryImpl) List() ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}
	return users, nil
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if email already exists
	if _, exists := r.byEmail[user.Email]; exists {
		return nil, errors.New("email already exists")
	}

	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	r.byEmail[user.Email] = user
	return user, nil
}

// Update updates an existing user
func (r *UserRepositoryImpl) Update(user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.users[user.ID]
	if !exists {
		return nil, errors.New("user not found")
	}

	// If email changed, check for conflicts
	if existing.Email != user.Email {
		if _, exists := r.byEmail[user.Email]; exists {
			return nil, errors.New("email already exists")
		}
		delete(r.byEmail, existing.Email)
		r.byEmail[user.Email] = user
	}

	r.users[user.ID] = user
	return user, nil
}

// Delete deletes a user
func (r *UserRepositoryImpl) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return errors.New("user not found")
	}

	delete(r.users, id)
	delete(r.byEmail, user.Email)
	return nil
}

// UserRepositoryFactory creates a new UserRepositoryImpl instance
func UserRepositoryFactory(deps map[string]any, config map[string]any) any {
	return NewUserRepository()
}
