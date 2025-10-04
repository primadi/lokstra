package auth

import (
	"context"
	"time"
)

type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
	IssueRefreshToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
	VerifyToken(ctx context.Context, token string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID    string         `json:"user_id"`
	TenantID  string         `json:"tenant_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	TokenType string         `json:"token_type"` // "access" or "refresh"
	IssuedAt  time.Time      `json:"issued_at"`
	ExpiresAt time.Time      `json:"expires_at"`
}
