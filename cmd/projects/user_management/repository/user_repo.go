package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

type UserRepository struct {
	dbExecutor serviceapi.DbExecutor
}

// CreateUser implements auth.UserRepository.
func (u *UserRepository) CreateUser(ctx context.Context, user *auth.User) error {
	// validate user
	if err := u.validateUser(ctx, user); err != nil {
		return err
	}

	u.dbExecutor.Exec(ctx,
		`INSERT INTO users (id, tenant_id, username, 
		email, full_name, password_hash, is_active, metadata) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`, user.ID, user.TenantID, user.Username,
		user.Email, user.FullName, user.PasswordHash, user.IsActive, user.Metadata)
	return nil
}

// DeleteUser implements auth.UserRepository.
func (u *UserRepository) DeleteUser(ctx context.Context, tenantID string, userName string) error {
	res, err := u.dbExecutor.Exec(ctx, `UPDATE users SET is_active=false WHERE 
		tenant_id=$1 AND username=$2`, tenantID, userName)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

// GetUserByName implements auth.UserRepository.
func (u *UserRepository) GetUserByName(ctx context.Context, tenantID string,
	userName string) (*auth.User, error) {
	var user auth.User
	err := u.dbExecutor.QueryRow(ctx,
		`SELECT id, tenant_id, username, email, full_name, is_active, metadata
		FROM users WHERE tenant_id=$1 AND username=$2`, tenantID, userName).
		Scan(&user.ID, &user.TenantID, &user.Username,
			&user.Email, &user.FullName, &user.IsActive, &user.Metadata)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID gets a user by their ID
func (u *UserRepository) GetUserByID(ctx context.Context, tenantID string,
	userID string) (*auth.User, error) {
	var user auth.User
	err := u.dbExecutor.QueryRow(ctx,
		`SELECT id, tenant_id, username, email, full_name, is_active, metadata
		FROM users WHERE tenant_id=$1 AND id=$2`, tenantID, userID).
		Scan(&user.ID, &user.TenantID, &user.Username,
			&user.Email, &user.FullName, &user.IsActive, &user.Metadata)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ListUsers implements auth.UserRepository.
func (u *UserRepository) ListUsers(ctx context.Context, tenantID string) ([]*auth.User, error) {
	var users []*auth.User
	rows, err := u.dbExecutor.Query(ctx,
		`SELECT id, tenant_id, username, email, full_name, is_active, metadata
		FROM users WHERE tenant_id=$1 ORDER BY username`, tenantID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user auth.User
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Username,
			&user.Email, &user.FullName, &user.IsActive, &user.Metadata); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// ListUsersWithPagination returns paginated users with total count
func (u *UserRepository) ListUsersWithPagination(ctx context.Context, tenantID string, page, pageSize int, filters map[string]string) ([]*auth.User, int, error) {
	// Build dynamic query with filters
	baseQuery := `SELECT id, tenant_id, username, email, full_name, is_active, metadata FROM users WHERE tenant_id=$1`
	countQuery := `SELECT COUNT(*) FROM users WHERE tenant_id=$1`

	var whereConditions []string
	var args []any
	args = append(args, tenantID)
	argIndex := 2

	// Apply filters
	for field, value := range filters {
		switch field {
		case "username":
			whereConditions = append(whereConditions, fmt.Sprintf("username ILIKE $%d", argIndex))
			args = append(args, "%"+value+"%")
			argIndex++
		case "email":
			whereConditions = append(whereConditions, fmt.Sprintf("email ILIKE $%d", argIndex))
			args = append(args, "%"+value+"%")
			argIndex++
		case "is_active":
			if value == "true" || value == "false" {
				whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
				args = append(args, value == "true")
				argIndex++
			}
		}
	}

	// Add where conditions to queries
	if len(whereConditions) > 0 {
		whereClause := " AND " + strings.Join(whereConditions, " AND ")
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Get total count
	var total int
	err := u.dbExecutor.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Add pagination
	offset := (page - 1) * pageSize
	baseQuery += fmt.Sprintf(" ORDER BY username LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, pageSize, offset)

	// Execute query
	var users []*auth.User
	rows, err := u.dbExecutor.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var user auth.User
		if err := rows.Scan(&user.ID, &user.TenantID, &user.Username,
			&user.Email, &user.FullName, &user.IsActive, &user.Metadata); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	return users, total, nil
}

// UpdateUser implements auth.UserRepository.
func (u *UserRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	// validate user
	if err := u.validateUser(ctx, user); err != nil {
		return err
	}

	res, err := u.dbExecutor.Exec(ctx,
		`UPDATE users SET email=$1, full_name=$2, password_hash=$3, 
		is_active=$4, metadata=$5 WHERE tenant_id=$6 AND username=$7`,
		user.Email, user.FullName, user.PasswordHash, user.IsActive, user.Metadata,
		user.TenantID, user.Username)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}

func NewUserRepository(dbExecutor serviceapi.DbExecutor) *UserRepository {
	return &UserRepository{
		dbExecutor: dbExecutor,
	}
}

var _ auth.UserRepository = (*UserRepository)(nil)

func (u *UserRepository) validateUser(ctx context.Context, user *auth.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.TenantID == "" {
		return errors.New("tenant_id is required")
	}

	exists, err := u.dbExecutor.IsExists(ctx, "SELECT 1 FROM tenants WHERE id=$1",
		user.TenantID)

	if err != nil {
		return err
	}
	if !exists {
		return errors.New("tenant_id does not exist")
	}

	if user.Username == "" {
		return errors.New("username is required")
	}
	if user.Email == "" {
		return errors.New("email is required")
	}
	if user.PasswordHash == "" {
		return errors.New("password_hash is required")
	}
	return nil
}
