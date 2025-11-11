package domain

// Request DTOs for User module

type GetUserRequest struct {
	ID int `path:"id" validate:"required"`
}

type ListUsersRequest struct {
	Status string `query:"status"`
}

type CreateUserRequest struct {
	Name   string `json:"name" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
	RoleID int    `json:"role_id" validate:"required"`
}

type UpdateUserRequest struct {
	ID     int    `path:"id" validate:"required"`
	Name   string `json:"name" validate:"required"`
	Email  string `json:"email" validate:"required,email"`
	RoleID int    `json:"role_id" validate:"required"`
}

type SuspendUserRequest struct {
	ID     int    `path:"id" validate:"required"`
	Reason string `json:"reason"`
}

type ActivateUserRequest struct {
	ID int `path:"id" validate:"required"`
}

type DeleteUserRequest struct {
	ID int `path:"id" validate:"required"`
}
