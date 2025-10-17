package auth_user_repo_pg

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/common/customtype"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const SERVICE_TYPE = "auth_user_repo_pg"

// Config represents the configuration for PostgreSQL-based UserRepository service.
type Config struct {
	DbPoolServiceName string `json:"dbpool_service_name" yaml:"dbpool_service_name"` // Name of the DbPool service to use
	Schema            string `json:"schema" yaml:"schema"`                           // Database schema
	TableName         string `json:"table_name" yaml:"table_name"`                   // Table name for users
}

type userRepoPg struct {
	cfg    *Config
	dbPool *service.Cached[serviceapi.DbPool]
}

var _ auth.UserRepository = (*userRepoPg)(nil)

func (r *userRepoPg) GetUserByName(ctx context.Context, tenantID, userName string) (*auth.User, error) {
	conn, err := r.dbPool.MustGet().Acquire(ctx, r.cfg.Schema)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := fmt.Sprintf(`
		SELECT id, tenant_id, username, email, full_name, password_hash, 
		       is_active, created_at, updated_at, last_login, metadata
		FROM %s
		WHERE tenant_id = $1 AND username = $2
	`, r.cfg.TableName)

	user := &auth.User{}
	var lastLogin *customtype.DateTime

	err = conn.SelectOne(ctx, query, []any{tenantID, userName},
		&user.ID, &user.TenantID, &user.Username, &user.Email, &user.FullName,
		&user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
		&lastLogin, &user.Metadata)

	if err != nil {
		if conn.IsErrorNoRows(err) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	if lastLogin != nil {
		user.LastLogin = *lastLogin
	}

	return user, nil
}

func (r *userRepoPg) CreateUser(ctx context.Context, user *auth.User) error {
	conn, err := r.dbPool.MustGet().Acquire(ctx, r.cfg.Schema)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := fmt.Sprintf(`
		INSERT INTO %s (id, tenant_id, username, email, full_name, password_hash, 
		                is_active, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, r.cfg.TableName)

	_, err = conn.Exec(ctx, query,
		user.ID, user.TenantID, user.Username, user.Email, user.FullName,
		user.PasswordHash, user.IsActive, user.CreatedAt, user.UpdatedAt, user.Metadata)

	return err
}

func (r *userRepoPg) UpdateUser(ctx context.Context, user *auth.User) error {
	conn, err := r.dbPool.MustGet().Acquire(ctx, r.cfg.Schema)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := fmt.Sprintf(`
		UPDATE %s
		SET email = $1, full_name = $2, password_hash = $3, is_active = $4,
		    updated_at = $5, last_login = $6, metadata = $7
		WHERE tenant_id = $8 AND username = $9
	`, r.cfg.TableName)

	result, err := conn.Exec(ctx, query,
		user.Email, user.FullName, user.PasswordHash, user.IsActive,
		user.UpdatedAt, user.LastLogin, user.Metadata,
		user.TenantID, user.Username)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepoPg) DeleteUser(ctx context.Context, tenantID, userName string) error {
	conn, err := r.dbPool.MustGet().Acquire(ctx, r.cfg.Schema)
	if err != nil {
		return err
	}
	defer conn.Release()

	query := fmt.Sprintf(`
		DELETE FROM %s
		WHERE tenant_id = $1 AND username = $2
	`, r.cfg.TableName)

	result, err := conn.Exec(ctx, query, tenantID, userName)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepoPg) ListUsers(ctx context.Context, tenantID string) ([]*auth.User, error) {
	conn, err := r.dbPool.MustGet().Acquire(ctx, r.cfg.Schema)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	query := fmt.Sprintf(`
		SELECT id, tenant_id, username, email, full_name, password_hash,
		       is_active, created_at, updated_at, last_login, metadata
		FROM %s
		WHERE tenant_id = $1
		ORDER BY username
	`, r.cfg.TableName)

	result, err := conn.SelectManyWithMapper(ctx, func(row serviceapi.Row) (any, error) {
		user := &auth.User{}
		var lastLogin *customtype.DateTime

		err := row.Scan(
			&user.ID, &user.TenantID, &user.Username, &user.Email, &user.FullName,
			&user.PasswordHash, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
			&lastLogin, &user.Metadata)

		if err != nil {
			return nil, err
		}

		if lastLogin != nil {
			user.LastLogin = *lastLogin
		}

		return user, nil
	}, query, tenantID)

	if err != nil {
		return nil, err
	}

	users := result.([]*auth.User)
	return users, nil
}

func (r *userRepoPg) Shutdown() error {
	return nil
}

func Service(cfg *Config, dbPool *service.Cached[serviceapi.DbPool]) *userRepoPg {
	return &userRepoPg{
		cfg:    cfg,
		dbPool: dbPool,
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		DbPoolServiceName: utils.GetValueFromMap(params, "dbpool_service_name", "dbpool_pg"),
		Schema:            utils.GetValueFromMap(params, "schema", "public"),
		TableName:         utils.GetValueFromMap(params, "table_name", "users"),
	}

	// Get DbPool service from registry
	dbPool := service.LazyLoad[serviceapi.DbPool](cfg.DbPoolServiceName)

	return Service(cfg, dbPool)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		lokstra_registry.AllowOverride(true))
}
