package services

import (
	"fmt"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ==============================================================================
// Auth Service Interface
// ==============================================================================

type AuthService interface {
	Login(ctx *request.Context, req *LoginRequest) (*LoginResponse, error)
	Logout(ctx *request.Context, req *LogoutRequest) (*LogoutResponse, error)
	ValidateToken(ctx *request.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error)
}

// ==============================================================================
// Request/Response Structs
// ==============================================================================

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

type LogoutRequest struct {
	Token string `header:"Authorization" validate:"required"`
}

type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ValidateTokenRequest struct {
	Token string `json:"token" validate:"required"`
}

type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
}

// ==============================================================================
// Local Implementation
// ==============================================================================

type authServiceLocal struct {
	jwtSecret   string
	tokenExpiry int
	userService *service.Cached[UserService] // Lazy dependency injection
	tokens      map[string]string            // token -> userID
}

func (s *authServiceLocal) Login(ctx *request.Context, req *LoginRequest) (*LoginResponse, error) {
	// Simple mock: any username/password combination works
	token := fmt.Sprintf("token_%s_%d", req.Username, s.tokenExpiry)
	s.tokens[token] = req.Username

	return &LoginResponse{
		Token:     token,
		ExpiresIn: s.tokenExpiry,
	}, nil
}

func (s *authServiceLocal) Logout(ctx *request.Context, req *LogoutRequest) (*LogoutResponse, error) {
	token := req.Token
	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	if _, ok := s.tokens[token]; ok {
		delete(s.tokens, token)
		return &LogoutResponse{Success: true, Message: "Successfully logged out"}, nil
	}
	return &LogoutResponse{Success: false, Message: "Token not found"}, nil
}

func (s *authServiceLocal) ValidateToken(ctx *request.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	userID, ok := s.tokens[req.Token]
	if !ok {
		return &ValidateTokenResponse{Valid: false}, nil
	}

	return &ValidateTokenResponse{
		Valid:  true,
		UserID: userID,
	}, nil
}

var _ AuthService = (*authServiceLocal)(nil)

// ==============================================================================
// Remote Implementation
// ==============================================================================

type authServiceRemote struct {
	client *api_client.RemoteService
}

// Pure delegation - no logic, just bridge to remote
func (s *authServiceRemote) Login(ctx *request.Context, req *LoginRequest) (*LoginResponse, error) {
	return api_client.CallRemoteService[*LoginResponse](s.client, "Login", ctx, req)
}

func (s *authServiceRemote) Logout(ctx *request.Context, req *LogoutRequest) (*LogoutResponse, error) {
	return api_client.CallRemoteService[*LogoutResponse](s.client, "Logout", ctx, req)
}

func (s *authServiceRemote) ValidateToken(ctx *request.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
	return api_client.CallRemoteService[*ValidateTokenResponse](s.client, "ValidateToken", ctx, req)
}

var _ AuthService = (*authServiceRemote)(nil)

// ==============================================================================
// Service Factories
// ==============================================================================

// CreateAuthServiceLocal creates local implementation
func CreateAuthServiceLocal(cfg map[string]any) any {
	jwtSecret := utils.GetValueFromMap(cfg, "jwt_secret", "default-secret")
	tokenExpiry := utils.GetValueFromMap(cfg, "token_expiry", 3600)

	fmt.Printf("[auth-service] Creating LOCAL with jwt_secret: %s, token_expiry: %d\n", jwtSecret, tokenExpiry)

	return &authServiceLocal{
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
		userService: service.MustLazyLoadFromConfig[UserService](cfg, "user-service"),
		tokens:      make(map[string]string),
	}
}

// CreateAuthServiceRemote creates HTTP client wrapper
func CreateAuthServiceRemote(cfg map[string]any) any {
	routerName := utils.GetValueFromMap(cfg, "router", "auth-service")
	pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/auth")

	fmt.Printf("[auth-service] Creating REMOTE client for router: %s, prefix: %s\n", routerName, pathPrefix)

	return &authServiceRemote{
		client: lokstra_registry.GetRemoteService(cfg),
	}
}

func RegisterAuthService() {
	lokstra_registry.RegisterServiceTypeLocalAndRemote("auth_service",
		CreateAuthServiceLocal, CreateAuthServiceRemote)
}
