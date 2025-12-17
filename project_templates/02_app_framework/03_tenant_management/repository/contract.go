package repository

import (
	"context"

	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/domain"
)

// ============================================================================
// CORE ENTITY STORES
// ============================================================================

// TenantStore defines the interface for tenant persistence
type TenantStore interface {
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
