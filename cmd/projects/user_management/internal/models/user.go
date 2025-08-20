package models

import (
	"time"

	"github.com/primadi/lokstra/common/customtype"
	"github.com/primadi/lokstra/serviceapi/auth"
)

// User model extends auth.User for additional fields
type User struct {
	auth.User
	// Additional fields can be added here if needed
	DisplayName string `json:"display_name,omitempty"`
	ProfileURL  string `json:"profile_url,omitempty"`
}

// CreateUserRequest represents request to create a new user
type CreateUserRequest struct {
	TenantID    string         `json:"tenant_id" validate:"required"`
	Username    string         `json:"username" validate:"required,min=3,max=50"`
	Email       string         `json:"email" validate:"required,email"`
	Password    string         `json:"password" validate:"required,min=8"`
	IsActive    *bool          `json:"is_active,omitempty"`
	DisplayName string         `json:"display_name,omitempty"`
	ProfileURL  string         `json:"profile_url,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateUserRequest represents request to update a user
type UpdateUserRequest struct {
	Username    *string        `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email       *string        `json:"email,omitempty" validate:"omitempty,email"`
	Password    *string        `json:"password,omitempty" validate:"omitempty,min=8"`
	IsActive    *bool          `json:"is_active,omitempty"`
	DisplayName *string        `json:"display_name,omitempty"`
	ProfileURL  *string        `json:"profile_url,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// ListUsersRequest represents request to list users
type ListUsersRequest struct {
	TenantID string `json:"tenant_id" validate:"required"`
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Limit    int    `json:"limit" validate:"min=1,max=100"`
	Offset   int    `json:"offset" validate:"min=0"`
	Search   string `json:"search,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
	SortBy   string `json:"sort_by,omitempty"`  // username, email, created_at, updated_at
	SortDir  string `json:"sort_dir,omitempty"` // asc, desc
}

// ListUsersResponse represents response for list users
type ListUsersResponse struct {
	Users  []UserResponse `json:"users"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

// UserResponse represents a single user response
type UserResponse struct {
	ID        string              `json:"id"`
	Username  string              `json:"username"`
	Email     string              `json:"email"`
	IsActive  bool                `json:"is_active"`
	CreatedAt customtype.DateTime `json:"created_at"`
	UpdatedAt customtype.DateTime `json:"updated_at"`
	LastLogin customtype.DateTime `json:"last_login,omitempty"`
	Metadata  map[string]any      `json:"metadata,omitempty"`
}

// ValidationError represents validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ToAuthUser converts our User model to auth.User
func (u *User) ToAuthUser() *auth.User {
	return &u.User
}

// FromAuthUser creates our User model from auth.User
func FromAuthUser(authUser *auth.User) *User {
	return &User{
		User: *authUser,
	}
}

// SetDefaults sets default values for CreateUserRequest
func (req *CreateUserRequest) SetDefaults() {
	if req.IsActive == nil {
		active := true
		req.IsActive = &active
	}
	if req.Metadata == nil {
		req.Metadata = make(map[string]any)
	}
}

// Validate validates CreateUserRequest fields
func (req *CreateUserRequest) Validate() error {
	if req.TenantID == "" {
		return &ValidationError{Field: "tenant_id", Message: "tenant_id is required"}
	}
	if req.Username == "" {
		return &ValidationError{Field: "username", Message: "username is required"}
	}
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return &ValidationError{Field: "username", Message: "username must be between 3 and 50 characters"}
	}
	if req.Email == "" {
		return &ValidationError{Field: "email", Message: "email is required"}
	}
	if req.Password == "" {
		return &ValidationError{Field: "password", Message: "password is required"}
	}
	if len(req.Password) < 8 {
		return &ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	return nil
}

// SetDefaults sets default values for ListUsersRequest
func (req *ListUsersRequest) SetDefaults() {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
}

// PasswordChangeRequest represents request to change password
type PasswordChangeRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// UserStatsResponse represents user statistics
type UserStatsResponse struct {
	TotalUsers        int `json:"total_users"`
	ActiveUsers       int `json:"active_users"`
	InactiveUsers     int `json:"inactive_users"`
	NewUsersToday     int `json:"new_users_today"`
	NewUsersThisWeek  int `json:"new_users_this_week"`
	NewUsersThisMonth int `json:"new_users_this_month"`
}

// Helper function to create auth.User from CreateUserRequest
func (req *CreateUserRequest) ToAuthUser() *auth.User {
	now := customtype.DateTime{Time: time.Now()}

	return &auth.User{
		TenantID:  req.TenantID,
		Username:  req.Username,
		Email:     req.Email,
		IsActive:  *req.IsActive,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  req.Metadata,
	}
}
