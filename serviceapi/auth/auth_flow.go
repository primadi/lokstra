package auth

import (
	"context"
	"errors"
)

const (
	AUTH_FLOW_PREFIX = "lokstra.auth_flow."
)

type Flow interface {
	Name() string
	Authenticate(ctx context.Context, payload map[string]any) (*Result, error)
}

type Result struct {
	UserID   string
	TenantID string
	Metadata map[string]any // optional (role, email, etc)
}

var ErrInvalidCredentials = errors.New("invalid credentials")
