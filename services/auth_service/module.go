package auth_service

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/old_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const SERVICE_TYPE = "auth_service"

// Config represents the configuration for auth service.
type Config struct {
	TokenIssuerServiceName string            `json:"token_issuer_service_name" yaml:"token_issuer_service_name"`
	SessionServiceName     string            `json:"session_service_name" yaml:"session_service_name"`
	FlowServiceNames       map[string]string `json:"flow_service_names" yaml:"flow_service_names"` // map[flowName]serviceName
	AccessTokenTTL         time.Duration     `json:"access_token_ttl" yaml:"access_token_ttl"`
	RefreshTokenTTL        time.Duration     `json:"refresh_token_ttl" yaml:"refresh_token_ttl"`
}

type authService struct {
	cfg         *Config
	tokenIssuer *service.Cached[auth.TokenIssuer]
	session     *service.Cached[auth.Session]
	flows       map[string]*service.Cached[auth.Flow]
}

var _ auth.Service = (*authService)(nil)

func (s *authService) Login(ctx context.Context, input auth.LoginRequest) (*auth.LoginResponse, error) {
	// Get the appropriate flow
	flow, exists := s.flows[input.Flow]
	if !exists {
		return nil, auth.ErrFlowNotFound
	}

	// Authenticate using the flow
	result, err := flow.MustGet().Authenticate(ctx, input.Payload)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.tokenIssuer.MustGet().IssueAccessToken(ctx, result, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	refreshToken, err := s.tokenIssuer.MustGet().IssueRefreshToken(ctx, result, s.cfg.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to issue refresh token: %w", err)
	}

	// Store session
	sessionData := &auth.SessionData{
		UserID:   result.UserID,
		TenantID: result.TenantID,
		Metadata: result.Metadata,
	}

	if err := s.session.MustGet().Set(ctx, refreshToken, sessionData, s.cfg.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*auth.LoginResponse, error) {
	// Get session data
	sessionData, err := s.session.MustGet().Get(ctx, refreshToken)
	if err != nil {
		return nil, auth.ErrTokenNotFound
	}

	// Verify refresh token with token issuer
	claims, err := s.tokenIssuer.MustGet().VerifyToken(ctx, refreshToken)
	if err != nil {
		return nil, auth.ErrTokenExpired
	}

	// Verify token type
	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type")
	}

	// Create new auth result from session data
	result := &auth.Result{
		UserID:   sessionData.UserID,
		TenantID: sessionData.TenantID,
		Metadata: sessionData.Metadata,
		IssuedAt: time.Now(),
	}

	// Generate new access token
	accessToken, err := s.tokenIssuer.MustGet().IssueAccessToken(ctx, result, s.cfg.AccessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to issue access token: %w", err)
	}

	// Optionally generate new refresh token (rotate refresh token)
	newRefreshToken, err := s.tokenIssuer.MustGet().IssueRefreshToken(ctx, result, s.cfg.RefreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to issue refresh token: %w", err)
	}

	// Delete old refresh token session
	if err := s.session.MustGet().Delete(ctx, refreshToken); err != nil {
		// Log error but don't fail the request
	}

	// Store new session
	if err := s.session.MustGet().Set(ctx, newRefreshToken, sessionData, s.cfg.RefreshTokenTTL); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &auth.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.session.MustGet().Delete(ctx, refreshToken)
}

func (s *authService) Shutdown() error {
	return nil
}

func Service(cfg *Config, tokenIssuer *service.Cached[auth.TokenIssuer],
	session *service.Cached[auth.Session],
	flows map[string]*service.Cached[auth.Flow]) *authService {
	return &authService{
		cfg:         cfg,
		tokenIssuer: tokenIssuer,
		session:     session,
		flows:       flows,
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		TokenIssuerServiceName: utils.GetValueFromMap(params, "token_issuer_service_name", "auth_token_jwt"),
		SessionServiceName:     utils.GetValueFromMap(params, "session_service_name", "auth_session_redis"),
		FlowServiceNames: utils.GetValueFromMap(params, "flow_service_names", map[string]string{
			"password": "auth_flow_password",
			"otp":      "auth_flow_otp",
		}),
		AccessTokenTTL:  utils.GetValueFromMap(params, "access_token_ttl", 15*time.Minute),
		RefreshTokenTTL: utils.GetValueFromMap(params, "refresh_token_ttl", 7*24*time.Hour),
	}

	// Get TokenIssuer service from registry
	tokenIssuer := service.LazyLoad[auth.TokenIssuer](cfg.TokenIssuerServiceName)

	// Get Session service from registry
	session := service.LazyLoad[auth.Session](cfg.SessionServiceName)

	// Get Flow services from registry
	flows := make(map[string]*service.Cached[auth.Flow])
	for flowName, serviceName := range cfg.FlowServiceNames {
		flow := service.LazyLoad[auth.Flow](serviceName)
		flows[flowName] = flow
	}

	return Service(cfg, tokenIssuer, session, flows)
}

func Register() {
	old_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		old_registry.AllowOverride(true))
}
