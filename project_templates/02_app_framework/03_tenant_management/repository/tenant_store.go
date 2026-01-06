package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/domain"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-tenant-store"
type PostgresTenantStore struct {
	// @Inject "db_auth"
	dbPool serviceapi.DbPool
}

var _ TenantStore = (*PostgresTenantStore)(nil)

func (s *PostgresTenantStore) Create(ctx context.Context, tenant *domain.Tenant) error {
	metadata, _ := json.Marshal(tenant.Metadata)
	settings, _ := json.Marshal(tenant.Settings)

	query := `
		INSERT INTO tenants (
			id, name, domain, db_dsn, db_schema, status,
			settings, metadata, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.dbPool.Exec(ctx, query,
		tenant.ID, tenant.Name, tenant.Domain, tenant.DBDsn, tenant.DBSchema, tenant.Status,
		settings, metadata, tenant.CreatedAt, tenant.UpdatedAt, tenant.DeletedAt,
	)
	return err
}

func (s *PostgresTenantStore) Get(ctx context.Context, tenantID string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		       settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE id = $1
	`
	tenant := &domain.Tenant{}
	var metadata, settings []byte

	err := s.dbPool.QueryRow(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
		&settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		tenant.Metadata = &m
	}
	if len(settings) > 0 {
		tenant.Settings = &domain.TenantSettings{}
		json.Unmarshal(settings, tenant.Settings)
	}

	return tenant, nil
}

func (s *PostgresTenantStore) Update(ctx context.Context, tenant *domain.Tenant) error {
	metadata, _ := json.Marshal(tenant.Metadata)
	settings, _ := json.Marshal(tenant.Settings)

	query := `
		UPDATE tenants
		SET name = $1, domain = $2, db_dsn = $3, db_schema = $4, status = $5,
		settings = $6, metadata = $7, updated_at = $8, deleted_at = $9
		WHERE id = $10
	`

	result, err := s.dbPool.Exec(ctx, query,
		tenant.Name, tenant.Domain, tenant.DBDsn, tenant.DBSchema, tenant.Status,
		settings, metadata, tenant.UpdatedAt, tenant.DeletedAt, tenant.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant not found: %s", tenant.ID)
	}
	return nil
}

func (s *PostgresTenantStore) Delete(ctx context.Context, tenantID string) error {
	query := `DELETE FROM tenants WHERE id = $1`

	_, err := s.dbPool.Exec(ctx, query, tenantID)
	return err
}

func (s *PostgresTenantStore) List(ctx context.Context) ([]*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		    settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanTenants(rows)
}

func (s *PostgresTenantStore) GetByName(ctx context.Context, name string) (*domain.Tenant, error) {
	query := `
		SELECT id, name, domain, db_dsn, db_schema, status,
		       settings, metadata, created_at, updated_at, deleted_at
		FROM tenants
		WHERE name = $1
	`

	tenant := &domain.Tenant{}
	var metadata, settings []byte

	err := s.dbPool.QueryRow(ctx, query, name).Scan(
		&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
		&settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
	)

	if s.dbPool.IsErrorNoRows(err) {
		return nil, fmt.Errorf("tenant not found with name: %s", name)
	}
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		var m map[string]any
		json.Unmarshal(metadata, &m)
		tenant.Metadata = &m
	}
	if len(settings) > 0 {
		tenant.Settings = &domain.TenantSettings{}
		json.Unmarshal(settings, tenant.Settings)
	}

	return tenant, nil
}

func (s *PostgresTenantStore) Exists(ctx context.Context, tenantID string) (bool, error) {
	query := `SELECT 1 FROM tenants WHERE id = $1`
	return s.dbPool.IsExists(ctx, query, tenantID)
}

func (s *PostgresTenantStore) scanTenants(rows serviceapi.Rows) ([]*domain.Tenant, error) {
	tenants := make([]*domain.Tenant, 0)

	for rows.Next() {
		tenant := &domain.Tenant{}
		var metadata, settings []byte

		err := rows.Scan(
			&tenant.ID, &tenant.Name, &tenant.Domain, &tenant.DBDsn, &tenant.DBSchema, &tenant.Status,
			&settings, &metadata, &tenant.CreatedAt, &tenant.UpdatedAt, &tenant.DeletedAt,
		)
		if err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			var m map[string]any
			json.Unmarshal(metadata, &m)
			tenant.Metadata = &m
		}
		if len(settings) > 0 {
			tenant.Settings = &domain.TenantSettings{}
			json.Unmarshal(settings, tenant.Settings)
		}

		tenants = append(tenants, tenant)
	}

	return tenants, rows.Err()
}
