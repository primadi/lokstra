package auth

import (
	"context"
	"time"
)

type Flow interface {
	Name() string
	Authenticate(ctx context.Context, payload map[string]any) (*Result, error)
}

type Result struct {
	UserID   string
	TenantID string
	Metadata map[string]any // optional (role, email, etc)
	IssuedAt time.Time
}
