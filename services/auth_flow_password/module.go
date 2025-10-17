package auth_flow_password

import (
	"context"
	"fmt"
	"time"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/serviceapi/auth"
	"golang.org/x/crypto/bcrypt"
)

const SERVICE_TYPE = "auth_flow_password"
const FLOW_NAME = "password"

// Config represents the configuration for password-based auth flow.
type Config struct {
	UserRepoServiceName string `json:"user_repo_service_name" yaml:"user_repo_service_name"`
}

type passwordFlow struct {
	cfg      *Config
	userRepo *service.Cached[auth.UserRepository]
}

var _ auth.Flow = (*passwordFlow)(nil)

func (f *passwordFlow) Name() string {
	return FLOW_NAME
}

func (f *passwordFlow) Authenticate(ctx context.Context, payload map[string]any) (*auth.Result, error) {
	// Extract credentials from payload
	tenantID, ok := payload["tenant_id"].(string)
	if !ok || tenantID == "" {
		return nil, auth.ErrInvalidCredentials
	}

	username, ok := payload["username"].(string)
	if !ok || username == "" {
		return nil, auth.ErrInvalidCredentials
	}

	password, ok := payload["password"].(string)
	if !ok || password == "" {
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

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, auth.ErrInvalidCredentials
	}

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

func (f *passwordFlow) Shutdown() error {
	return nil
}

func Service(cfg *Config, userRepo *service.Cached[auth.UserRepository]) *passwordFlow {
	return &passwordFlow{
		cfg:      cfg,
		userRepo: userRepo,
	}
}

func ServiceFactory(params map[string]any) any {
	cfg := &Config{
		UserRepoServiceName: utils.GetValueFromMap(params,
			"user_repo_service_name", "auth_user_repo_pg"),
	}

	// Get UserRepository service from registry
	userRepo := service.LazyLoad[auth.UserRepository](cfg.UserRepoServiceName)

	return Service(cfg, userRepo)
}

func Register() {
	lokstra_registry.RegisterServiceType(SERVICE_TYPE, ServiceFactory,
		lokstra_registry.AllowOverride(true))
}
