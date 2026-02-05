package main

import (
	"github.com/primadi/lokstra/core/annotation/internal/multifile_test/authdomain"
	"github.com/primadi/lokstra/core/service"
)

// @Handler name="auth-service", prefix="/api/v1/auth"
type AuthService struct {
	// @Inject "auth-repo"
	AuthRepo *service.Cached[any]
}

// @Route "POST /login"
func (s *AuthService) Login(req *authdomain.LoginRequest) (*authdomain.AuthResponse, error) {
	// Validate and authenticate user credentials
	return &authdomain.AuthResponse{
		Token:   "token-123",
		Message: "Welcome " + req.Username,
	}, nil
}

// @Route "POST /logout"
func (s *AuthService) Logout(req *authdomain.LogoutRequest) (*authdomain.AuthResponse, error) {
	return &authdomain.AuthResponse{
		Message: "Logged out successfully",
	}, nil
}
