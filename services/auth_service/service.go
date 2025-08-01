package auth_service

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra/serviceapi/auth"
)

type AuthServiceImpl struct {
	FlowRegistry    map[string]auth.Flow
	TokenIssuer     auth.TokenIssuer
	SessionService  auth.Session
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// Login implements auth.Service.
func (a *AuthServiceImpl) Login(ctx context.Context, input auth.LoginRequest) (*auth.LoginResponse, error) {
	// Validate flow
	flow, ok := a.FlowRegistry[input.Flow]
	if !ok {
		return nil, auth.ErrFlowNotFound // Assuming this error is defined in the auth package
	}

	// Authenticate using the flow
	result, err := flow.Authenticate(ctx, input.Payload)
	if err != nil {
		return nil, err
	}

	// Issue an access token
	accessToken, err := a.TokenIssuer.IssueAccessToken(ctx, result, a.AccessTokenTTL)
	if err != nil {
		return nil, err
	}

	// Issue a refresh token
	refreshToken, err := a.TokenIssuer.IssueRefreshToken(ctx, result, a.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}

	// Store session data
	data := &auth.SessionData{
		TenantID: result.TenantID,
		UserID:   result.UserID,
		Metadata: result.Metadata,
	}
	if err := a.SessionService.Set(ctx, refreshToken, data, a.RefreshTokenTTL); err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(a.RefreshTokenTTL.Seconds()),
	}, nil
}

// Logout implements auth.Service.
func (a *AuthServiceImpl) Logout(ctx context.Context, refreshToken string) error {
	if err := a.SessionService.Delete(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

// RefreshToken implements auth.Service.
func (a *AuthServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*auth.LoginResponse, error) {
	session, err := a.SessionService.Get(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	_ = a.SessionService.Delete(ctx, refreshToken) // Delete old session

	authResult := &auth.Result{
		UserID:   session.UserID,
		TenantID: session.TenantID,
		Metadata: session.Metadata,
	}

	// Issue a new access token
	accessToken, err := a.TokenIssuer.IssueAccessToken(ctx, authResult, a.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	// Issue a new refresh token
	newRefreshToken, err := a.TokenIssuer.IssueRefreshToken(ctx, authResult, a.RefreshTokenTTL)
	if err != nil {
		return nil, err
	}

	// Update session with new refresh token
	if err := a.SessionService.Set(ctx, newRefreshToken, session, a.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("failed to extend session: %w", err)
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(a.AccessTokenTTL.Seconds()),
	}, nil
}

var _ auth.Service = (*AuthServiceImpl)(nil)
