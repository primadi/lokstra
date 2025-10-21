package jwtauth

import (
	"context"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/old_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const MIDDLEWARE_TYPE = "jwtauth"

// ContextKey is the key used to store auth info in the request context
type ContextKey string

const (
	UserInfoKey    ContextKey = "user_info"
	TokenClaimsKey ContextKey = "token_claims"
)

// Config represents the configuration for JWT authentication middleware
type Config struct {
	ValidatorServiceName string   `json:"validator_service_name" yaml:"validator_service_name"`
	TokenHeader          string   `json:"token_header" yaml:"token_header"`   // Header name for token (default: "Authorization")
	TokenPrefix          string   `json:"token_prefix" yaml:"token_prefix"`   // Token prefix (default: "Bearer ")
	SkipPaths            []string `json:"skip_paths" yaml:"skip_paths"`       // Paths to skip auth
	ErrorMessage         string   `json:"error_message" yaml:"error_message"` // Custom error message
}

// Middleware creates a JWT authentication middleware
// It validates the JWT token and adds user info to the request context
func Middleware(cfg *Config) request.HandlerFunc {
	validator := service.LazyLoad[auth.Validator](cfg.ValidatorServiceName)

	return request.HandlerFunc(func(c *request.Context) error {
		// Check if path should skip auth
		path := c.R.URL.Path
		for _, skipPath := range cfg.SkipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Extract token from header
		authHeader := c.R.Header.Get(cfg.TokenHeader)
		if authHeader == "" {
			c.W.WriteHeader(http.StatusUnauthorized)
			c.W.Header().Set("Content-Type", "application/json")
			c.W.Write([]byte(`{"error":"missing authentication token"}`))
			return nil
		}

		// Remove prefix if present
		token := authHeader
		if cfg.TokenPrefix != "" && strings.HasPrefix(authHeader, cfg.TokenPrefix) {
			token = strings.TrimPrefix(authHeader, cfg.TokenPrefix)
		}

		// Validate token
		ctx := c.R.Context()
		claims, err := validator.MustGet().ValidateAccessToken(ctx, token)
		if err != nil {
			c.W.WriteHeader(http.StatusUnauthorized)
			c.W.Header().Set("Content-Type", "application/json")
			errMsg := cfg.ErrorMessage
			if errMsg == "" {
				errMsg = "invalid or expired token"
			}
			c.W.Write([]byte(`{"error":"` + errMsg + `"}`))
			return nil
		}

		// Get user info
		userInfo, err := validator.MustGet().GetUserInfo(ctx, claims)
		if err != nil {
			c.W.WriteHeader(http.StatusUnauthorized)
			c.W.Header().Set("Content-Type", "application/json")
			c.W.Write([]byte(`{"error":"failed to get user info"}`))
			return nil
		}

		// Store user info and claims in context
		ctx = context.WithValue(ctx, UserInfoKey, userInfo)
		ctx = context.WithValue(ctx, TokenClaimsKey, claims)
		c.R = c.R.WithContext(ctx)

		return c.Next()
	})
}

func MiddlewareFactory(params map[string]any) request.HandlerFunc {
	cfg := &Config{
		ValidatorServiceName: utils.GetValueFromMap(params, "validator_service_name", "auth_validator"),
		TokenHeader:          utils.GetValueFromMap(params, "token_header", "Authorization"),
		TokenPrefix:          utils.GetValueFromMap(params, "token_prefix", "Bearer "),
		SkipPaths:            utils.GetValueFromMap(params, "skip_paths", []string{}),
		ErrorMessage:         utils.GetValueFromMap(params, "error_message", ""),
	}

	return Middleware(cfg)
}

func Register() {
	old_registry.RegisterMiddlewareFactory(MIDDLEWARE_TYPE, MiddlewareFactory,
		old_registry.AllowOverride(true))
}

// Helper functions to extract auth info from context

// GetUserInfo extracts user info from the request context
func GetUserInfo(ctx context.Context) (*auth.UserInfo, bool) {
	userInfo, ok := ctx.Value(UserInfoKey).(*auth.UserInfo)
	return userInfo, ok
}

// GetTokenClaims extracts token claims from the request context
func GetTokenClaims(ctx context.Context) (*auth.TokenClaims, bool) {
	claims, ok := ctx.Value(TokenClaimsKey).(*auth.TokenClaims)
	return claims, ok
}

// GetCurrentUserID is a convenience function to get the current user ID
func GetCurrentUserID(ctx context.Context) string {
	if userInfo, ok := GetUserInfo(ctx); ok {
		return userInfo.UserID
	}
	return ""
}

// GetCurrentTenantID is a convenience function to get the current tenant ID
func GetCurrentTenantID(ctx context.Context) string {
	if userInfo, ok := GetUserInfo(ctx); ok {
		return userInfo.TenantID
	}
	return ""
}
