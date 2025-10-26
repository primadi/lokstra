package auth_token_jwt

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const SERVICE_TYPE = "auth_token_jwt"

// Config represents the configuration for JWT token issuer service.
type Config struct {
	SecretKey  string        `json:"secret_key" yaml:"secret_key"`   // Secret key for signing tokens
	Issuer     string        `json:"issuer" yaml:"issuer"`           // Token issuer name
	AccessTTL  time.Duration `json:"access_ttl" yaml:"access_ttl"`   // Default access token TTL
	RefreshTTL time.Duration `json:"refresh_ttl" yaml:"refresh_ttl"` // Default refresh token TTL
}

type tokenIssuerJWT struct {
	cfg *Config
}

var _ auth.TokenIssuer = (*tokenIssuerJWT)(nil)

type Claims struct {
	UserID    string         `json:"user_id"`
	TenantID  string         `json:"tenant_id"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	TokenType string         `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func (t *tokenIssuerJWT) IssueAccessToken(ctx context.Context, result *auth.Result, ttl time.Duration) (string, error) {
	if ttl == 0 {
		ttl = t.cfg.AccessTTL
	}

	claims := Claims{
		UserID:    result.UserID,
		TenantID:  result.TenantID,
		Metadata:  result.Metadata,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(result.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(result.IssuedAt.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.cfg.SecretKey))
}

func (t *tokenIssuerJWT) IssueRefreshToken(ctx context.Context, result *auth.Result, ttl time.Duration) (string, error) {
	if ttl == 0 {
		ttl = t.cfg.RefreshTTL
	}

	claims := Claims{
		UserID:    result.UserID,
		TenantID:  result.TenantID,
		Metadata:  result.Metadata,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(result.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(result.IssuedAt.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(t.cfg.SecretKey))
}

func (t *tokenIssuerJWT) VerifyToken(ctx context.Context, tokenString string) (*auth.TokenClaims, error) {
	claims, err := t.ParseToken(tokenString)
	if err != nil {
		return nil, err
	}

	return &auth.TokenClaims{
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		Metadata:  claims.Metadata,
		TokenType: claims.TokenType,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
func (t *tokenIssuerJWT) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.cfg.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (t *tokenIssuerJWT) Shutdown() error {
	return nil
}

func Service(cfg *Config) *tokenIssuerJWT {
	return &tokenIssuerJWT{cfg: cfg}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		SecretKey:  utils.GetValueFromMap(params, "secret_key", "change-me-in-production"),
		Issuer:     utils.GetValueFromMap(params, "issuer", "lokstra"),
		AccessTTL:  utils.GetValueFromMap(params, "access_ttl", 15*time.Minute),
		RefreshTTL: utils.GetValueFromMap(params, "refresh_ttl", 7*24*time.Hour),
	}
	return Service(cfg)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		nil)
}
