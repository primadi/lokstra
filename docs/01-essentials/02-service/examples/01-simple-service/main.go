package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ============================================================================
// Models
// ============================================================================

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin user guest"`
}

type UpdateUserRequest struct {
	ID    int    `path:"id"`
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin user guest"`
}

type GetUserRequest struct {
	ID int `path:"id"`
}

type DeleteUserRequest struct {
	ID int `path:"id"`
}

// ============================================================================
// Service Definition
// ============================================================================

type UserService struct {
	users  []User
	nextID int
}

// GetAll returns all users
func (s *UserService) GetAll() ([]User, error) {
	return s.users, nil
}

// GetByID returns a user by ID
func (s *UserService) GetByID(id int) (*User, error) {
	for _, user := range s.users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user with ID %d not found", id)
}

// Create creates a new user
func (s *UserService) Create(name, email, role string) (*User, error) {
	user := User{
		ID:    s.nextID,
		Name:  name,
		Email: email,
		Role:  role,
	}
	s.users = append(s.users, user)
	s.nextID++
	return &user, nil
}

// Update updates an existing user
func (s *UserService) Update(id int, name, email, role string) (*User, error) {
	for i, user := range s.users {
		if user.ID == id {
			s.users[i].Name = name
			s.users[i].Email = email
			s.users[i].Role = role
			return &s.users[i], nil
		}
	}
	return nil, fmt.Errorf("user with ID %d not found", id)
}

// Delete deletes a user
func (s *UserService) Delete(id int) error {
	for i, user := range s.users {
		if user.ID == id {
			s.users = append(s.users[:i], s.users[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("user with ID %d not found", id)
}

// ============================================================================
// Service Factory (Simple version - for basic examples)
// ============================================================================

func NewUserService() *UserService {
	// Initialize with sample data
	return &UserService{
		users: []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com", Role: "admin"},
			{ID: 2, Name: "Bob", Email: "bob@example.com", Role: "user"},
			{ID: 3, Name: "Charlie", Email: "charlie@example.com", Role: "user"},
		},
		nextID: 4,
	}
}

// ============================================================================
// LazyLoad Service (Package-level - Cached!)
// ============================================================================

var userService = service.LazyLoad[*UserService]("users")

// ============================================================================
// Main
// ============================================================================

func main() {
	// ========================================================================
	// STEP 1: Register Service Instance
	// ========================================================================
	// Create and register the service instance
	userSvc := NewUserService()
	lokstra_registry.RegisterService("users", userSvc)

	// ========================================================================
	// STEP 2: Create Router and Handlers
	// ========================================================================
	r := lokstra.NewRouter("api")

	// LIST: GET /users
	r.GET("/users", func() (*response.ApiHelper, error) {
		api := response.NewApiHelper()

		// Using LazyLoad - Cached after first access!
		users, err := userService.MustGet().GetAll()
		if err != nil {
			api.InternalError(err.Error())
			return api, nil
		}

		api.Ok(users)
		return api, nil
	})

	// GET: GET /users/{id}
	r.GET("/users/{id}", func(req *GetUserRequest) (*response.ApiHelper, error) {
		api := response.NewApiHelper()

		user, err := userService.MustGet().GetByID(req.ID)
		if err != nil {
			api.NotFound(err.Error())
			return api, nil
		}

		api.Ok(user)
		return api, nil
	})

	// CREATE: POST /users
	r.POST("/users", func(req *CreateUserRequest) (*response.ApiHelper, error) {
		api := response.NewApiHelper()

		user, err := userService.MustGet().Create(req.Name, req.Email, req.Role)
		if err != nil {
			api.InternalError(err.Error())
			return api, nil
		}

		api.Created(user, "User created successfully")
		return api, nil
	})

	// UPDATE: PUT /users/{id}
	r.PUT("/users/{id}", func(req *UpdateUserRequest) (*response.ApiHelper, error) {
		api := response.NewApiHelper()

		user, err := userService.MustGet().Update(req.ID, req.Name, req.Email, req.Role)
		if err != nil {
			api.NotFound(err.Error())
			return api, nil
		}

		api.Ok(user)
		return api, nil
	})

	// DELETE: DELETE /users/{id}
	r.DELETE("/users/{id}", func(req *DeleteUserRequest) (*response.ApiHelper, error) {
		api := response.NewApiHelper()

		err := userService.MustGet().Delete(req.ID)
		if err != nil {
			api.NotFound(err.Error())
			return api, nil
		}

		api.Ok(map[string]string{"message": "User deleted successfully"})
		return api, nil
	})

	// ========================================================================
	// STEP 3: Create App and Run
	// ========================================================================
	app := lokstra.NewApp("simple-service-demo", ":3000", r)

	fmt.Println("========================================")
	fmt.Println("Simple Service Example")
	fmt.Println("========================================")
	fmt.Println("Server: http://localhost:3000")
	fmt.Println()
	fmt.Println("Endpoints:")
	fmt.Println("  GET    /users       - List all users")
	fmt.Println("  GET    /users/{id}  - Get user by ID")
	fmt.Println("  POST   /users       - Create new user")
	fmt.Println("  PUT    /users/{id}  - Update user")
	fmt.Println("  DELETE /users/{id}  - Delete user")
	fmt.Println("========================================")

	app.Run(30 * time.Second)
}
