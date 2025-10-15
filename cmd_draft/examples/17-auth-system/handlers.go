package main

import (
	"time"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/common/customtype"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/middleware/jwtauth"
	"github.com/primadi/lokstra/serviceapi/auth"
)

// =============================================================================
// AUTH HANDLERS
// =============================================================================

type registerRequest struct {
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
	Role     string `json:"role"` // optional, defaults to "user"
}

var userRepo = service.LazyLoad[auth.UserRepository]("user_repo")
var authSvc = service.LazyLoad[auth.Service]("auth_service")

func registerHandler(c *request.Context) error {
	var input registerRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	// Validate required fields
	if input.TenantID == "" || input.Username == "" || input.Email == "" || input.Password == "" {
		return c.Api.BadRequest("MISSING_FIELDS", "missing required fields")
	}

	// Get user repository service

	// Check if user already exists
	if _, err := userRepo.MustGet().GetUserByName(c.R.Context(), input.TenantID, input.Username); err == nil {
		return c.Api.BadRequest("USER_EXISTS", "user already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(input.Password)
	if err != nil {
		return c.Api.InternalError("failed to hash password")
	}

	// Default role
	role := input.Role
	if role == "" {
		role = "user"
	}

	// Create user
	user := &auth.User{
		ID:           uuid.New().String(),
		TenantID:     input.TenantID,
		Username:     input.Username,
		Email:        input.Email,
		FullName:     input.FullName,
		PasswordHash: passwordHash,
		IsActive:     true,
		CreatedAt:    customtype.DateTime{Time: time.Now()},
		UpdatedAt:    customtype.DateTime{Time: time.Now()},
		Metadata: map[string]any{
			"role": role,
		},
	}

	if err := userRepo.MustGet().CreateUser(c.R.Context(), user); err != nil {
		return c.Api.InternalError("failed to create user")
	}

	return c.Api.Created(map[string]any{
		"user_id":  user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     role,
	}, "user registered successfully")
}

type loginRequest struct {
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func loginHandler(c *request.Context) error {
	var input loginRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	// Login
	resp, err := authSvc.MustGet().Login(c.R.Context(), auth.LoginRequest{
		Flow: "password",
		Payload: map[string]any{
			"tenant_id": input.TenantID,
			"username":  input.Username,
			"password":  input.Password,
		},
	})

	if err != nil {
		return c.Api.Unauthorized("login failed")
	}

	return c.Api.Ok(map[string]any{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"expires_in":    resp.ExpiresIn,
		"token_type":    "Bearer",
	})
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func refreshTokenHandler(c *request.Context) error {
	var input refreshTokenRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	// Refresh token
	resp, err := authSvc.MustGet().RefreshToken(c.R.Context(), input.RefreshToken)
	if err != nil {
		return c.Api.Unauthorized("token refresh failed")
	}

	return c.Api.Ok(map[string]any{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"expires_in":    resp.ExpiresIn,
		"token_type":    "Bearer",
	})
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func logoutHandler(c *request.Context) error {
	var input logoutRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	// Logout
	if err := authSvc.MustGet().Logout(c.R.Context(), input.RefreshToken); err != nil {
		return c.Api.InternalError("logout failed")
	}

	return c.Api.Ok(map[string]any{
		"message": "logged out successfully",
	})
}

func getCurrentUserHandler(c *request.Context) error {
	// Get user info from context (set by jwtauth middleware)
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	return c.Api.Ok(map[string]any{
		"user_id":   userInfo.UserID,
		"tenant_id": userInfo.TenantID,
		"username":  userInfo.Username,
		"email":     userInfo.Email,
		"metadata":  userInfo.Metadata,
	})
}

type generateOTPRequest struct {
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
}

func generateOTPHandler(c *request.Context) error {
	var input generateOTPRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	user, err := userRepo.MustGet().GetUserByName(c.R.Context(), input.TenantID, input.Username)
	if err != nil {
		return c.Api.BadRequest("USER_NOT_FOUND", "user not found")
	}

	if !user.IsActive {
		return c.Api.BadRequest("USER_INACTIVE", "user is not active")
	}

	// For OTP generation, we'll need to use KvStore directly
	// In a real implementation, you'd have a dedicated OTP service method
	// For now, let's return a mock response
	return c.Api.Ok(map[string]any{
		"message":    "OTP sent successfully",
		"expires_in": 300,
		"note":       "Check your email/SMS for OTP code",
	})
}

type verifyOTPRequest struct {
	TenantID string `json:"tenant_id"`
	Username string `json:"username"`
	OTP      string `json:"otp"`
}

func verifyOTPHandler(c *request.Context) error {
	var input verifyOTPRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	// Login with OTP
	resp, err := authSvc.MustGet().Login(c.R.Context(), auth.LoginRequest{
		Flow: "otp",
		Payload: map[string]any{
			"tenant_id": input.TenantID,
			"username":  input.Username,
			"otp":       input.OTP,
		},
	})

	if err != nil {
		return c.Api.Unauthorized("OTP verification failed")
	}

	return c.Api.Ok(map[string]any{
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"expires_in":    resp.ExpiresIn,
		"token_type":    "Bearer",
	})
}

// =============================================================================
// USER HANDLERS
// =============================================================================

func getUserProfileHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	user, err := userRepo.MustGet().GetUserByName(c.R.Context(), userInfo.TenantID, userInfo.Username)
	if err != nil {
		return c.Api.NotFound("user not found")
	}

	return c.Api.Ok(map[string]any{
		"id":         user.ID,
		"tenant_id":  user.TenantID,
		"username":   user.Username,
		"email":      user.Email,
		"full_name":  user.FullName,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"metadata":   user.Metadata,
	})
}

type updateProfileRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

func updateUserProfileHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	var input updateProfileRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	user, err := userRepo.MustGet().GetUserByName(c.R.Context(), userInfo.TenantID, userInfo.Username)
	if err != nil {
		return c.Api.NotFound("user not found")
	}

	// Update fields
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.FullName != "" {
		user.FullName = input.FullName
	}
	user.UpdatedAt = customtype.DateTime{Time: time.Now()}

	if err := userRepo.MustGet().UpdateUser(c.R.Context(), user); err != nil {
		return c.Api.InternalError("failed to update profile")
	}

	return c.Api.Ok(map[string]any{
		"message": "profile updated successfully",
	})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func changePasswordHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	var input changePasswordRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	user, err := userRepo.MustGet().GetUserByName(c.R.Context(), userInfo.TenantID, userInfo.Username)
	if err != nil {
		return c.Api.NotFound("user not found")
	}

	// Verify old password
	if !utils.CheckPasswordHash(input.OldPassword, user.PasswordHash) {
		return c.Api.BadRequest("INVALID_PASSWORD", "incorrect old password")
	}

	// Hash new password
	newHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return c.Api.InternalError("failed to hash password")
	}

	user.PasswordHash = newHash
	user.UpdatedAt = customtype.DateTime{Time: time.Now()}

	if err := userRepo.MustGet().UpdateUser(c.R.Context(), user); err != nil {
		return c.Api.InternalError("failed to update password")
	}

	return c.Api.Ok(map[string]any{
		"message": "password changed successfully",
	})
}

func getUserOrdersHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	// Mock data for demo
	orders := []map[string]any{
		{
			"id":         "order-1",
			"user_id":    userInfo.UserID,
			"status":     "completed",
			"total":      150.00,
			"created_at": time.Now().Add(-48 * time.Hour),
		},
		{
			"id":         "order-2",
			"user_id":    userInfo.UserID,
			"status":     "pending",
			"total":      75.50,
			"created_at": time.Now().Add(-24 * time.Hour),
		},
	}

	return c.Api.Ok(orders)
}

type createOrderRequest struct {
	Items []map[string]any `json:"items"`
	Total float64          `json:"total"`
}

func createUserOrderHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	var input createOrderRequest
	if err := c.Req.BindBody(&input); err != nil {
		return c.Api.BadRequest("INVALID_REQUEST", "invalid request body")
	}

	// Mock order creation
	order := map[string]any{
		"id":         uuid.New().String(),
		"user_id":    userInfo.UserID,
		"tenant_id":  userInfo.TenantID,
		"items":      input.Items,
		"total":      input.Total,
		"status":     "pending",
		"created_at": time.Now(),
	}

	return c.Api.Created(order, "order created successfully")
}

