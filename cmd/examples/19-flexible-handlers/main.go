package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
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

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SearchRequest struct {
	DepartmentID string `path:"dep"`
	Query        string `query:"q"`
	Page         int    `query:"page"`
}

// ============================================================================
// Service
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
// Handler Patterns - All supported by adaptSmart
// ============================================================================

// Pattern A: Multiple string path params - NEW!
// GET /departments/{dep}/users/{id}
func (s *UserService) GetUser(dep, id string) (*User, error) {
	fmt.Printf("🔍 GetUser called: dep=%q, id=%q\n", dep, id)
	if deptUsers, ok := s.users[dep]; ok {
		if user, ok := deptUsers[id]; ok {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found in department")
}

// Pattern B: Context + Multiple string params - NEW!
// DELETE /departments/{dep}/users/{id}
func (s *UserService) DeleteUser(ctx *request.Context, dep, id string) error {
	fmt.Printf("🔍 DeleteUser called: dep=%q, id=%q\n", dep, id)
	if deptUsers, ok := s.users[dep]; ok {
		if _, ok := deptUsers[id]; ok {
			delete(deptUsers, id)
			ctx.Resp.Json(map[string]any{
				"success": true,
				"message": fmt.Sprintf("User %s deleted from %s", id, dep),
			})
			return nil
		}
	}
	return fmt.Errorf("user not found")
}

// Pattern C: Struct with tags (path + query)
// GET /departments/{dep}/users/search?q=alice&page=1
func (s *UserService) SearchUsers(req *SearchRequest) ([]*User, error) {
	results := make([]*User, 0)
	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		for _, user := range deptUsers {
			results = append(results, user)
		}
	}
	return results, nil
}

// Pattern D: Context + Struct (path + body)
// POST /departments/{dep}/users
func (s *UserService) CreateUser(ctx *request.Context, dep string, req *CreateUserRequest) (*User, error) {
	if _, ok := s.users[dep]; !ok {
		s.users[dep] = make(map[string]*User)
	}

	id := fmt.Sprintf("%d", len(s.users[dep])+1)
	user := &User{
		ID:           id,
		DepartmentID: dep,
		Name:         req.Name,
		Email:        req.Email,
	}
	s.users[dep][id] = user
	return user, nil
}

// Pattern E: Just context - manual everything
// GET /departments
func (s *UserService) ListDepartments(ctx *request.Context) error {
	deps := make([]string, 0, len(s.users))
	for dep := range s.users {
		deps = append(deps, dep)
	}
	ctx.Resp.Json(map[string]any{
		"success":     true,
		"departments": deps,
	})
	return nil
}

// ============================================================================
// Main
// ============================================================================

func main() {
	service := NewUserService()
	r := router.New("flexible-handlers")

	// All these handlers are automatically adapted by adaptSmart!
	// No need for service router convention

	r.GET("/departments", service.ListDepartments)
	r.GET("/departments/{dep}/users/{id}", service.GetUser)
	r.DELETE("/departments/{dep}/users/{id}", service.DeleteUser)
	r.GET("/departments/{dep}/users/search", service.SearchUsers)
	r.POST("/departments/{dep}/users", service.CreateUser)

	app := app.New("flexible-handlers", ":3000", r)

	fmt.Println("=== Server Started on :3000 ===")
	fmt.Println("\nSupported Handler Patterns:")
	fmt.Println("✅ Pattern A: func(dep, id string) (data, error)")
	fmt.Println("✅ Pattern B: func(ctx, dep, id string) error")
	fmt.Println("✅ Pattern C: func(req *Struct) (data, error)")
	fmt.Println("✅ Pattern D: func(ctx, dep string, req *Struct) (data, error)")
	fmt.Println("✅ Pattern E: func(ctx) error")

	fmt.Println("\nEndpoints:")
	fmt.Println("  GET    http://localhost:3000/departments")
	fmt.Println("  GET    http://localhost:3000/departments/engineering/users/1")
	fmt.Println("  DELETE http://localhost:3000/departments/engineering/users/2")
	fmt.Println("  GET    http://localhost:3000/departments/sales/users/search?q=charlie")
	fmt.Println("  POST   http://localhost:3000/departments/engineering/users")

	if err := app.Run(0); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
