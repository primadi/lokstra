package mainapp

import (
	"fmt"
)

// User represents a user entity
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Role represents a role entity
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// ==================== User Handlers ====================

// HandleGetUsers returns a list of all users
func HandleGetUsers() ([]User, error) {
	// In a real application, this would fetch from a database
	users := []User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}

	return users, nil
}

type getUserParams struct {
	ID string `path:"id" validate:"required"`
}

// HandleGetUser returns a single user by ID
func HandleGetUser(p *getUserParams) (*User, error) {
	// In a real application, this would fetch from a database
	user := &User{
		ID:    p.ID,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	return user, nil
}

type createUserParams struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// HandleCreateUser creates a new user
func HandleCreateUser(p *createUserParams) (*User, error) {
	// In a real application, this would save to a database
	// For demo purposes, just set an ID
	user := &User{
		ID:    "123",
		Name:  p.Name,
		Email: p.Email,
	}

	return user, nil
}

type updateUserParams struct {
	ID    string `path:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// HandleUpdateUser updates an existing user (full update)
func HandleUpdateUser(p *updateUserParams) (*User, error) {
	// In a real application, this would update the database
	user := &User{
		ID:    p.ID,
		Name:  p.Name,
		Email: p.Email,
	}

	return user, nil
}

type patchUserParams struct {
	ID      string         `path:"id" validate:"required"`
	Updates map[string]any `json:"*"` // Wildcard: captures entire body
}

// HandlePatchUser partially updates a user
// Note: Uses json:"*" wildcard tag to capture entire request body as map[string]any
// This allows flexible partial updates without defining all possible fields
func HandlePatchUser(p *patchUserParams) (map[string]any, error) {
	// In a real application, this would partially update the database
	result := map[string]any{
		"id":      p.ID,
		"updated": p.Updates,
	}

	return result, nil
}

type deleteUserParams struct {
	ID string `path:"id" validate:"required"`
}

// HandleDeleteUser deletes a user
func HandleDeleteUser(p *deleteUserParams) (string, error) {
	// In a real application, this would delete from the database
	message := fmt.Sprintf("User with ID %s deleted successfully", p.ID)

	return message, nil
}

// ==================== Role Handlers ====================

// HandleGetRoles returns a list of all roles
func HandleGetRoles() ([]Role, error) {
	// In a real application, this would fetch from a database
	roles := []Role{
		{ID: "1", Name: "Admin", Description: "Administrator role", Permissions: []string{"read", "write", "delete"}},
		{ID: "2", Name: "User", Description: "Regular user role", Permissions: []string{"read", "write"}},
		{ID: "3", Name: "Guest", Description: "Guest role with limited access", Permissions: []string{"read"}},
	}

	return roles, nil
}

type getRoleParams struct {
	ID string `path:"id" validate:"required"`
}

// HandleGetRole returns a single role by ID
func HandleGetRole(p *getRoleParams) (*Role, error) {
	// In a real application, this would fetch from a database
	role := &Role{
		ID:          p.ID,
		Name:        "Admin",
		Description: "Administrator role",
		Permissions: []string{"read", "write", "delete"},
	}

	return role, nil
}

type createRoleParams struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Permissions []string `json:"permissions" validate:"required"`
}

// HandleCreateRole creates a new role
func HandleCreateRole(p *createRoleParams) (*Role, error) {
	// In a real application, this would save to a database
	// For demo purposes, just set an ID
	role := &Role{
		ID:          "456",
		Name:        p.Name,
		Description: p.Description,
		Permissions: p.Permissions,
	}

	return role, nil
}

type updateRoleParams struct {
	ID          string   `path:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	Permissions []string `json:"permissions" validate:"required"`
}

// HandleUpdateRole updates an existing role (full update)
func HandleUpdateRole(p *updateRoleParams) (*Role, error) {
	// In a real application, this would update the database
	role := &Role{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Permissions: p.Permissions,
	}

	return role, nil
}

type patchRoleParams struct {
	ID      string         `path:"id" validate:"required"`
	Updates map[string]any `json:"*"` // Wildcard: captures entire body
}

// HandlePatchRole partially updates a role
// Note: Uses json:"*" wildcard tag to capture entire request body as map[string]any
// This allows flexible partial updates without defining all possible fields
func HandlePatchRole(p *patchRoleParams) (map[string]any, error) {
	// In a real application, this would partially update the database
	result := map[string]any{
		"id":      p.ID,
		"updated": p.Updates,
	}

	return result, nil
}

type deleteRoleParams struct {
	ID string `path:"id" validate:"required"`
}

// HandleDeleteRole deletes a role
func HandleDeleteRole(p *deleteRoleParams) (string, error) {
	// In a real application, this would delete from the database
	message := fmt.Sprintf("Role with ID %s deleted successfully", p.ID)

	return message, nil
}

type assignRoleToUserParams struct {
	RoleID string `path:"id" validate:"required"`
	UserID string `path:"userId" validate:"required"`
}

// HandleAssignRoleToUser assigns a role to a user
// This demonstrates working with nested resources and multiple path parameters
func HandleAssignRoleToUser(p *assignRoleToUserParams) (map[string]string, error) {
	// In a real application, this would create a relationship in the database
	result := map[string]string{
		"message": fmt.Sprintf("Role %s assigned to user %s", p.RoleID, p.UserID),
		"roleId":  p.RoleID,
		"userId":  p.UserID,
	}

	return result, nil
}