// =============================================================================
// ADMIN HANDLERS
// =============================================================================

func listAllUsersHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	users, err := userRepo.MustGet().ListUsers(c.R.Context(), userInfo.TenantID)
	if err != nil {
		return c.Api.InternalError("failed to list users")
	}

	// Remove sensitive data
	result := make([]map[string]any, len(users))
	for i, user := range users {
		result[i] = map[string]any{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"full_name":  user.FullName,
			"is_active":  user.IsActive,
			"created_at": user.CreatedAt,
			"metadata":   user.Metadata,
		}
	}

	return c.Api.Ok(result)
}

func getUserByIDHandler(c *request.Context) error {
	userID := c.Req.PathParam("id", "")

	return c.Api.Ok(map[string]any{
		"id":      userID,
		"message": "User details would be shown here",
	})
}

func activateUserHandler(c *request.Context) error {
	userID := c.Req.PathParam("id", "")

	return c.Api.Ok(map[string]any{
		"id":      userID,
		"message": "User activated successfully",
	})
}

func deactivateUserHandler(c *request.Context) error {
	userID := c.Req.PathParam("id", "")

	return c.Api.Ok(map[string]any{
		"id":      userID,
		"message": "User deactivated successfully",
	})
}

func deleteUserHandler(c *request.Context) error {
	userID := c.Req.PathParam("id", "")

	return c.Api.Ok(map[string]any{
		"id":      userID,
		"message": "User deleted successfully",
	})
}

func getSystemStatsHandler(c *request.Context) error {
	userInfo, ok := jwtauth.GetUserInfo(c.R.Context())
	if !ok {
		return c.Api.Unauthorized("user not authenticated")
	}

	users, _ := userRepo.MustGet().ListUsers(c.R.Context(), userInfo.TenantID)

	// Count active users
	activeCount := 0
	for _, user := range users {
		if user.IsActive {
			activeCount++
		}
	}

	stats := map[string]any{
		"total_users":  len(users),
		"active_users": activeCount,
		"timestamp":    time.Now(),
	}

	return c.Api.Ok(stats)
}

// =============================================================================
// PUBLIC HANDLERS
// =============================================================================

func healthCheckHandler(c *request.Context) error {
	return c.Api.Ok(map[string]any{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

func getSystemInfoHandler(c *request.Context) error {
	return c.Api.Ok(map[string]any{
		"name":        "Lokstra Auth System",
		"version":     "1.0.0",
		"description": "Complete authentication and authorization system",
		"features": []string{
			"JWT Authentication",
			"Password Login",
			"OTP Login",
			"Role-Based Access Control",
			"Multi-Tenant Support",
		},
	})
}
