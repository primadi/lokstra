package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/models"
)

// UserRepository interface defines the contract for user data access
type UserRepository interface {
	// Create creates a new user and returns the created user
	Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// Update updates a user and returns the updated user
	Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error)

	// Delete soft deletes a user
	Delete(ctx context.Context, id int64) error

	// HardDelete permanently deletes a user
	HardDelete(ctx context.Context, id int64) error

	// List retrieves users with pagination and optional search
	List(ctx context.Context, req *models.ListUsersRequest) ([]*models.User, int64, error)

	// ExistsByEmail checks if a user exists with the given email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByEmailExcludingID checks if a user exists with the given email, excluding a specific ID
	ExistsByEmailExcludingID(ctx context.Context, email string, excludeID int64) (bool, error)
}

// postgresUserRepository implements UserRepository for PostgreSQL
type postgresUserRepository struct {
	pool      *pgxpool.Pool
	tableName string
}

// NewPostgresUserRepository creates a new PostgreSQL user repository
func NewPostgresUserRepository(pool *pgxpool.Pool, tableName string) UserRepository {
	if tableName == "" {
		tableName = "users"
	}
	return &postgresUserRepository{
		pool:      pool,
		tableName: tableName,
	}
}

// Create creates a new user
func (r *postgresUserRepository) Create(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	query := fmt.Sprintf(`
		INSERT INTO %s (name, email, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, email, created_at, updated_at, deleted_at`,
		r.tableName)

	now := time.Now()
	var user models.User

	err := r.pool.QueryRow(ctx, query, req.Name, req.Email, now, now).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("user with email '%s' already exists", req.Email)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *postgresUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	query := fmt.Sprintf(`
		SELECT id, name, email, created_at, updated_at, deleted_at 
		FROM %s 
		WHERE id = $1 AND deleted_at IS NULL`,
		r.tableName)

	var user models.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := fmt.Sprintf(`
		SELECT id, name, email, created_at, updated_at, deleted_at 
		FROM %s 
		WHERE email = $1 AND deleted_at IS NULL`,
		r.tableName)

	var user models.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with email '%s' not found", email)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *postgresUserRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	// Build dynamic query based on provided fields
	var setParts []string
	var args []any
	argIndex := 1

	if req.Name != nil {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.Email != nil {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, *req.Email)
		argIndex++
	}

	if len(setParts) == 0 {
		return r.GetByID(ctx, id) // No updates to perform, return current user
	}

	// Add updated_at
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add ID for WHERE clause
	args = append(args, id)

	query := fmt.Sprintf(`
		UPDATE %s 
		SET %s 
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, name, email, created_at, updated_at, deleted_at`,
		r.tableName, strings.Join(setParts, ", "), argIndex)

	var user models.User
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return nil, fmt.Errorf("user with email '%s' already exists", *req.Email)
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

// Delete soft deletes a user
func (r *postgresUserRepository) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf(`
		UPDATE %s 
		SET deleted_at = $1, updated_at = $1 
		WHERE id = $2 AND deleted_at IS NULL`,
		r.tableName)

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}

	return nil
}

// HardDelete permanently deletes a user
func (r *postgresUserRepository) HardDelete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", r.tableName)

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to hard delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}

	return nil
}

// List retrieves users with pagination and optional search
func (r *postgresUserRepository) List(ctx context.Context, req *models.ListUsersRequest) ([]*models.User, int64, error) {
	// Set default values
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	offset := (req.Page - 1) * req.PageSize

	// Build WHERE clause for search
	whereClause := "WHERE deleted_at IS NULL"
	var args []any
	argIndex := 1

	if req.Search != "" {
		whereClause += fmt.Sprintf(" AND (name ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex)
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", r.tableName, whereClause)
	var totalItems int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with pagination
	args = append(args, req.PageSize, offset)
	listQuery := fmt.Sprintf(`
		SELECT id, name, email, created_at, updated_at, deleted_at 
		FROM %s %s 
		ORDER BY created_at DESC 
		LIMIT $%d OFFSET $%d`,
		r.tableName, whereClause, argIndex, argIndex+1)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %w", err)
	}

	return users, totalItems, nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *postgresUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE email = $1 AND deleted_at IS NULL", r.tableName)
	var count int
	err := r.pool.QueryRow(ctx, query, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByEmailExcludingID checks if a user exists with the given email, excluding a specific ID
func (r *postgresUserRepository) ExistsByEmailExcludingID(ctx context.Context, email string, excludeID int64) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE email = $1 AND id != $2 AND deleted_at IS NULL", r.tableName)
	var count int
	err := r.pool.QueryRow(ctx, query, email, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}
