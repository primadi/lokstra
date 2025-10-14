package services

import (
	"fmt"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ==============================================================================
// User Service Interface
// ==============================================================================

type UserService interface {
	CreateUser(ctx *request.Context, req *CreateUserRequest) (*CreateUserResponse, error)
	GetUser(ctx *request.Context, req *GetUserRequest) (*GetUserResponse, error)
}

// ==============================================================================
// Request/Response Structs
// ==============================================================================

type CreateUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
}

type CreateUserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type GetUserRequest struct {
	UserID string `path:"id" json:"user_id" validate:"required"`
}

type GetUserResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ==============================================================================
// Local Implementation
// ==============================================================================

type userServiceLocal struct {
	storage string
	users   map[string]*GetUserResponse
}

func (s *userServiceLocal) CreateUser(ctx *request.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	userID := fmt.Sprintf("user_%d", len(s.users)+1)
	user := &GetUserResponse{
		UserID:   userID,
		Username: req.Username,
		Email:    req.Email,
	}
	s.users[userID] = user

	return &CreateUserResponse{
		UserID:   userID,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func (s *userServiceLocal) GetUser(ctx *request.Context, req *GetUserRequest) (*GetUserResponse, error) {
	user, ok := s.users[req.UserID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// ==============================================================================
// Service Factories
// ==============================================================================

// CreateUserServiceLocal creates local implementation (for same-process calls)
func CreateUserServiceLocal(cfg map[string]any) any {
	storage := utils.GetValueFromMap(cfg, "storage", "memory")

	fmt.Printf("[user-service] Creating LOCAL with storage: %s\n", storage)

	return &userServiceLocal{
		storage: storage,
		users:   make(map[string]*GetUserResponse),
	}
}

// CreateUserServiceRemote creates HTTP client wrapper (for cross-process calls)
func CreateUserServiceRemote(cfg map[string]any) any {
	routerName := utils.GetValueFromMap(cfg, "router", "user-service")
	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/users")

	fmt.Printf("[user-service] Creating REMOTE client for router: %s, prefix: %s\n", routerName, pathPrefix)

	return &userServiceRemote{
		client: lokstra_registry.GetRemoteService(cfg),
	}
}

// ==============================================================================
// Remote Implementation (HTTP Client Wrapper)
// ==============================================================================

type userServiceRemote struct {
	client *api_client.RemoteService
}

// Pure delegation - no logic, just bridge to remote
func (s *userServiceRemote) CreateUser(ctx *request.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	return api_client.CallRemoteService[*CreateUserResponse](s.client, "CreateUser", ctx, req)
}

func (s *userServiceRemote) GetUser(ctx *request.Context, req *GetUserRequest) (*GetUserResponse, error) {
	return api_client.CallRemoteService[*GetUserResponse](s.client, "GetUser", ctx, req)
}

func RegisterUserService() {
	lokstra_registry.RegisterServiceFactoryLocalAndRemote("user_service",
		CreateUserServiceLocal, CreateUserServiceRemote)
}
