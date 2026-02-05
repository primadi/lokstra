package repository

import (
	"context"

	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/domain"
)

// ============================================================================
// CORE ENTITY STORES
// ============================================================================

// TenantRepository defines the interface for tenant persistence
type TenantRepository interface {
	// Create creates a new tenant
	Create(ctx context.Context, tenant *domain.Tenant) error

	// Get retrieves a tenant by ID
	Get(ctx context.Context, tenantID string) (*domain.Tenant, error)

	// Update updates an existing tenant
	Update(ctx context.Context, tenant *domain.Tenant) error

	// Delete deletes a tenant
	Delete(ctx context.Context, tenantID string) error

	// List lists all tenants
	List(ctx context.Context) ([]*domain.Tenant, error)

	// GetByName retrieves a tenant by name
	GetByName(ctx context.Context, name string) (*domain.Tenant, error)

	// Exists checks if a tenant exists
	Exists(ctx context.Context, tenantID string) (bool, error)
}

// UserRepository defines the interface for user persistence
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *domain.User) error

	// Get retrieves a user by ID
	Get(ctx context.Context, userID string) (*domain.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *domain.User) error

	// Delete deletes a user
	Delete(ctx context.Context, userID string) error

	// ListByTenant lists all users for a tenant
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.User, error)

	// GetByEmail retrieves a user by email within a tenant
	GetByEmail(ctx context.Context, tenantID string, email string) (*domain.User, error)

	// Exists checks if a user exists
	Exists(ctx context.Context, userID string) (bool, error)
}
