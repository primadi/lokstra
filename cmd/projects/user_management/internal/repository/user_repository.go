package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/customtype"
	"github.com/primadi/lokstra/core/dsl"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

// UserRepository implements serviceapi/auth.UserRepository using DSL
type UserRepository struct {
	serviceVars *dsl.ServiceVar[UserRepositoryParams]
}

// UserRepositoryParams contains parameters for repository operations
type UserRepositoryParams struct {
	TenantID string
	UserID   string
	Username string
	Email    string
	User     *auth.User
}

// NewUserRepository creates a new user repository
func NewUserRepository(
	dbPool serviceapi.DbPool,
	logger serviceapi.Logger,
	metrics serviceapi.Metrics,
	i18n serviceapi.I18n,
) auth.UserRepository {
	return &UserRepository{
		serviceVars: dsl.NewServiceVar(
			dbPool,
			"public", // default schema
			logger,
			metrics,
			i18n,
			&UserRepositoryParams{},
			make(map[string]any),
		),
	}
}

// GetUserByName retrieves a user by their TenantID and UserName
func (r *UserRepository) GetUserByName(ctx context.Context, tenantID, userName string) (*auth.User, error) {
	// Set parameters
	r.serviceVars.Param.TenantID = tenantID
	r.serviceVars.Param.Username = userName

	// Create request context
	reqCtx := &request.Context{Context: ctx}

	// Create DSL flow
	flow := dsl.NewFlow("GetUserByName", r.serviceVars)

	// Validate input parameters
	flow.Validate(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
		params := ctx.GetParam()
		if params.TenantID == "" {
			return dsl.ErrRequiredField("tenant_id")
		}
		if params.Username == "" {
			return dsl.ErrRequiredField("username")
		}
		return nil
	}).

		// Query user by username and tenant
		QueryOneSaveAs(`
		SELECT id, tenant_id, username, email, password_hash, is_active, 
		       created_at, updated_at, last_login, metadata
		FROM users 
		WHERE tenant_id = $1 AND username = $2 AND deleted_at IS NULL
	`, "user_row", tenantID, userName).

		// Convert result to auth.User
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			userRow, exists := ctx.GetVar("user_row")
			if !exists {
				return dsl.ErrNotFound("user", userName)
			}

			rowMap, ok := userRow.(serviceapi.RowMap)
			if !ok {
				return dsl.ErrDatabaseOperation("type_conversion", fmt.Errorf("invalid row type"))
			}

			user, err := r.rowMapToUser(rowMap)
			if err != nil {
				return err
			}

			ctx.SetVar("user", user)
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		if localizedErr, ok := err.(*dsl.LocalizedError); ok {
			if localizedErr.Code == "resource.not_found" {
				return nil, sql.ErrNoRows
			}
		}
		return nil, err
	}

	// Get result
	userVar, _ := r.serviceVars.Vars["user"]
	user, _ := userVar.(*auth.User)

	return user, nil
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *auth.User) error {
	// Set parameters
	r.serviceVars.Param.User = user

	// Create request context
	reqCtx := &request.Context{Context: ctx}

	// Create DSL flow
	flow := dsl.NewFlow("CreateUser", r.serviceVars)

	// Validate input
	flow.Validate(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
		user := ctx.GetParam().User
		if user.TenantID == "" {
			return dsl.ErrRequiredField("tenant_id")
		}
		if user.Username == "" {
			return dsl.ErrRequiredField("username")
		}
		if user.Email == "" {
			return dsl.ErrRequiredField("email")
		}
		return nil
	}).

		// Check if username already exists
		ErrorIfExists(
			dsl.ErrValidationFailed("username", user.Username),
			"SELECT 1 FROM users WHERE tenant_id = $1 AND username = $2 AND deleted_at IS NULL",
			user.TenantID, user.Username,
		).

		// Check if email already exists
		ErrorIfExists(
			dsl.ErrValidationFailed("email", user.Email),
			"SELECT 1 FROM users WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL",
			user.TenantID, user.Email,
		).

		// Generate ID and timestamps
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			user := ctx.GetParam().User
			if user.ID == "" {
				user.ID = r.generateUserID()
			}
			now := customtype.DateTime{Time: time.Now()}
			user.CreatedAt = now
			user.UpdatedAt = now
			return nil
		}).

		// Begin transaction
		BeginTx().

		// Insert user
		ExecSql(`
		INSERT INTO users (id, tenant_id, username, email, password_hash, is_active, 
		                  created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, user.ID, user.TenantID, user.Username, user.Email, user.PasswordHash,
			user.IsActive, user.CreatedAt, user.UpdatedAt, user.Metadata).

		// Commit transaction
		CommitOrRollback()

	// Execute flow
	return flow.Run(reqCtx)
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user *auth.User) error {
	// Set parameters
	r.serviceVars.Param.User = user

	// Create request context
	reqCtx := &request.Context{Context: ctx}

	// Create DSL flow
	flow := dsl.NewFlow("UpdateUser", r.serviceVars)

	// Validate input
	flow.Validate(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
		user := ctx.GetParam().User
		if user.ID == "" {
			return dsl.ErrRequiredField("id")
		}
		if user.TenantID == "" {
			return dsl.ErrRequiredField("tenant_id")
		}
		return nil
	}).

		// Check if user exists
		ErrorIfNotExists(
			dsl.ErrNotFound("user", user.ID),
			"SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL",
			user.ID, user.TenantID,
		).

		// Check if new username is taken (if username is being changed)
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			user := ctx.GetParam().User
			if user.Username != "" {
				// Check if another user already has this username
				return nil // We'll handle this in a separate step
			}
			return nil
		}).

		// Update timestamp
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			user := ctx.GetParam().User
			user.UpdatedAt = customtype.DateTime{Time: time.Now()}
			return nil
		}).

		// Begin transaction
		BeginTx().

		// Update user
		ExecSqlSaveAs(`
		UPDATE users 
		SET username = $3, email = $4, password_hash = $5, is_active = $6, 
		    updated_at = $7, last_login = $8, metadata = $9
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, "rows_affected",
			user.ID, user.TenantID, user.Username, user.Email, user.PasswordHash,
			user.IsActive, user.UpdatedAt, user.LastLogin, user.Metadata,
		).

		// Verify update succeeded
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			rowsAffected, exists := ctx.GetVar("rows_affected")
			if !exists || rowsAffected.(int64) == 0 {
				return dsl.ErrNotFound("user", ctx.GetParam().User.ID)
			}
			return nil
		}).

		// Commit transaction
		CommitOrRollback()

	// Execute flow
	return flow.Run(reqCtx)
}

