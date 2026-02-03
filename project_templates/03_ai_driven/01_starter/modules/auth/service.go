package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"golang.org/x/crypto/bcrypt"
)

// CustomClaims represents JWT claims for access tokens
type CustomClaims struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// @Service "auth-service"
type AuthService struct {
	// @Inject "auth-repository"
	repo *AuthRepository

	// @Inject "cfg:jwt.secret"
	jwtSecret string

	// @Inject "cfg:jwt.access_token_expiry"
	accessTokenExpiry time.Duration

	// @Inject "cfg:jwt.refresh_token_expiry"
	refreshTokenExpiry time.Duration
}

// Register creates a new user within a tenant
func (s *AuthService) Register(ctx *request.Context, tenantID string, params *RegisterRequest) (*RegisterResponse, error) {
	// Check if tenant is active
	// TODO: Inject tenant service and validate

	// Check if username is globally unique
	existingUser, err := s.repo.FindByUsername(ctx, params.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUsernameTaken
	}

	// Check if email is unique within tenant
	existingUser, err = s.repo.FindByEmailInTenant(ctx, tenantID, params.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailExistsInTenant
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), 12)
	if err != nil {
		return nil, err
	}

	// Set default role if not provided
	role := params.Role
	if role == "" {
		role = RoleMember
	}

	// Create user
	user := &User{
		TenantID:     tenantID,
		Email:        params.Email,
		Username:     params.Username,
		PasswordHash: string(passwordHash),
		Role:         role,
		Status:       UserStatusActive,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      user.Role,
		Status:    user.Status,
		CreatedAt: user.CreatedAt,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx *request.Context, params *LoginRequest) (*LoginResponse, error) {
	// Find user by email within tenant
	user, err := s.repo.FindByEmailInTenant(ctx, params.TenantID, params.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check if account is locked
	if user.LockedUntil != nil {
		lockedUntil, _ := time.Parse(time.RFC3339, *user.LockedUntil)
		if time.Now().Before(lockedUntil) {
			remaining := int(time.Until(lockedUntil).Minutes()) + 1
			return nil, NewAccountLockedError(remaining)
		}
	}

	// Check user status
	if user.Status == UserStatusSuspended {
		return nil, ErrAccountSuspended
	}
	if user.Status == UserStatusInactive {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.Password)); err != nil {
		// Increment failed login attempts
		s.repo.IncrementFailedAttempts(ctx, user.ID)
		return nil, ErrInvalidCredentials
	}

	// Reset failed attempts on successful login
	s.repo.ResetFailedAttempts(ctx, user.ID)

	// Get user permissions based on role
	permissions := s.getPermissionsForRole(user.Role)

	// Generate tokens
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &Session{
		UserID:    user.ID,
		TenantID:  user.TenantID,
		IPAddress: utils.ClientIP(ctx.R),
		UserAgent: ctx.Req.HeaderParam("User-Agent", ""),
	}
	s.repo.CreateSession(ctx, session)

	// Update last login
	s.repo.UpdateLastLogin(ctx, user.ID)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.accessTokenExpiry.Seconds()),
		User: UserResponse{
			ID:          user.ID,
			TenantID:    user.TenantID,
			Email:       user.Email,
			Username:    user.Username,
			Role:        user.Role,
			Permissions: permissions,
			Status:      user.Status,
		},
	}, nil
}

// Logout invalidates the current session
func (s *AuthService) Logout(ctx *request.Context, params *LogoutRequest) (*MessageResponse, error) {
	userID := ctx.Get("user_id").(string)

	if params.AllDevices {
		s.repo.RevokeAllSessions(ctx, userID)
	} else {
		// Revoke current session only
		token := ctx.Req.HeaderParam("Authorization", "")
		s.repo.RevokeSession(ctx, token)
	}

	return &MessageResponse{Message: "Logged out successfully"}, nil
}

// RefreshToken exchanges a refresh token for new tokens
func (s *AuthService) RefreshToken(ctx *request.Context, params *RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate refresh token
	tokenData, err := s.repo.FindRefreshToken(ctx, params.RefreshToken)
	if err != nil || tokenData == nil {
		return nil, ErrInvalidToken
	}

	// Check if token is expired
	expiresAt, _ := time.Parse(time.RFC3339, tokenData.ExpiresAt)
	if time.Now().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	// Check if token was already used (rotation)
	if tokenData.UsedAt != nil {
		return nil, ErrInvalidToken
	}

	// Get user
	user, err := s.repo.FindByID(ctx, tokenData.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	// Check user and tenant status
	if user.Status != UserStatusActive {
		return nil, ErrAccountSuspended
	}

	// Mark old refresh token as used
	s.repo.MarkRefreshTokenUsed(ctx, tokenData.ID)

	// Generate new tokens
	permissions := s.getPermissionsForRole(user.Role)
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.accessTokenExpiry.Seconds()),
	}, nil
}

// ValidateToken validates the current token
func (s *AuthService) ValidateToken(ctx *request.Context) (*ValidateTokenResponse, error) {
	// Token already validated by middleware, extract data
	userID := ctx.Get("user_id").(string)
	tenantID := ctx.Get("tenant_id").(string)
	role := ctx.Get("role").(string)
	permissions := ctx.Get("permissions").([]string)
	expiresAt := ctx.Get("expires_at").(string)

	return &ValidateTokenResponse{
		Valid:       true,
		UserID:      userID,
		TenantID:    tenantID,
		Role:        role,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
	}, nil
}

