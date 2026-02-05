package handler

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/domain"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/repository"
)

// @Handler name="tenant-handler", prefix="${api-auth-prefix:/api/auth}/core/tenants", middlewares=["recovery", "request_logger"]
type TenantHandler struct {
	// @Inject "@repository.tenant-repository"
	TenantRepository repository.TenantRepository

	// @Inject "@repository.user-repository"
	UserRepository repository.UserRepository
}

// @Route "POST /"
func (s *TenantHandler) CreateTenant(ctx *request.Context,
	req *domain.CreateTenantRequest) (*domain.Tenant, error) {

	// Begin transaction - lazy created on first database operation
	// Auto-finalized (commit/rollback) when response is written
	ctx.BeginTransaction("db_auth")

	// Check if tenant name already exists
	existing, err := s.TenantRepository.GetByName(ctx, req.Name)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tenant with name '%s' already exists", req.Name)
	}

	// Check if tenant ID already exists
	existingByID, _ := s.TenantRepository.Get(ctx, req.ID)
	if existingByID != nil {
		return nil, fmt.Errorf("tenant with ID '%s' already exists", req.ID)
	}

	// Generate owner user ID
	ownerID := fmt.Sprintf("%s-owner-%d", req.ID, time.Now().Unix())

	// Initialize settings - use provided or defaults
	settings := &domain.TenantSettings{}
	if req.Settings != nil {
		settings = req.Settings
	}

	// Initialize metadata - use provided or empty map
	metadata := &map[string]any{}
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	// Create tenant
	tenant := &domain.Tenant{
		ID:        req.ID,
		Name:      req.Name,
		OwnerID:   ownerID,
		DBDsn:     req.DBDsn,
		DBSchema:  req.DBSchema,
		Status:    domain.TenantStatusActive,
		Settings:  settings,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create tenant (transaction auto-starts here)
	if err := s.TenantRepository.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create owner user (joins same transaction)
	ownerUser := &domain.User{
		ID:        ownerID,
		TenantID:  tenant.ID,
		Email:     req.OwnerEmail,
		Name:      req.OwnerName,
		Role:      domain.UserRoleOwner,
		Status:    domain.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.UserRepository.Create(ctx, ownerUser); err != nil {
		return nil, fmt.Errorf("failed to create owner user: %w", err)
	}

	// Transaction commits automatically on success, rollback on error
	return tenant, nil
}

// @Route "GET /{id}"
func (s *TenantHandler) GetTenant(ctx *request.Context, req *domain.GetTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.TenantRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// @Route "PUT /{id}"
func (s *TenantHandler) UpdateTenant(ctx *request.Context, req *domain.UpdateTenantRequest) (*domain.Tenant, error) {
	// Get existing tenant
	tenant, err := s.TenantRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Update fields
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.DBDsn != "" {
		tenant.DBDsn = req.DBDsn
	}
	if req.DBSchema != "" {
		tenant.DBSchema = req.DBSchema
	}
	if req.Settings != nil {
		tenant.Settings = req.Settings
	}
	if req.Metadata != nil {
		tenant.Metadata = req.Metadata
	}

	// Update timestamp
	tenant.UpdatedAt = time.Now()

	// Save to repository
	if err := s.TenantRepository.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// @Route "DELETE /{id}"
func (s *TenantHandler) DeleteTenant(ctx *request.Context, req *domain.DeleteTenantRequest) error {
	// Check if tenant exists
	exists, err := s.TenantRepository.Exists(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("tenant not found: %s", req.ID)
	}

	// Delete from repository
	if err := s.TenantRepository.Delete(ctx, req.ID); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	return nil
}

// @Route "GET /"
func (s *TenantHandler) ListTenants(ctx *request.Context, req *domain.ListTenantsRequest) ([]*domain.Tenant, error) {
	tenants, err := s.TenantRepository.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	return tenants, nil
}

// @Route "POST /{id}/activate"
func (s *TenantHandler) ActivateTenant(ctx *request.Context, req *domain.ActivateTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.TenantRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Status = domain.TenantStatusActive
	tenant.UpdatedAt = time.Now()

	if err := s.TenantRepository.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to activate tenant: %w", err)
	}

	return tenant, nil
}

// @Route "POST /{id}/suspend"
func (s *TenantHandler) SuspendTenant(ctx *request.Context, req *domain.SuspendTenantRequest) (*domain.Tenant, error) {
	tenant, err := s.TenantRepository.Get(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	tenant.Status = domain.TenantStatusSuspended
	tenant.UpdatedAt = time.Now()

	if err := s.TenantRepository.Update(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to suspend tenant: %w", err)
	}

	return tenant, nil
}
