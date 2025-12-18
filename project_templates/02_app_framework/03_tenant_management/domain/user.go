package domain

import (
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserInactive      = errors.New("user is inactive")
	ErrUserDeleted       = errors.New("user is deleted")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrInvalidEmail      = errors.New("invalid email address")
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusDeleted  UserStatus = "deleted"
)

// UserRole represents user roles within a tenant
type UserRole string

const (
	UserRoleOwner  UserRole = "owner"  // Tenant owner (billing, full control)
	UserRoleAdmin  UserRole = "admin"  // Full administrative access
	UserRoleMember UserRole = "member" // Regular member
	UserRoleGuest  UserRole = "guest"  // Limited access
)

// User represents a user within a tenant
type User struct {
	ID        string     `json:"id"`        // Unique identifier
	TenantID  string     `json:"tenant_id"` // Tenant this user belongs to
	Email     string     `json:"email"`     // Email address (unique per tenant)
	Name      string     `json:"name"`      // Full name
	Role      UserRole   `json:"role"`      // User role within tenant
	Status    UserStatus `json:"status"`    // active, inactive, deleted
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsOwner checks if user is tenant owner
func (u *User) IsOwner() bool {
	return u.Role == UserRoleOwner
}

// IsAdmin checks if user is admin or owner
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin || u.Role == UserRoleOwner
}

// Validate validates user data
func (u *User) Validate() error {
	if u.ID == "" {
		return ErrInvalidUserID
	}
	if u.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if u.Email == "" {
		return ErrInvalidEmail
	}
	if u.Name == "" {
		return errors.New("user name is required")
	}
	if u.Status == "" {
		u.Status = UserStatusActive
	}
	if u.Role == "" {
		u.Role = UserRoleMember
	}
	return nil
}

// =============================================================================
// User DTOs
// =============================================================================

// CreateUserRequest request to create a user
type CreateUserRequest struct {
	TenantID string   `json:"tenant_id" validate:"required"`
	Email    string   `json:"email" validate:"required,email"`
	Name     string   `json:"name" validate:"required"`
	Role     UserRole `json:"role,omitempty"`
}

// GetUserRequest request to get a user
type GetUserRequest struct {
	ID string `path:"id" validate:"required"`
}

// UpdateUserRequest request to update a user
type UpdateUserRequest struct {
	ID     string     `path:"id" validate:"required"`
	Name   string     `json:"name,omitempty"`
	Role   UserRole   `json:"role,omitempty"`
	Status UserStatus `json:"status,omitempty"`
}

// DeleteUserRequest request to delete a user
type DeleteUserRequest struct {
	ID string `path:"id" validate:"required"`
}

// ListUsersRequest request to list users by tenant
type ListUsersRequest struct {
	TenantID string `query:"tenant_id" validate:"required"`
}
