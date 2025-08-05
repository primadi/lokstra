package auth

import "errors"

const (
	MODULE_PREFIX = "lokstra.auth_"

	FLOW_PREFIX         = MODULE_PREFIX + "flow."
	SERVICE_PREFIX      = MODULE_PREFIX + "service."
	TOKEN_ISSUER_PREFIX = MODULE_PREFIX + "token."
	SESSION_PREFIX      = MODULE_PREFIX + "session."
	USER_REPO_PREFIX    = MODULE_PREFIX + "user_repo."
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrFlowNotFound = errors.New("auth flow not found")
var ErrTokenNotFound = errors.New("token not found")
var ErrTokenExpired = errors.New("token expired")
