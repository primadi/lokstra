package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// Type alias for request.Context - should be detected correctly!
type RequestContext = request.Context

// ============================================================================
// Models
// ============================================================================

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ============================================================================
// Request Structs
// ============================================================================

type GetUserRequest struct {
	UserID string `path:"id"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ============================================================================
// User Service - Using Type Alias for Context
// ============================================================================

type UserService struct {
	users map[string]*User
}

func NewUserService() *UserService {
	return &UserService{
		users: map[string]*User{
			"1": {ID: "1", Name: "Alice", Email: "alice@example.com"},
			"2": {ID: "2", Name: "Bob", Email: "bob@example.com"},
		},
	}
}

// Test 1: Using type alias RequestContext (should be detected as context, not struct param)
func (s *UserService) ListUsers() ([]*User, error) {
	fmt.Printf("✅ ListUsers called with type alias *RequestContext\n")
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users, nil
}

// Test 2: Using original *request.Context
func (s *UserService) GetUser(req *GetUserRequest) (*User, error) {
	fmt.Printf("✅ GetUser called with original *request.Context + struct param\n")
	user, exists := s.users[req.UserID]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// Test 3: Using type alias with struct param
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
	fmt.Printf("✅ CreateUser called with type alias *RequestContext + struct param\n")
	id := fmt.Sprintf("%d", len(s.users)+1)
	user := &User{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}
	s.users[id] = user
	return user, nil
}

// Test 4: No context, just struct param
func (s *UserService) DeleteUser(req *GetUserRequest) error {
	fmt.Printf("✅ DeleteUser called with struct param only\n")
	if _, exists := s.users[req.UserID]; !exists {
		return fmt.Errorf("user not found")
	}
	delete(s.users, req.UserID)
	return nil
}

// ============================================================================
// Main
// ============================================================================

func main() {
	// Create service
	userService := NewUserService()

	// Create router from service
	r := router.NewFromService(userService, router.DefaultServiceRouterOptions())

	fmt.Println("=== Type Alias Detection Test ===")
	fmt.Println()
	fmt.Println("Testing type comparison for *request.Context detection:")
	fmt.Println("  - type RequestContext = request.Context")
	fmt.Println()
	fmt.Println("✅ If type comparison works correctly:")
	fmt.Println("  - ListUsers: Should detect RequestContext as context (not struct param)")
	fmt.Println("  - GetUser: Should detect GetUserRequest as struct param")
	fmt.Println("  - CreateUser: Should detect CreateUserRequest as struct param")
	fmt.Println("  - DeleteUser: Should detect GetUserRequest as struct param")
	fmt.Println()
	fmt.Println("Generated Routes:")
	r.PrintRoutes()
	fmt.Println()
	fmt.Println("Server starting on :3001")

	// Create application and start
	application := app.New("type-alias-test", ":3001", r)
	if err := application.Run(0); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
