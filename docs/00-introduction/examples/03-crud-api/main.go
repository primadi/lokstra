package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ========================================
// Models
// ========================================

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

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

// ========================================
// Service with Lazy DI
// ========================================

type UserService struct {
	DB *service.Cached[*Database]
}

// Request types for service methods
type GetByIDParams struct {
	ID int `path:"id"`
}

type CreateParams struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateParams struct {
	ID    int    `path:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type DeleteParams struct {
	ID int `path:"id"`
}

// Service methods
func (s *UserService) GetAll() ([]*User, error) {
	return s.DB.MustGet().GetAll()
}

func (s *UserService) GetByID(p *GetByIDParams) (*User, error) {
	user, err := s.DB.MustGet().GetByID(p.ID)
	if err != nil {
		return nil, fmt.Errorf("user with ID %d not found", p.ID)
	}
	return user, nil
}

func (s *UserService) Create(p *CreateParams) (*User, error) {
	user, err := s.DB.MustGet().Create(p.Name, p.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Update(p *UpdateParams) (*User, error) {
	user, err := s.DB.MustGet().Update(p.ID, p.Name, p.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Delete(p *DeleteParams) error {
	err := s.DB.MustGet().Delete(p.ID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// ========================================
// Custom Handlers (Alternative to Service-as-Router)
// ========================================

// Package-level cached service for optimal performance
// - Loaded once on first access (lazy initialization)
// - Cached for all subsequent calls (zero registry lookup cost)
// - Thread-safe via sync.Once
// - MustGet() panics with clear error if service not found (fail-fast)
var userService = service.LazyLoad[*UserService]("users")

func listUsersHandler(ctx *request.Context) error {
	users, err := userService.MustGet().GetAll()
	if err != nil {
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
	var params GetByIDParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	user, err := userService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}

	return ctx.Api.Ok(user)
}

func createUserHandler(ctx *request.Context) error {
	var params CreateParams
	if err := ctx.Req.BindBody(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request body")
	}

	user, err := userService.MustGet().Create(&params)
	if err != nil {
		if err.Error() == "email already exists" {
			return ctx.Api.Error(409, "DUPLICATE", err.Error())
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Created(user, "User created successfully")
}

func updateUserHandler(ctx *request.Context) error {
	var params UpdateParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}
	if err := ctx.Req.BindBody(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request body")
	}

	user, err := userService.MustGet().Update(&params)
	if err != nil {
		if err.Error() == "user not found" {
			return ctx.Api.Error(404, "NOT_FOUND", err.Error())
		}
		if err.Error() == "email already exists" {
			return ctx.Api.Error(409, "DUPLICATE", err.Error())
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Ok(user)
}

func deleteUserHandler(ctx *request.Context) error {
	var params DeleteParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	err := userService.MustGet().Delete(&params)
	if err != nil {
		if err.Error() == "failed to delete user: user not found" {
			return ctx.Api.Error(404, "NOT_FOUND", "User not found")
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.OkWithMessage(nil, "User deleted successfully")
}

// ========================================
// Main
// ========================================

func main() {
	// Register services
	// lokstra_registry.RegisterServiceFactory("db", func() any {
	// 	return NewDatabase()
	// })
	// register service factory: dbFactory
	lokstra_registry.RegisterServiceType("dbFactory", NewDatabase)
	// regiuster lazy service: db using dbFactory
	lokstra_registry.RegisterLazyService("db", "dbFactory", nil)

	// lokstra_registry.RegisterServiceFactory("users", func() any {
	// 	return &UserService{
	// 		DB: service.LazyLoad[*Database]("db"),
	// 	}
	// })
	// register service factory: usersFactory
	lokstra_registry.RegisterServiceType("usersFactory", func() any {
		return &UserService{
			DB: service.LazyLoad[*Database]("db"),
		}
	})
	// register lazy service: users using usersFactory
	lokstra_registry.RegisterLazyService("users", "usersFactory", nil)

	// Create router
	r := lokstra.NewRouter("api")

	// Route 1: Manual routes with custom handlers
	r.Group("/api/v1/users", func(api lokstra.Router) {
		api.GET("/", listUsersHandler)
		api.GET("/{id}", getUserHandler)
		api.POST("/", createUserHandler)
		api.PUT("/{id}", updateUserHandler)
		api.DELETE("/{id}", deleteUserHandler)
	})

	// Info endpoint
	r.GET("/", func() map[string]any {
		return map[string]any{
			"service": "User CRUD API",
			"version": "1.0.0",
			"endpoints": map[string]string{
				"GET /api/v1/users":         "List all users",
				"GET /api/v1/users/{id}":    "Get user by ID",
				"POST /api/v1/users":        "Create user",
				"PUT /api/v1/users/{id}":    "Update user",
				"DELETE /api/v1/users/{id}": "Delete user",
			},
		}
	})

	// Create app
	app := lokstra.NewApp("crud-api", ":3002", r)

	app.PrintStartInfo()
	app.Run(30 * time.Second)
}
