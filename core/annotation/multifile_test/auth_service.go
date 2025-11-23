package main

import "github.com/primadi/lokstra/core/service"

// @RouterService name="auth-service", prefix="/api/v1/auth"
type AuthService struct {
	// @Inject "auth-repo"
	AuthRepo *service.Cached[interface{}]
}

// @Route "POST /login"
func (s *AuthService) Login(username string, password string) (string, error) {
	return "token", nil
}

// @Route "POST /logout"
func (s *AuthService) Logout() error {
	return nil
}
