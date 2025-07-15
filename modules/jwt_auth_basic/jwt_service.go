package jwt_auth_basic

import (
	"fmt"
	"lokstra/common/iface"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthService struct {
	instanceName string
	secret       string
	expires      time.Duration
	validator    AuthCredentialValidator
	config       map[string]any
}

func NewJWTAuthService(instanceName string, cfg map[string]any) (*JWTAuthService, error) {
	secret, ok := cfg["secret"].(string)
	if !ok {
		return nil, fmt.Errorf("jwt service requires 'secret' in config")
	}

	expiresHours := 24
	if exp, ok := cfg["expires_hours"].(int); ok {
		expiresHours = exp
	}

	return &JWTAuthService{
		instanceName: instanceName,
		secret:       secret,
		expires:      time.Duration(expiresHours) * time.Hour,
		config:       cfg,
	}, nil
}

var _ iface.Service = (*JWTAuthService)(nil)

func (j *JWTAuthService) InstanceName() string {
	return j.instanceName
}

func (j *JWTAuthService) GetConfig(key string) any {
	return j.config[key]
}

func (j *JWTAuthService) SetValidator(v AuthCredentialValidator) {
	j.validator = v
}

func (j *JWTAuthService) GetValidator() AuthCredentialValidator {
	return j.validator
}

func (j *JWTAuthService) GenerateToken(userID, roleID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role_id": roleID,
		"exp":     time.Now().Add(j.expires).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTAuthService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (j *JWTAuthService) Login(username, password string) (string, error) {
	if j.validator == nil {
		return "", fmt.Errorf("no credential validator configured")
	}

	userID, roleID, err := j.validator.ValidateCredentials(username, password)
	if err != nil {
		return "", err
	}

	return j.GenerateToken(userID, roleID)
}
