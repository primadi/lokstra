package auth

import (
	"context"
)

type Service interface {
	Login(ctx context.Context, input LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

type LoginRequest struct {
	Flow    string         // example: "password", "otp", "oauth2", "magic_link", etc.
	Payload map[string]any // sent to the auth service, e.g. {"username": "user", "password": "pass"}
}

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // dalam detik
}
