package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/domain"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_tenant_management/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// @Service "postgres-user-repository"
type PostgresUserRepository struct {
	// @Inject "db_main"
	dbPool serviceapi.DbPool
}

var _ repository.UserRepository = (*PostgresUserRepository)(nil)

func (s *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, tenant_id, email, name, role, status,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.dbPool.Exec(ctx, query,
		user.ID, user.TenantID, user.Email, user.Name, user.Role, user.Status,
		user.CreatedAt, user.UpdatedAt, user.DeletedAt,
	)
	return err
}

func (s *PostgresUserRepository) Get(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, name, role, status,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
	`
	user := &domain.User{}

	err := s.dbPool.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Name, &user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, role = $3, status = $4,
		    updated_at = $5, deleted_at = $6
		WHERE id = $7
	`

	result, err := s.dbPool.Exec(ctx, query,
		user.Email, user.Name, user.Role, user.Status,
		user.UpdatedAt, user.DeletedAt, user.ID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	return nil
}

func (s *PostgresUserRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := s.dbPool.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found: %s", userID)
	}
	return nil
}

func (s *PostgresUserRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, name, role, status,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.dbPool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.Name, &user.Role, &user.Status,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *PostgresUserRepository) GetByEmail(ctx context.Context, tenantID string, email string) (*domain.User, error) {
	query := `
		SELECT id, tenant_id, email, name, role, status,
		       created_at, updated_at, deleted_at
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`
	user := &domain.User{}

	err := s.dbPool.QueryRow(ctx, query, tenantID, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Name, &user.Role, &user.Status,
		&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is not an error
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *PostgresUserRepository) Exists(ctx context.Context, userID string) (bool, error) {
	query := `SELECT 1 FROM users WHERE id = $1`
	var exists int
	err := s.dbPool.QueryRow(ctx, query, userID).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
