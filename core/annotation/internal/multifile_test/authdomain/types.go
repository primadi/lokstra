package authdomain

import "context"

// LoginRequest login request
type LoginRequest struct {
	Username string
	Password string
}

// AuthResponse authentication response
type AuthResponse struct {
	Token   string
	Message string
}

// LogoutRequest logout request with context
type LogoutRequest struct {
	Ctx context.Context
}
