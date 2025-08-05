package auth_module

import (
	"context"

	"github.com/primadi/lokstra/serviceapi/auth"
	"golang.org/x/crypto/bcrypt"
)

type flowPassword struct {
	userRepo auth.UserRepository
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Authenticate implements auth.Flow.
func (f *flowPassword) Authenticate(ctx context.Context, payload map[string]any) (*auth.Result, error) {
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

	user, err := f.userRepo.GetUserByName(ctx, tenantID, username)
	if err != nil {
		return nil, err
	}

	if user == nil || !CheckPasswordHash(password, user.PasswordHash) {
		return nil, auth.ErrInvalidCredentials
	}

	return &auth.Result{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Metadata: user.Metadata,
	}, nil
}

// Name implements auth.Flow.
func (f *flowPassword) Name() string {
	return "password"
}

var _ auth.Flow = (*flowPassword)(nil)

func NewFlowPassword(userRepo auth.UserRepository) (auth.Flow, error) {
	return &flowPassword{
		userRepo: userRepo,
	}, nil
}
