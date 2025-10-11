package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/core/app"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
	"github.com/primadi/lokstra/lokstra_registry"
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
// Request/Response Structs with Path Tags
// ============================================================================

// GetUserRequest with path tags - will be detected by RemoteService
type GetUserRequest struct {
	DepartmentID string `path:"dep"` // Extracted from URL path
	UserID       string `path:"id"`  // Extracted from URL path
}

type GetUserResponse struct {
	Success bool   `json:"success"`
	Data    *User  `json:"data"`
	Message string `json:"message,omitempty"`
}

// ListUsersRequest with path tag + query params
type ListUsersRequest struct {
	DepartmentID string `path:"dep"`   // Path param
	Query        string `query:"q"`    // Query param
	Page         int    `query:"page"` // Query param
}

type ListUsersResponse struct {
	Success bool    `json:"success"`
	Data    []*User `json:"data"`
	Total   int     `json:"total"`
}

// CreateUserRequest with path tag + body
type CreateUserRequest struct {
	DepartmentID string `path:"dep"` // Path param
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
}

type CreateUserResponse struct {
	Success bool   `json:"success"`
	Data    *User  `json:"data"`
	Message string `json:"message,omitempty"`
}

// UpdateUserRequest with multiple path tags + body
type UpdateUserRequest struct {
	DepartmentID string `path:"dep"` // Path param
	UserID       string `path:"id"`  // Path param
	Name         string `json:"name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
}

type UpdateUserResponse struct {
	Success bool   `json:"success"`
	Data    *User  `json:"data"`
	Message string `json:"message,omitempty"`
}

// DeleteUserRequest with path tags
type DeleteUserRequest struct {
	DepartmentID string `path:"dep"`
	UserID       string `path:"id"`
}

type DeleteUserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ============================================================================
// User Service Implementation (Server-Side)
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

func (s *UserService) GetUser(req *GetUserRequest) (*GetUserResponse, error) {
	fmt.Printf("🔍 [Server] GetUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)

	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		if user, ok := deptUsers[req.UserID]; ok {
			return &GetUserResponse{
				Success: true,
				Data:    user,
			}, nil
		}
	}
	return &GetUserResponse{
		Success: false,
		Message: "user not found",
	}, nil
}

func (s *UserService) ListUsers(req *ListUsersRequest) (*ListUsersResponse, error) {
	fmt.Printf("🔍 [Server] ListUsers: dep=%q, q=%q, page=%d\n", req.DepartmentID, req.Query, req.Page)

	deptUsers, ok := s.users[req.DepartmentID]
	if !ok {
		return &ListUsersResponse{
			Success: true,
			Data:    []*User{},
			Total:   0,
		}, nil
	}

	results := make([]*User, 0)
	for _, user := range deptUsers {
		results = append(results, user)
	}

	return &ListUsersResponse{
		Success: true,
		Data:    results,
		Total:   len(results),
	}, nil
}

func (s *UserService) CreateUser(req *CreateUserRequest) (*CreateUserResponse, error) {
	fmt.Printf("🔍 [Server] CreateUser: dep=%q, name=%q, email=%q\n", req.DepartmentID, req.Name, req.Email)

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

	return &CreateUserResponse{
		Success: true,
		Data:    user,
		Message: "User created successfully",
	}, nil
}

func (s *UserService) UpdateUser(req *UpdateUserRequest) (*UpdateUserResponse, error) {
	fmt.Printf("🔍 [Server] UpdateUser: dep=%q, id=%q, name=%q\n", req.DepartmentID, req.UserID, req.Name)

	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		if user, ok := deptUsers[req.UserID]; ok {
			user.Name = req.Name
			user.Email = req.Email
			return &UpdateUserResponse{
				Success: true,
				Data:    user,
				Message: "User updated successfully",
			}, nil
		}
	}

	return &UpdateUserResponse{
		Success: false,
		Message: "user not found",
	}, nil
}

func (s *UserService) DeleteUser(req *DeleteUserRequest) (*DeleteUserResponse, error) {
	fmt.Printf("🔍 [Server] DeleteUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)

	if deptUsers, ok := s.users[req.DepartmentID]; ok {
		if _, ok := deptUsers[req.UserID]; ok {
			delete(deptUsers, req.UserID)
			return &DeleteUserResponse{
				Success: true,
				Message: "User deleted successfully",
			}, nil
		}
	}

	return &DeleteUserResponse{
		Success: false,
		Message: "user not found",
	}, nil
}

// ============================================================================
// Remote User Service (Client-Side)
// ============================================================================

type RemoteUserService struct {
	client *api_client.RemoteService
}

func NewRemoteUserService(clientRouter *api_client.ClientRouter, basePath string) *RemoteUserService {
	return &RemoteUserService{
		client: api_client.NewRemoteService(clientRouter, basePath),
	}
}

// GetUser calls remote service - path params auto-detected from struct tags!
func (s *RemoteUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*GetUserResponse, error) {
	fmt.Printf("📡 [Client] GetUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)
	resp, err := api_client.CallRemoteService[*GetUserResponse](s.client, "GetUser", ctx, req)
	return resp, err
}

func (s *RemoteUserService) ListUsers(ctx *request.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	fmt.Printf("📡 [Client] ListUsers: dep=%q\n", req.DepartmentID)
	resp, err := api_client.CallRemoteService[*ListUsersResponse](s.client, "ListUsers", ctx, req)
	return resp, err
}

func (s *RemoteUserService) CreateUser(ctx *request.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	fmt.Printf("📡 [Client] CreateUser: dep=%q, name=%q\n", req.DepartmentID, req.Name)
	resp, err := api_client.CallRemoteService[*CreateUserResponse](s.client, "CreateUser", ctx, req)
	return resp, err
}

func (s *RemoteUserService) UpdateUser(ctx *request.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
	fmt.Printf("📡 [Client] UpdateUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)
	resp, err := api_client.CallRemoteService[*UpdateUserResponse](s.client, "UpdateUser", ctx, req)
	return resp, err
}

func (s *RemoteUserService) DeleteUser(ctx *request.Context, req *DeleteUserRequest) (*DeleteUserResponse, error) {
	fmt.Printf("📡 [Client] DeleteUser: dep=%q, id=%q\n", req.DepartmentID, req.UserID)
	resp, err := api_client.CallRemoteService[*DeleteUserResponse](s.client, "DeleteUser", ctx, req)
	return resp, err
}

// ============================================================================
// Main
// ============================================================================

func main() {
	fmt.Println("=== Remote Service with Path Tag Detection ===")
	fmt.Println()

	// 1. Start server in background
	go startServer()
	time.Sleep(1 * time.Second) // Wait for server to start

	// 2. Setup client
	remoteService := setupClient()

	// 3. Test remote calls
	testRemoteCalls(remoteService)
}

func startServer() {
	fmt.Println("🚀 Starting server on :8080...")

	// Create service and router
	userService := NewUserService()
	r := router.NewFromService(userService, router.DefaultServiceRouterOptions())

	// Create and run application
	application := app.New("user-service", ":8080", r)
	if err := application.Run(0); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func setupClient() *RemoteUserService {
	fmt.Println("📡 Setting up client...")

	// Register client router
	lokstra_registry.RegisterClientRouter(
		"user-service",          // routerName
		"user-service-server",   // serverName
		"http://localhost:8080", // baseURL
		"",                      // addr (empty since baseURL already has full path)
		30*time.Second,          // timeout
	)

	// Build running registry
	lokstra_registry.SetCurrentServerName("user-service-server")

	// Create remote service
	client := lokstra_registry.GetClientRouter("user-service")
	if client == nil {
		log.Fatal("Failed to get client router")
	}
	return NewRemoteUserService(client, "/users")
}

func testRemoteCalls(service *RemoteUserService) {
	fmt.Println()
	fmt.Println("=== Testing Remote Calls with Path Tag Detection ===")
	fmt.Println()

	ctx := &request.Context{}

	// Test 1: GetUser - path params auto-detected: /users/{dep}/{id}
	fmt.Println("1️⃣ Test GetUser (path params: dep=engineering, id=1)")
	resp1, err := service.GetUser(ctx, &GetUserRequest{
		DepartmentID: "engineering",
		UserID:       "1",
	})
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: %+v\n", resp1.Data)
	}
	fmt.Println()

	// Test 2: ListUsers - path param + query params
	fmt.Println("2️⃣ Test ListUsers (path: dep=engineering, query: page=1)")
	resp2, err := service.ListUsers(ctx, &ListUsersRequest{
		DepartmentID: "engineering",
		Page:         1,
	})
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: Found %d users\n", resp2.Total)
		for _, user := range resp2.Data {
			fmt.Printf("      - %s (%s)\n", user.Name, user.Email)
		}
	}
	fmt.Println()

	// Test 3: CreateUser - path param + body
	fmt.Println("3️⃣ Test CreateUser (path: dep=engineering, body: name, email)")
	resp3, err := service.CreateUser(ctx, &CreateUserRequest{
		DepartmentID: "engineering",
		Name:         "David",
		Email:        "david@example.com",
	})
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: %+v\n", resp3.Data)
	}
	fmt.Println()

	// Test 4: UpdateUser - multiple path params + body
	fmt.Println("4️⃣ Test UpdateUser (path: dep=engineering, id=1, body: name, email)")
	resp4, err := service.UpdateUser(ctx, &UpdateUserRequest{
		DepartmentID: "engineering",
		UserID:       "1",
		Name:         "Alice Updated",
		Email:        "alice.updated@example.com",
	})
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: %+v\n", resp4.Data)
	}
	fmt.Println()

	// Test 5: DeleteUser - path params
	fmt.Println("5️⃣ Test DeleteUser (path: dep=sales, id=3)")
	resp5, err := service.DeleteUser(ctx, &DeleteUserRequest{
		DepartmentID: "sales",
		UserID:       "3",
	})
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
	} else {
		fmt.Printf("   ✅ Success: %s\n", resp5.Message)
	}
	fmt.Println()

	fmt.Println("=== Summary ===")
	fmt.Println("✅ Path tags detected automatically from struct!")
	fmt.Println("✅ HTTP methods inferred from method names!")
	fmt.Println("✅ REST-style paths generated: /users/{dep}/{id}")
	fmt.Println("✅ No manual path construction needed!")
}
