package userdomain

import "time"

// UserDTO user data transfer object
type UserDTO struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

// CreateUserRequest create user request
type CreateUserRequest struct {
	Name  string
	Email string
}
