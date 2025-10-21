package auth_validator

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/old_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const SERVICE_TYPE = "auth_validator"

// Config represents the configuration for auth validator service.
type Config struct {
	TokenIssuerServiceName string `json:"token_issuer_service_name" yaml:"token_issuer_service_name"`
	UserRepoServiceName    string `json:"user_repo_service_name" yaml:"user_repo_service_name"`
}

type authValidator struct {
	cfg         *Config
	tokenIssuer *service.Cached[auth.TokenIssuer]
	userRepo    *service.Cached[auth.UserRepository]
}

var _ auth.Validator = (*authValidator)(nil)

func (v *authValidator) ValidateAccessToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	claims, err := v.tokenIssuer.MustGet().VerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, fmt.Errorf("invalid token type, expected access token")
	}

	return claims, nil
}

func (v *authValidator) ValidateRefreshToken(ctx context.Context, token string) (*auth.TokenClaims, error) {
	claims, err := v.tokenIssuer.MustGet().VerifyToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, fmt.Errorf("invalid token type, expected refresh token")
	}

	return claims, nil
}

func (v *authValidator) GetUserInfo(ctx context.Context, claims *auth.TokenClaims) (*auth.UserInfo, error) {
	userInfo := &auth.UserInfo{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Metadata: claims.Metadata,
	}

	// Extract username and email from metadata if available
	if username, ok := claims.Metadata["username"].(string); ok {
		userInfo.Username = username
	}
	if email, ok := claims.Metadata["email"].(string); ok {
		userInfo.Email = email
	}

	// If user repo is available and metadata doesn't have complete info,
	// fetch from database
	if v.userRepo != nil && (userInfo.Username == "" || userInfo.Email == "") {
		// Note: This requires username to be in metadata or using UserID as lookup
		// For now, we'll just return what we have from metadata
	}

	return userInfo, nil
}

func (v *authValidator) Shutdown() error {
	return nil
}

func Service(cfg *Config, tokenIssuer *service.Cached[auth.TokenIssuer],
	userRepo *service.Cached[auth.UserRepository]) *authValidator {
	return &authValidator{
		cfg:         cfg,
		tokenIssuer: tokenIssuer,
		userRepo:    userRepo,
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		TokenIssuerServiceName: utils.GetValueFromMap(params, "token_issuer_service_name", "auth_token_jwt"),
		UserRepoServiceName:    utils.GetValueFromMap(params, "user_repo_service_name", ""),
	}

	// Get TokenIssuer service from registry
	tokenIssuer := service.LazyLoad[auth.TokenIssuer](cfg.TokenIssuerServiceName)

	// Get UserRepository service from registry (optional)
	var userRepo *service.Cached[auth.UserRepository]
	if cfg.UserRepoServiceName != "" {
		userRepo = service.LazyLoad[auth.UserRepository](cfg.UserRepoServiceName)
	}

	return Service(cfg, tokenIssuer, userRepo)
}

func Register() {
	old_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		old_registry.AllowOverride(true))
}
