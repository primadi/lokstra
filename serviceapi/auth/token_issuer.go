package auth

import (
	"context"
	"time"
)

type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
	IssueRefreshToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
}
