package user_service

import (
	"database/sql"
	"fmt"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ============================================================================
// Service Interface
// ============================================================================

type UserService interface {
	GetUser(ctx *request.Context, req *GetUserRequest) (*User, error)
	ListUsers(ctx *request.Context) ([]*User, error)
	CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error)
}

// ============================================================================
// Request/Response Models
// ============================================================================

type GetUserRequest struct {
	UserID string `path:"id" json:"user_id"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ============================================================================
// LOCAL Implementation - Real business logic
// ============================================================================

type LocalUserService struct {
	// Using service.Lazy[T] for type-safe lazy dependency injection
	// Database pool will be loaded on first .Get() call and cached
	db *service.Cached[sql.DB]
}

func (s *LocalUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
	fmt.Printf("📍 LOCAL: GetUser called for ID=%s\n", req.UserID)

	if req.UserID == "" {
		return nil, ctx.Api.BadRequest("INVALID_ID", "User ID is required")
	}

	// In real app, get database lazily (loaded on first call, cached thereafter):
	// db := s.db.Get()
	// row := db.QueryRow("SELECT * FROM users WHERE id = $1", req.UserID)
	// ... scan result ...

	if s.db != nil {
		fmt.Printf("   💾 Database service '%s' ready for lazy loading\n", s.db.ServiceName())
		if s.db.IsLoaded() {
			fmt.Println("   ✅ Database already loaded (cached)")
		} else {
			fmt.Println("   ⏳ Database not yet loaded (will load on first .Get())")
		}
	}

	// Simulate result for demo
	return &User{
		ID:    req.UserID,
		Name:  "Local User " + req.UserID,
		Email: "user" + req.UserID + "@local.com",
	}, nil
}

func (s *LocalUserService) ListUsers(ctx *request.Context) ([]*User, error) {
	fmt.Println("📍 LOCAL: ListUsers called")

	// In real app: db := s.db.Get() then query
	if s.db != nil {
		fmt.Printf("   💾 Database service '%s' available\n", s.db.ServiceName())
	}

	// Simulate result for demo
	return []*User{
		{ID: "1", Name: "Alice", Email: "alice@local.com"},
		{ID: "2", Name: "Bob", Email: "bob@local.com"},
	}, nil
}

func (s *LocalUserService) CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error) {
	fmt.Printf("📍 LOCAL: CreateUser called - Name=%s, Email=%s\n", req.Name, req.Email)

	// In real app: db := s.db.Get() then insert
	if s.db != nil {
		fmt.Printf("   💾 Database service '%s' available\n", s.db.ServiceName())
	}

	// Simulate result for demo
	newUser := &User{
		ID:    "new-" + req.Name,
		Name:  req.Name,
		Email: req.Email,
	}

	return newUser, nil
}

// ============================================================================
// REMOTE Implementation - HTTP client wrapper
// ============================================================================

type RemoteUserService struct {
	client *api_client.ClientRouter
}

func (s *RemoteUserService) GetUser(ctx *request.Context, req *GetUserRequest) (*User, error) {
	fmt.Printf("🌐 REMOTE: GetUser called for ID=%s (via HTTP)\n", req.UserID)

	// Call remote service via HTTP
	path := "/users/" + req.UserID
	return api_client.FetchAndCast[*User](s.client, path)
}

func (s *RemoteUserService) ListUsers(ctx *request.Context) ([]*User, error) {
	fmt.Println("🌐 REMOTE: ListUsers called (via HTTP)")

	// Call remote service via HTTP
	users, err := api_client.FetchAndCast[[]*User](s.client, "/users")
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *RemoteUserService) CreateUser(ctx *request.Context, req *CreateUserRequest) (*User, error) {
	fmt.Printf("🌐 REMOTE: CreateUser called - Name=%s (via HTTP)\n", req.Name)

	// Call remote service via HTTP
	return api_client.FetchAndCast[*User](s.client, "/users",
		api_client.WithMethod("POST"),
		api_client.WithBody(req),
	)
}

// ============================================================================
// Factory Functions - Called by Framework
// ============================================================================

// CreateLocalUserService creates a LOCAL instance with real implementation
func CreateLocalUserService(config map[string]any) any {
	fmt.Println("🏭 Factory: Creating LOCAL UserService")

	// Extract lazy service dependency from config using helper
	// This handles both GenericLazyService (from layered config) and string references
	dbLazy := service.MustLazyLoadFromConfig[sql.DB](config, "db")

	if dbLazy != nil {
		fmt.Printf("   📦 Database dependency registered: '%s' (lazy loading enabled)\n", dbLazy.ServiceName())
	}

	return &LocalUserService{
		db: dbLazy,
	}
}

// CreateRemoteUserService creates a REMOTE client that calls via HTTP
func CreateRemoteUserService(config map[string]any) any {
	fmt.Println("🏭 Factory: Creating REMOTE UserService client")

	// Get router name from config
	routerName := "UserService" // Default router name
	if rn, ok := config["router"].(string); ok {
		routerName = rn
	}

	// Get ClientRouter from registry (framework handles this)
	client := lokstra_registry.GetClientRouter(routerName)
	if client == nil {
		panic(fmt.Sprintf("ClientRouter '%s' not found in registry", routerName))
	} else {
		fmt.Printf("   ✅ ClientRouter '%s' loaded (IsLocal=%v)\n", routerName, client.IsLocal)
	}

	return &RemoteUserService{
		client: client,
	}
}
