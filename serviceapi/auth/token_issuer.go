package auth

import (
	"context"
	"errors"
	"time"
)

type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
	IssueRefreshToken(ctx context.Context, auth *Result, ttl time.Duration) (string, error)
}

var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")
