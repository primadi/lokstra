package auth

import (
	"context"
)

// Validator provides methods to validate and extract auth information from tokens.
// This is useful for middleware and request handlers.
type Validator interface {
	// ValidateAccessToken validates an access token and returns the claims.
	ValidateAccessToken(ctx context.Context, token string) (*TokenClaims, error)

	// ValidateRefreshToken validates a refresh token and returns the claims.
	ValidateRefreshToken(ctx context.Context, token string) (*TokenClaims, error)

	// GetUserInfo retrieves user information from the token claims.
	GetUserInfo(ctx context.Context, claims *TokenClaims) (*UserInfo, error)
}

// UserInfo represents authenticated user information extracted from token.
type UserInfo struct {
	UserID   string
	TenantID string
	Username string
	Email    string
	Metadata map[string]any
}
