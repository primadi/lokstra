package auth_module

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/primadi/lokstra/serviceapi/auth"
)

type JWTTokenIssuer struct {
	SecretKey    any
	SignInMethod any
}

// IssueAccessToken implements auth.TokenIssuer.
func (j *JWTTokenIssuer) IssueAccessToken(ctx context.Context, auth *auth.Result,
	ttl time.Duration) (string, error) {
	now := auth.IssuedAt
	if now.IsZero() {
		now = time.Now()
	}
	claims := JwtClaims{
		UserID:   auth.UserID,
		TenantID: auth.TenantID,
		Metadata: auth.Metadata,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	var (
		signingMethod jwt.SigningMethod
		signingKey    any
	)
	switch sm := j.SignInMethod.(type) {
	case jwt.SigningMethod:
		signingMethod = sm
		signingKey = j.SecretKey
	case string:
		switch sm {
		case "HS256":
			signingMethod = jwt.SigningMethodHS256
			switch key := j.SecretKey.(type) {
			case []byte:
				signingKey = key
			case string:
				signingKey = []byte(key)
			default:
				return "", fmt.Errorf("invalid secret key type for HS256: %T", j.SecretKey)
			}
		case "RS256":
			signingMethod = jwt.SigningMethodRS256
			switch key := j.SecretKey.(type) {
			case *rsa.PrivateKey:
				signingKey = key
			case string:
				block, _ := pem.Decode([]byte(key))
				if block == nil || block.Type != "RSA PRIVATE KEY" {
					return "", fmt.Errorf("invalid RSA private key PEM data")
				}
				privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				if err != nil {
					return "", fmt.Errorf("failed to parse RSA private key: %w", err)
				}
				signingKey = privKey
			case []byte:
				block, _ := pem.Decode(key)
				if block == nil || block.Type != "RSA PRIVATE KEY" {
					return "", fmt.Errorf("invalid RSA private key PEM data")
				}
				privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				if err != nil {
					return "", fmt.Errorf("failed to parse RSA private key: %w", err)
				}
				signingKey = privKey
			default:
				return "", fmt.Errorf("invalid secret key type for RS256: %T", j.SecretKey)
			}
		default:
			return "", fmt.Errorf("unsupported signing method: %s", sm)
		}
	}
	token := jwt.NewWithClaims(signingMethod, claims)
	return token.SignedString(signingKey)
}

// IssueRefreshToken implements auth.TokenIssuer.
func (j *JWTTokenIssuer) IssueRefreshToken(ctx context.Context, auth *auth.Result, ttl time.Duration) (string, error) {
	return generateRandomTokenString()
}

type JwtClaims struct {
	jwt.RegisteredClaims
	UserID   string         `json:"user_id"`
	TenantID string         `json:"tid,omitempty"`
	Metadata map[string]any `json:"meta,omitempty"`
}

var _ auth.TokenIssuer = (*JWTTokenIssuer)(nil)

func generateRandomTokenString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
