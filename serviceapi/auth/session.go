package auth

import (
	"context"
	"time"
)

type SessionData struct {
	UserID   string         `json:"user_id"`
	TenantID string         `json:"tenant_id"`
	Metadata map[string]any `json:"metadata,omitempty"` // optional (role, email, etc)
}

type Session interface {
	Set(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	Delete(ctx context.Context, sessionID string) error
	DeleteByUser(ctx context.Context, tenantID, userID string) error
	ListByUser(ctx context.Context, tenantID, userID string) ([]string, error)
}
