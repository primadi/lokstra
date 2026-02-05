package auth

import (
	"github.com/primadi/lokstra/core/request"
)

// @Handler name="auth-handler", prefix="/api/v1/auth"
type AuthHandler struct {
	// @Inject "auth-service"
	authService *AuthService
}

// @Route "POST /register", middlewares=["auth", "tenant_admin"]
// Register creates a new user within a tenant (admin only)
func (h *AuthHandler) Register(ctx *request.Context, params *RegisterRequest) (*RegisterResponse, error) {
	return h.authService.Register(ctx, params.TenantID, params)
}

// @Route "POST /login"
// Login authenticates a user and returns tokens
func (h *AuthHandler) Login(ctx *request.Context, params *LoginRequest) (*LoginResponse, error) {
	return h.authService.Login(ctx, params)
}

// @Route "POST /logout", middlewares=["auth"]
// Logout invalidates the current session
func (h *AuthHandler) Logout(ctx *request.Context, params *LogoutRequest) (*MessageResponse, error) {
	return h.authService.Logout(ctx, params)
}

// @Route "POST /refresh"
// Refresh exchanges a refresh token for new tokens
func (h *AuthHandler) Refresh(ctx *request.Context, params *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	return h.authService.RefreshToken(ctx, params)
}

// @Route "POST /validate", middlewares=["auth"]
// Validate checks if a token is valid (internal use)
func (h *AuthHandler) Validate(ctx *request.Context) (*ValidateTokenResponse, error) {
	return h.authService.ValidateToken(ctx)
}

// @Route "GET /me", middlewares=["auth"]
// Me returns the current authenticated user
func (h *AuthHandler) Me(ctx *request.Context) (*MeResponse, error) {
	return h.authService.GetCurrentUser(ctx)
}

// @Route "POST /password/forgot"
// ForgotPassword sends a password reset email
func (h *AuthHandler) ForgotPassword(ctx *request.Context, params *ForgotPasswordRequest) (*MessageResponse, error) {
	return h.authService.ForgotPassword(ctx, params)
}

// @Route "POST /password/reset"
// ResetPassword resets a password using a reset token
func (h *AuthHandler) ResetPassword(ctx *request.Context, params *ResetPasswordRequest) (*MessageResponse, error) {
	return h.authService.ResetPassword(ctx, params)
}

// @Route "POST /password/change", middlewares=["auth"]
// ChangePassword changes the current user's password
func (h *AuthHandler) ChangePassword(ctx *request.Context, params *ChangePasswordRequest) (*MessageResponse, error) {
	return h.authService.ChangePassword(ctx, params)
}

// @Route "POST /switch-tenant", middlewares=["auth", "super_admin"]
// SwitchTenant changes tenant context for super admins
func (h *AuthHandler) SwitchTenant(ctx *request.Context, params *SwitchTenantRequest) (*SwitchTenantResponse, error) {
	return h.authService.SwitchTenant(ctx, params)
}
