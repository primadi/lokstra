package models

import (
	"time"
)

// User represents a user entity in the system
type User struct {
	ID        int64      `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=100"`
	Email string `json:"email" validate:"required,email,max=255"`
}

// UpdateUserRequest represents the request payload for updating a user
type UpdateUserRequest struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Email *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
}

// ListUsersRequest represents query parameters for listing users
type ListUsersRequest struct {
	Page     int    `query:"page" validate:"min=1"`
	PageSize int    `query:"page_size" validate:"min=1,max=100"`
	Search   string `query:"search"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []*User         `json:"users"`
	Pagination *PaginationMeta `json:"pagination"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// UserResponse represents a standardized user response
type UserResponse struct {
	User *User `json:"user"`
}

// IsDeleted returns true if the user is soft deleted
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// Validate validates the CreateUserRequest
func (req *CreateUserRequest) Validate() error {
	if req.Name == "" {
		return NewValidationError("name", "Name is required")
	}
	if len(req.Name) < 2 {
		return NewValidationError("name", "Name must be at least 2 characters")
	}
	if len(req.Name) > 100 {
		return NewValidationError("name", "Name must be less than 100 characters")
	}
	if req.Email == "" {
		return NewValidationError("email", "Email is required")
	}
	// Simple email validation - in production, use a proper email validation library
	if len(req.Email) > 255 {
		return NewValidationError("email", "Email must be less than 255 characters")
	}
	return nil
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