// DeleteUser deletes a user by their ID (soft delete)
func (r *UserRepository) DeleteUser(ctx context.Context, tenantID, userName string) error {
	// Set parameters
	r.serviceVars.Param.TenantID = tenantID
	r.serviceVars.Param.Username = userName

	// Create request context
	reqCtx := &request.Context{Context: ctx}

	// Create DSL flow
	flow := dsl.NewFlow("DeleteUser", r.serviceVars)

	// Validate input
	flow.Validate(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
		params := ctx.GetParam()
		if params.TenantID == "" {
			return dsl.ErrRequiredField("tenant_id")
		}
		if params.Username == "" {
			return dsl.ErrRequiredField("username")
		}
		return nil
	}).

		// Check if user exists
		ErrorIfNotExists(
			dsl.ErrNotFound("user", userName),
			"SELECT 1 FROM users WHERE tenant_id = $1 AND username = $2 AND deleted_at IS NULL",
			tenantID, userName,
		).

		// Begin transaction
		BeginTx().

		// Soft delete user
		ExecSqlSaveAs(`
		UPDATE users 
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND username = $2 AND deleted_at IS NULL
	`, "rows_affected", tenantID, userName).

		// Verify deletion succeeded
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			rowsAffected, exists := ctx.GetVar("rows_affected")
			if !exists || rowsAffected.(int64) == 0 {
				return dsl.ErrNotFound("user", ctx.GetParam().Username)
			}
			return nil
		}).

		// Commit transaction
		CommitOrRollback()

	// Execute flow
	return flow.Run(reqCtx)
}