// GetCurrentUser returns the current authenticated user
func (s *AuthService) GetCurrentUser(ctx *request.Context) (*MeResponse, error) {
	userID := ctx.Get("user_id").(string)
	tenantID := ctx.Get("tenant_id").(string)

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	permissions := s.getPermissionsForRole(user.Role)

	// TODO: Get tenant info from tenant service
	tenant := TenantInfo{
		ID:     tenantID,
		Name:   "Organization",
		Status: "active",
	}

	return &MeResponse{
		ID:          user.ID,
		TenantID:    user.TenantID,
		Email:       user.Email,
		Username:    user.Username,
		Role:        user.Role,
		Permissions: permissions,
		Tenant:      tenant,
		LastLoginAt: user.LastLoginAt,
	}, nil
}

// ForgotPassword sends a password reset email
func (s *AuthService) ForgotPassword(ctx *request.Context, params *ForgotPasswordRequest) (*MessageResponse, error) {
	// Find user (don't reveal if email exists)
	user, _ := s.repo.FindByEmailInTenant(ctx, params.TenantID, params.Email)

	if user != nil {
		// Generate reset token
		token := generateSecureToken(32)
		resetToken := &PasswordResetToken{
			UserID:   user.ID,
			TenantID: user.TenantID,
		}
		s.repo.CreatePasswordResetToken(ctx, resetToken, token)

		// TODO: Send email via notification service
		_ = token // Use token in email
	}

	// Always return success (security - don't reveal if email exists)
	return &MessageResponse{
		Message: "If the email exists, a reset link has been sent",
	}, nil
}

// ResetPassword resets a password using a reset token
func (s *AuthService) ResetPassword(ctx *request.Context, params *ResetPasswordRequest) (*MessageResponse, error) {
	// Find and validate reset token
	resetToken, err := s.repo.FindPasswordResetToken(ctx, params.Token)
	if err != nil || resetToken == nil {
		return nil, ErrInvalidResetToken
	}

	// Check if token is expired
	expiresAt, _ := time.Parse(time.RFC3339, resetToken.ExpiresAt)
	if time.Now().After(expiresAt) {
		return nil, ErrInvalidResetToken
	}

	// Check if token was already used
	if resetToken.UsedAt != nil {
		return nil, ErrInvalidResetToken
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.Password), 12)
	if err != nil {
		return nil, err
	}

	// Update password
	s.repo.UpdatePassword(ctx, resetToken.UserID, string(passwordHash))

	// Mark token as used
	s.repo.MarkResetTokenUsed(ctx, resetToken.ID)

	// Revoke all sessions
	s.repo.RevokeAllSessions(ctx, resetToken.UserID)

	return &MessageResponse{Message: "Password reset successfully"}, nil
}

// ChangePassword changes the current user's password
func (s *AuthService) ChangePassword(ctx *request.Context, params *ChangePasswordRequest) (*MessageResponse, error) {
	userID := ctx.Get("user_id").(string)

	// Get current user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(params.CurrentPassword)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(params.NewPassword), 12)
	if err != nil {
		return nil, err
	}

	// Update password
	s.repo.UpdatePassword(ctx, userID, string(passwordHash))

	// Revoke all other sessions
	s.repo.RevokeAllSessionsExceptCurrent(ctx, userID, ctx.Req.HeaderParam("Authorization", ""))

	return &MessageResponse{Message: "Password changed successfully"}, nil
}

// SwitchTenant switches tenant context for super admins
func (s *AuthService) SwitchTenant(ctx *request.Context, params *SwitchTenantRequest) (*SwitchTenantResponse, error) {
	userID := ctx.Get("user_id").(string)
	role := ctx.Get("role").(string)

	// Only super admin can switch tenants
	if role != RoleSuperAdmin {
		return nil, ErrInsufficientPermissions
	}

	// Get user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, ErrInvalidToken
	}

	// TODO: Validate tenant exists and is active via tenant service

	// Generate new access token with new tenant context
	user.TenantID = params.TenantID
	permissions := s.getPermissionsForRole(user.Role)
	accessToken, err := s.generateAccessToken(user, permissions)
	if err != nil {
		return nil, err
	}

	// TODO: Get actual tenant info
	tenant := TenantInfo{
		ID:     params.TenantID,
		Name:   "Switched Tenant",
		Status: "active",
	}

	return &SwitchTenantResponse{
		AccessToken: accessToken,
		Tenant:      tenant,
	}, nil
}

// Helper functions

func (s *AuthService) generateAccessToken(user *User, permissions []string) (string, error) {
	expiresAt := jwt.NewNumericDate(time.Now().Add(s.accessTokenExpiry))

	claims := &CustomClaims{
		UserID:      user.ID,
		TenantID:    user.TenantID,
		Username:    user.Username,
		Email:       user.Email,
		Role:        user.Role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: expiresAt,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth-service",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) generateRefreshToken(ctx context.Context, user *User) (string, error) {
	token := generateSecureToken(32)

	refreshToken := &RefreshToken{
		UserID:   user.ID,
		TenantID: user.TenantID,
	}

	s.repo.CreateRefreshToken(ctx, refreshToken, token)

	return token, nil
}

func (s *AuthService) getPermissionsForRole(role string) []string {
	switch role {
	case RoleSuperAdmin:
		return []string{"*:*"}
	case RoleTenantAdmin:
		return []string{"tenant:*"}
	case RoleManager:
		return []string{"*:read", "team:*"}
	case RoleMember:
		return []string{"*:read", "own:*"}
	case RoleViewer:
		return []string{"*:read"}
	default:
		return []string{}
	}
}

func generateSecureToken(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
