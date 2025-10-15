package auth_flow_otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

const SERVICE_TYPE = "auth_flow_otp"
const FLOW_NAME = "otp"

// Config represents the configuration for OTP-based auth flow.
type Config struct {
	UserRepoServiceName string `json:"user_repo_service_name" yaml:"user_repo_service_name"`
	KvStoreServiceName  string `json:"kvstore_service_name" yaml:"kvstore_service_name"`
	OTPLength           int    `json:"otp_length" yaml:"otp_length"`
	OTPTTLSeconds       int    `json:"otp_ttl_seconds" yaml:"otp_ttl_seconds"`
	MaxAttempts         int    `json:"max_attempts" yaml:"max_attempts"`
}

type otpFlow struct {
	cfg      *Config
	userRepo *service.Cached[auth.UserRepository]
	kvStore  *service.Cached[serviceapi.KvStore]
}

var _ auth.Flow = (*otpFlow)(nil)

func (f *otpFlow) Name() string {
	return FLOW_NAME
}

// Authenticate handles OTP verification
// Payload must contain: tenant_id, username, otp (6-digit code)
// To generate OTP, call GenerateOTP method separately
func (f *otpFlow) Authenticate(ctx context.Context, payload map[string]any) (*auth.Result, error) {
	tenantID, ok := payload["tenant_id"].(string)
	if !ok || tenantID == "" {
		return nil, auth.ErrInvalidCredentials
	}

	username, ok := payload["username"].(string)
	if !ok || username == "" {
		return nil, auth.ErrInvalidCredentials
	}

	otpCode, ok := payload["otp"].(string)
	if !ok || otpCode == "" {
		return nil, auth.ErrInvalidCredentials
	}

	// Get user from repository
	user, err := f.userRepo.MustGet().GetUserByName(ctx, tenantID, username)
	if err != nil {
		return nil, auth.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user is not active")
	}

	// Verify OTP
	otpKey := f.getOTPKey(tenantID, username)
	var storedOTP string
	if err := f.kvStore.MustGet().Get(ctx, otpKey, &storedOTP); err != nil {
		return nil, fmt.Errorf("OTP not found or expired")
	}

	if storedOTP != otpCode {
		// Track failed attempts
		f.incrementAttempts(ctx, tenantID, username)
		return nil, auth.ErrInvalidCredentials
	}

	// Delete used OTP
	f.kvStore.MustGet().Delete(ctx, otpKey)
	f.deleteAttempts(ctx, tenantID, username)

	// Return auth result
	return &auth.Result{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Metadata: map[string]any{
			"username":  user.Username,
			"email":     user.Email,
			"full_name": user.FullName,
		},
		IssuedAt: time.Now(),
	}, nil
}

// GenerateOTP generates a new OTP for the user
func (f *otpFlow) GenerateOTP(ctx context.Context, tenantID, username string) (string, error) {
	// Check if user exists
	user, err := f.userRepo.MustGet().GetUserByName(ctx, tenantID, username)
	if err != nil {
		return "", fmt.Errorf("user not found")
	}

	if !user.IsActive {
		return "", fmt.Errorf("user is not active")
	}

	// Check attempts
	attempts := f.getAttempts(ctx, tenantID, username)
	if attempts >= f.cfg.MaxAttempts {
		return "", fmt.Errorf("too many OTP generation attempts, please try again later")
	}

	// Generate OTP
	otp, err := f.generateRandomOTP(f.cfg.OTPLength)
	if err != nil {
		return "", err
	}

	// Store OTP in KvStore with TTL
	otpKey := f.getOTPKey(tenantID, username)
	ttl := time.Duration(f.cfg.OTPTTLSeconds) * time.Second
	if err := f.kvStore.MustGet().Set(ctx, otpKey, otp, ttl); err != nil {
		return "", err
	}

	return otp, nil
}

func (f *otpFlow) generateRandomOTP(length int) (string, error) {
	const digits = "0123456789"
	otp := make([]byte, length)
	for i := range otp {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		otp[i] = digits[num.Int64()]
	}
	return string(otp), nil
}

func (f *otpFlow) getOTPKey(tenantID, username string) string {
	return fmt.Sprintf("otp:%s:%s", tenantID, username)
}

func (f *otpFlow) getAttemptsKey(tenantID, username string) string {
	return fmt.Sprintf("otp_attempts:%s:%s", tenantID, username)
}

func (f *otpFlow) getAttempts(ctx context.Context, tenantID, username string) int {
	var attempts int
	key := f.getAttemptsKey(tenantID, username)
	f.kvStore.MustGet().Get(ctx, key, &attempts)
	return attempts
}

func (f *otpFlow) incrementAttempts(ctx context.Context, tenantID, username string) {
	attempts := f.getAttempts(ctx, tenantID, username)
	key := f.getAttemptsKey(tenantID, username)
	ttl := time.Duration(f.cfg.OTPTTLSeconds) * time.Second
	f.kvStore.MustGet().Set(ctx, key, attempts+1, ttl)
}

func (f *otpFlow) deleteAttempts(ctx context.Context, tenantID, username string) {
	key := f.getAttemptsKey(tenantID, username)
	f.kvStore.MustGet().Delete(ctx, key)
}

func (f *otpFlow) Shutdown() error {
	return nil
}

func Service(cfg *Config, userRepo *service.Cached[auth.UserRepository],
	kvStore *service.Cached[serviceapi.KvStore]) *otpFlow {
	return &otpFlow{
		cfg:      cfg,
		userRepo: userRepo,
		kvStore:  kvStore,
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		UserRepoServiceName: utils.GetValueFromMap(params, "user_repo_service_name", "auth_user_repo_pg"),
		KvStoreServiceName:  utils.GetValueFromMap(params, "kvstore_service_name", "kvstore_redis"),
		OTPLength:           utils.GetValueFromMap(params, "otp_length", 6),
		OTPTTLSeconds:       utils.GetValueFromMap(params, "otp_ttl_seconds", 300), // 5 minutes
		MaxAttempts:         utils.GetValueFromMap(params, "max_attempts", 5),
	}

	// Get UserRepository service from registry
	userRepo := service.LazyLoad[auth.UserRepository](cfg.UserRepoServiceName)

	// Get KvStore service from registry
	kvStore := service.LazyLoad[serviceapi.KvStore](cfg.KvStoreServiceName)

	return Service(cfg, userRepo, kvStore)
}

func Register() {
	lokstra_registry.RegisterServiceFactory(SERVICE_TYPE, ServiceFactory,
		lokstra_registry.AllowOverride(true))
}
