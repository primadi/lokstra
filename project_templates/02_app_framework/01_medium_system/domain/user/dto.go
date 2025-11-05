package user

// Request/Response DTOs (Data Transfer Objects)

// GetUserParams contains parameters for getting a single user
type GetUserParams struct {
	ID int `path:"id" validate:"required"`
}

// ListUsersParams contains parameters for listing users
type ListUsersParams struct{}

// CreateUserParams contains parameters for creating a user
type CreateUserParams struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// UpdateUserParams contains parameters for updating a user
type UpdateUserParams struct {
	ID    int    `path:"id" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// DeleteUserParams contains parameters for deleting a user
type DeleteUserParams struct {
	ID int `path:"id" validate:"required"`
}
