package auth

import (
	"context"

	"github.com/primadi/lokstra/common/customtype"
)

type UserRepository interface {
	// GetUserByName retrieves a user by their TenantID, UserName.
	GetUserByName(ctx context.Context, tenantID, userName string) (*User, error)
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, user *User) error
	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, user *User) error
	// DeleteUser deletes a user by their ID.
	DeleteUser(ctx context.Context, tenantID, userName string) error
	// ListUsers lists all users in a tenant.
	ListUsers(ctx context.Context, tenantID string) ([]*User, error)
}

type User struct {
	ID           string              `json:"id"`
	TenantID     string              `json:"tenant_id"`
	Username     string              `json:"username"`
	Email        string              `json:"email"`
	FullName     string              `json:"full_name"`
	PasswordHash string              `json:"-"`
	IsActive     bool                `json:"is_active"`
	CreatedAt    customtype.DateTime `json:"created_at"`
	UpdatedAt    customtype.DateTime `json:"updated_at"`
	LastLogin    customtype.DateTime `json:"last_login,omitempty"` // optional, can be nil if never logged in
	Metadata     map[string]any      `json:"metadata,omitempty"`   // optional (role, preferences, etc)
}
