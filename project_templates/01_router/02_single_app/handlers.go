package main

import (
	"fmt"
	"time"
)

// ==================== Health Check Handlers ====================

// HealthStatus represents the health check response
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// handleHealth returns the health status of the application
func handleHealth() (*HealthStatus, error) {
	return &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   "1.0.0",
	}, nil
}

// ReadyStatus represents the readiness check response
type ReadyStatus struct {
	Ready     bool              `json:"ready"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// handleReady checks if the application is ready to serve requests
func handleReady() (*ReadyStatus, error) {
	// In a real application, check database, cache, external services, etc.
	return &ReadyStatus{
		Ready:     true,
		Timestamp: time.Now(),
		Services: map[string]string{
			"database": "ok",
			"cache":    "ok",
		},
	}, nil
}

// ==================== Domain Models ====================

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

// handleGetUsers returns a list of all users
func handleGetUsers() ([]User, error) {
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

// handleGetUser returns a single user by ID
func handleGetUser(p *getUserParams) (*User, error) {
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

// handleCreateUser creates a new user
func handleCreateUser(p *createUserParams) (*User, error) {
	// Parse request body
	var user User

	// In a real application, this would save to a database
	// For demo purposes, just set an ID
	user.ID = "123"
	user.Name = p.Name
	user.Email = p.Email

	return &user, nil
}

type updateUserParams struct {
	ID    string `path:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// handleUpdateUser updates an existing user (full update)
func handleUpdateUser(p *updateUserParams) (*User, error) {
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

// handlePatchUser partially updates a user
// Note: Uses json:"*" wildcard tag to capture entire request body as map[string]any
// This allows flexible partial updates without defining all possible fields
func handlePatchUser(p *patchUserParams) (map[string]any, error) {
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

// handleDeleteUser deletes a user
func handleDeleteUser(p *deleteUserParams) (string, error) {
	// In a real application, this would delete from the database
	message := fmt.Sprintf("User with ID %s deleted successfully", p.ID)

	return message, nil
}

// ==================== Role Handlers ====================

// handleGetRoles returns a list of all roles
func handleGetRoles() ([]Role, error) {
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

// handleGetRole returns a single role by ID
func handleGetRole(p *getRoleParams) (*Role, error) {
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

// handleCreateRole creates a new role
func handleCreateRole(p *createRoleParams) (*Role, error) {
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

// handleUpdateRole updates an existing role (full update)
func handleUpdateRole(p *updateRoleParams) (*Role, error) {
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

// handlePatchRole partially updates a role
// Note: Uses json:"*" wildcard tag to capture entire request body as map[string]any
// This allows flexible partial updates without defining all possible fields
func handlePatchRole(p *patchRoleParams) (map[string]any, error) {
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

// handleDeleteRole deletes a role
func handleDeleteRole(p *deleteRoleParams) (string, error) {
	// In a real application, this would delete from the database
	message := fmt.Sprintf("Role with ID %s deleted successfully", p.ID)

	return message, nil
}

type assignRoleToUserParams struct {
	RoleID string `path:"id" validate:"required"`
	UserID string `path:"userId" validate:"required"`
}

// handleAssignRoleToUser assigns a role to a user
// This demonstrates working with nested resources and multiple path parameters
func handleAssignRoleToUser(p *assignRoleToUserParams) (map[string]string, error) {
	// In a real application, this would create a relationship in the database
	result := map[string]string{
		"message": fmt.Sprintf("Role %s assigned to user %s", p.RoleID, p.UserID),
		"roleId":  p.RoleID,
		"userId":  p.UserID,
	}

	return result, nil
}