// ListUsers lists all users in a tenant
func (r *UserRepository) ListUsers(ctx context.Context, tenantID string) ([]*auth.User, error) {
	// Set parameters
	r.serviceVars.Param.TenantID = tenantID

	// Create request context
	reqCtx := &request.Context{Context: ctx}

	// Create DSL flow
	flow := dsl.NewFlow("ListUsers", r.serviceVars)

	// Validate input
	flow.Validate(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
		params := ctx.GetParam()
		if params.TenantID == "" {
			return dsl.ErrRequiredField("tenant_id")
		}
		return nil
	}).

		// Query all users in tenant
		QuerySaveAs(`
		SELECT id, tenant_id, username, email, password_hash, is_active, 
		       created_at, updated_at, last_login, metadata
		FROM users 
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, "user_rows", tenantID).

		// Convert results to auth.User slice
		Do(func(ctx *dsl.FlowContext[UserRepositoryParams]) error {
			userRows, exists := ctx.GetVar("user_rows")
			if !exists {
				ctx.SetVar("users", []*auth.User{})
				return nil
			}

			rowMaps, ok := userRows.([]serviceapi.RowMap)
			if !ok {
				return dsl.ErrDatabaseOperation("type_conversion", fmt.Errorf("invalid rows type"))
			}

			users := make([]*auth.User, 0, len(rowMaps))
			for _, rowMap := range rowMaps {
				user, err := r.rowMapToUser(rowMap)
				if err != nil {
					return err
				}
				users = append(users, user)
			}

			ctx.SetVar("users", users)
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		return nil, err
	}

	// Get result
	usersVar, _ := r.serviceVars.Vars["users"]
	users, _ := usersVar.([]*auth.User)

	return users, nil
}

// Helper methods

// rowMapToUser converts a database row to auth.User
func (r *UserRepository) rowMapToUser(rowMap serviceapi.RowMap) (*auth.User, error) {
	user := &auth.User{}

	if id, ok := rowMap["id"].(string); ok {
		user.ID = id
	}

	if tenantID, ok := rowMap["tenant_id"].(string); ok {
		user.TenantID = tenantID
	}

	if username, ok := rowMap["username"].(string); ok {
		user.Username = username
	}

	if email, ok := rowMap["email"].(string); ok {
		user.Email = email
	}

	if passwordHash, ok := rowMap["password_hash"].(string); ok {
		user.PasswordHash = passwordHash
	}

	if isActive, ok := rowMap["is_active"].(bool); ok {
		user.IsActive = isActive
	}

	if createdAt, ok := rowMap["created_at"].(time.Time); ok {
		user.CreatedAt = customtype.DateTime{Time: createdAt}
	}

	if updatedAt, ok := rowMap["updated_at"].(time.Time); ok {
		user.UpdatedAt = customtype.DateTime{Time: updatedAt}
	}

	if lastLogin, ok := rowMap["last_login"].(time.Time); ok && !lastLogin.IsZero() {
		user.LastLogin = customtype.DateTime{Time: lastLogin}
	}

	if metadata, ok := rowMap["metadata"].(map[string]any); ok {
		user.Metadata = metadata
	}

	return user, nil
}

// generateUserID generates a unique user ID
func (r *UserRepository) generateUserID() string {
	hash := sha256.Sum256(fmt.Appendf(nil, "%d", time.Now().UnixNano()))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for shorter ID
}

// HashPassword hashes a password using SHA256 (in production, use bcrypt or scrypt)
func (r *UserRepository) HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

var _ auth.UserRepository = (*UserRepository)(nil)
