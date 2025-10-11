package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/router"
)

// ============================================================================
// Models
// ============================================================================

type User struct {
	ID           string `json:"id"`
	DepartmentID string `json:"department_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
}

// ============================================================================
// Request Structs with Tags for Path/Query/Body Parameters
// ============================================================================

// GetUserRequest - Path parameters via tags
type GetUserRequest struct {
	DepartmentID string `path:"dep"` // Extracted from /departments/{dep}/users/{id}
	UserID       string `path:"id"`  // Extracted from /departments/{dep}/users/{id}
}

// ListUsersRequest - Path + Query parameters
type ListUsersRequest struct {
	DepartmentID string `path:"dep"`   // Path param
	Query        string `query:"q"`    // Query param: ?q=search
	Page         int    `query:"page"` // Query param: ?page=1
}

// CreateUserRequest - Path + Body parameters
type CreateUserRequest struct {
	DepartmentID string `path:"dep"` // Path: /departments/{dep}/users
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
}

// UpdateUserRequest - Path + Body parameters
type UpdateUserRequest struct {
	DepartmentID string `path:"dep"` // Path param
	UserID       string `path:"id"`  // Path param
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
}

// DeleteUserRequest - Multiple path parameters
type DeleteUserRequest struct {
	DepartmentID string `path:"dep"`
	UserID       string `path:"id"`
}

// ============================================================================
// User Service - Struct-Based Parameters
// ============================================================================

type UserService struct {
	users map[string]map[string]*User // departmentID -> userID -> User
}

func NewUserService() *UserService {
	return &UserService{
		users: map[string]map[string]*User{
			"engineering": {
				"1": {ID: "1", DepartmentID: "engineering", Name: "Alice", Email: "alice@example.com"},
				"2": {ID: "2", DepartmentID: "engineering", Name: "Bob", Email: "bob@example.com"},
			},
			"sales": {
				"3": {ID: "3", DepartmentID: "sales", Name: "Charlie", Email: "charlie@example.com"},
			},
		},
	}
}

// ============================================================================
// Service Methods - All using Struct Parameters
// ============================================================================

// GetUser - GET /users/{dep}/{id}
// Auto-generated path from struct tags: path:"dep" + path:"id"
func (s *UserService) GetUser(req *GetUserRequest) (*User, error) {
	fmt.Printf("🔍 GetUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)

	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		if user, ok := deptUsers[req.UserID]; ok {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found in department")
}

// ListUsers - GET /users/{dep}?q=search&page=1
// Auto-generated path from struct tags: path:"dep" + query params
func (s *UserService) ListUsers(req *ListUsersRequest) ([]*User, error) {
	fmt.Printf("🔍 ListUsers: dep=%q, q=%q, page=%d\n", req.DepartmentID, req.Query, req.Page)

	deptUsers, ok := s.users[req.DepartmentID]
	if !ok {
		return []*User{}, nil
	}

	results := make([]*User, 0)
	for _, user := range deptUsers {
		if req.Query == "" || containsIgnoreCase(user.Name, req.Query) || containsIgnoreCase(user.Email, req.Query) {
			results = append(results, user)
		}
	}
	return results, nil
}

// CreateUser - POST /users/{dep}
// Auto-generated path from struct tags: path:"dep" + body from json tags
func (s *UserService) CreateUser(req *CreateUserRequest) (*User, error) {
	fmt.Printf("🔍 CreateUser: dep=%q, name=%q, email=%q\n", req.DepartmentID, req.Name, req.Email)

	if s.users[req.DepartmentID] == nil {
		s.users[req.DepartmentID] = make(map[string]*User)
	}

	id := fmt.Sprintf("%d", len(s.users[req.DepartmentID])+1)
	user := &User{
		ID:           id,
		DepartmentID: req.DepartmentID,
		Name:         req.Name,
		Email:        req.Email,
	}
	s.users[req.DepartmentID][id] = user
	return user, nil
}

// UpdateUser - PUT /users/{dep}/{id}
// Auto-generated path from struct tags: path:"dep" + path:"id" + body from json tags
func (s *UserService) UpdateUser(req *UpdateUserRequest) (*User, error) {
	fmt.Printf("🔍 UpdateUser: dep=%q, id=%q, name=%q\n", req.DepartmentID, req.UserID, req.Name)

	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		if user, ok := deptUsers[req.UserID]; ok {
			user.Name = req.Name
			user.Email = req.Email
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

// DeleteUser - DELETE /users/{dep}/{id}
// Auto-generated path from struct tags: path:"dep" + path:"id"
func (s *UserService) DeleteUser(req *DeleteUserRequest) error {
	fmt.Printf("🔍 DeleteUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)

	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		if _, ok := deptUsers[req.UserID]; ok {
			delete(deptUsers, req.UserID)
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

// ============================================================================
// Helper Functions
// ============================================================================

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}

// ============================================================================
// Main
// ============================================================================

func main() {
	// Create service
	userService := NewUserService()

	// Create router from service - Auto-generates routes from struct tags!
	r := router.NewFromService(userService, router.DefaultServiceRouterOptions())

	fmt.Println("=== Service Router with Struct-Based Parameters ===")
	fmt.Println()
	fmt.Println("✅ Benefits:")
	fmt.Println("  1. Path params extracted from struct tags: `path:\"dep\"`, `path:\"id\"`")
	fmt.Println("  2. Query params from: `query:\"q\"`")
	fmt.Println("  3. Body params from: `json:\"name\"`")
	fmt.Println("  4. Type-safe and self-documenting")
	fmt.Println("  5. No reflection limitation - tag names are explicit!")
	fmt.Println()
	fmt.Println("Generated Routes:")
	fmt.Println("  GET    /users/{dep}/{id}         - GetUser")
	fmt.Println("  GET    /users/{dep}?q=&page=     - ListUsers")
	fmt.Println("  POST   /users/{dep}              - CreateUser")
	fmt.Println("  PUT    /users/{dep}/{id}         - UpdateUser")
	fmt.Println("  DELETE /users/{dep}/{id}         - DeleteUser")
	fmt.Println()
	fmt.Println("Server starting on :3000")

	// Create application and start
	application := app.New("user-service", ":3000", r)
	if err := application.Run(0); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
