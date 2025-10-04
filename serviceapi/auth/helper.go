package auth

import "errors"

const (
	// Auth related service types
	FLOW_TYPE         = "auth_flow"
	SERVICE_TYPE      = "auth_service"
	TOKEN_ISSUER_TYPE = "auth_token_issuer"
	SESSION_TYPE      = "auth_session"
	USER_REPO_TYPE    = "auth_user_repo"
	VALIDATOR_TYPE    = "auth_validator"
)

// Common auth errors
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrFlowNotFound = errors.New("auth flow not found")
var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")
