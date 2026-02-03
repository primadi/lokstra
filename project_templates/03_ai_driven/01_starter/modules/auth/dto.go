package auth

// =====================================
// Request DTOs
// =====================================

// RegisterRequest is the DTO for user registration
type RegisterRequest struct {
	TenantID string `header:"X-Tenant-ID" validate:"required,uuid"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum_underscore"`
	Password string `json:"password" validate:"required,min=8,password_complexity"`
	Role     string `json:"role" validate:"omitempty,oneof=tenant_admin manager member viewer"`
}

// LoginRequest is the DTO for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenant_id" validate:"required,uuid"`
}

// RefreshTokenRequest is the DTO for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest is the DTO for logout
type LogoutRequest struct {
	AllDevices bool `json:"all_devices"`
}

// ForgotPasswordRequest is the DTO for password reset request
type ForgotPasswordRequest struct {
	Email    string `json:"email" validate:"required,email"`
	TenantID string `json:"tenant_id" validate:"required,uuid"`
}

// ResetPasswordRequest is the DTO for password reset
type ResetPasswordRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=8,password_complexity"`
}

// ChangePasswordRequest is the DTO for password change
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,password_complexity"`
}

// SwitchTenantRequest is the DTO for tenant switching (super admin only)
type SwitchTenantRequest struct {
	TenantID string `json:"tenant_id" validate:"required,uuid"`
}

// =====================================
// Response DTOs
// =====================================

// UserResponse is the DTO for user data in responses
type UserResponse struct {
	ID          string   `json:"id"`
	TenantID    string   `json:"tenant_id"`
	Email       string   `json:"email"`
	Username    string   `json:"username"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
	Status      string   `json:"status"`
	LastLoginAt *string  `json:"last_login_at,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

// RegisterResponse is the DTO for registration response
type RegisterResponse struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// LoginResponse is the DTO for login response
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         UserResponse `json:"user"`
}

// RefreshTokenResponse is the DTO for token refresh response
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// ValidateTokenResponse is the DTO for token validation response
type ValidateTokenResponse struct {
	Valid       bool     `json:"valid"`
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	ExpiresAt   string   `json:"expires_at"`
}

// TenantInfo is the DTO for tenant information
type TenantInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Plan   string `json:"plan,omitempty"`
}

// MeResponse is the DTO for current user response
type MeResponse struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	Email       string     `json:"email"`
	Username    string     `json:"username"`
	Role        string     `json:"role"`
	Permissions []string   `json:"permissions"`
	Tenant      TenantInfo `json:"tenant"`
	LastLoginAt *string    `json:"last_login_at,omitempty"`
}

// SwitchTenantResponse is the DTO for tenant switch response
type SwitchTenantResponse struct {
	AccessToken string     `json:"access_token"`
	Tenant      TenantInfo `json:"tenant"`
}

// MessageResponse is the DTO for simple message responses
type MessageResponse struct {
	Message string `json:"message"`
}
