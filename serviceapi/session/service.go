package session

import (
	"context"
	"time"
)

type SessionData struct {
	UserID    string
	TenantID  string
	CreatedAt time.Time
	ExpiresAt time.Time
	Metadata  map[string]string // optional (IP, device, role, etc)
}

type Service interface {
	Set(ctx context.Context, sessionID string, data *SessionData, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	Delete(ctx context.Context, sessionID string) error
	DeleteByUser(ctx context.Context, tenantID, userID string) error
	ListByUser(ctx context.Context, tenantID, userID string) ([]string, error)
}
