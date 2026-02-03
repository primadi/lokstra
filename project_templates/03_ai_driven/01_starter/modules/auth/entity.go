package auth

// User represents an authenticated user in the system
type User struct {
	ID                  string  `json:"id"`
	TenantID            string  `json:"tenant_id"`
	Email               string  `json:"email"`
	Username            string  `json:"username"`
	PasswordHash        string  `json:"-"` // Never exposed in JSON
	Role                string  `json:"role"`
	Status              string  `json:"status"` // active, inactive, suspended
	FailedLoginAttempts int     `json:"failed_login_attempts"`
	LockedUntil         *string `json:"locked_until,omitempty"`
	LastLoginAt         *string `json:"last_login_at,omitempty"`
	PasswordChangedAt   *string `json:"password_changed_at,omitempty"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
	DeletedAt           *string `json:"deleted_at,omitempty"`
}

// Session represents an active user session
type Session struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	TenantID        string  `json:"tenant_id"`
	AccessTokenHash string  `json:"-"`
	IPAddress       string  `json:"ip_address"`
	UserAgent       string  `json:"user_agent"`
	ExpiresAt       string  `json:"expires_at"`
	CreatedAt       string  `json:"created_at"`
	RevokedAt       *string `json:"revoked_at,omitempty"`
}

// RefreshToken represents a refresh token for token renewal
type RefreshToken struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	TenantID  string  `json:"tenant_id"`
	TokenHash string  `json:"-"`
	ExpiresAt string  `json:"expires_at"`
	UsedAt    *string `json:"used_at,omitempty"`
	CreatedAt string  `json:"created_at"`
}

// Role represents a user role with permissions
type Role struct {
	ID           string   `json:"id"`
	TenantID     *string  `json:"tenant_id,omitempty"` // null for global roles
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Permissions  []string `json:"permissions"`
	IsSystemRole bool     `json:"is_system_role"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	TenantID  string  `json:"tenant_id"`
	TokenHash string  `json:"-"`
	ExpiresAt string  `json:"expires_at"`
	UsedAt    *string `json:"used_at,omitempty"`
	CreatedAt string  `json:"created_at"`
}

// UserStatus enum values
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
)

// Predefined role names
const (
	RoleSuperAdmin  = "super_admin"
	RoleTenantAdmin = "tenant_admin"
	RoleManager     = "manager"
	RoleMember      = "member"
	RoleViewer      = "viewer"
)
