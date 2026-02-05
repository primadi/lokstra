package auth

import "fmt"

// Error codes for authentication module
const (
	ErrCodeInvalidCredentials  = "AUTH_001"
	ErrCodeTokenExpired        = "AUTH_002"
	ErrCodeInvalidToken        = "AUTH_003"
	ErrCodeAccountSuspended    = "AUTH_004"
	ErrCodeTenantInactive      = "AUTH_005"
	ErrCodeAccountLocked       = "AUTH_006"
	ErrCodeInsufficientPerms   = "AUTH_007"
	ErrCodePasswordComplexity  = "AUTH_008"
	ErrCodeUsernameTaken       = "AUTH_009"
	ErrCodeEmailExistsInTenant = "AUTH_010"
	ErrCodeRateLimited         = "AUTH_011"
	ErrCodeInvalidResetToken   = "AUTH_012"
)

// AuthError represents an authentication error
type AuthError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Predefined errors
var (
	ErrInvalidCredentials = &AuthError{
		Code:       ErrCodeInvalidCredentials,
		Message:    "Email or password is incorrect",
		HTTPStatus: 401,
	}

	ErrTokenExpired = &AuthError{
		Code:       ErrCodeTokenExpired,
		Message:    "Your session has expired. Please login again",
		HTTPStatus: 401,
	}

	ErrInvalidToken = &AuthError{
		Code:       ErrCodeInvalidToken,
		Message:    "Invalid authentication token",
		HTTPStatus: 401,
	}

	ErrAccountSuspended = &AuthError{
		Code:       ErrCodeAccountSuspended,
		Message:    "Your account has been suspended. Contact administrator",
		HTTPStatus: 403,
	}

	ErrTenantInactive = &AuthError{
		Code:       ErrCodeTenantInactive,
		Message:    "Organization account is not active",
		HTTPStatus: 403,
	}

	ErrAccountLocked = &AuthError{
		Code:       ErrCodeAccountLocked,
		Message:    "Account locked due to multiple failed attempts. Try again in 15 minutes",
		HTTPStatus: 403,
	}

	ErrInsufficientPermissions = &AuthError{
		Code:       ErrCodeInsufficientPerms,
		Message:    "You don't have permission to perform this action",
		HTTPStatus: 403,
	}

	ErrPasswordComplexity = &AuthError{
		Code:       ErrCodePasswordComplexity,
		Message:    "Password does not meet complexity requirements",
		HTTPStatus: 400,
	}

	ErrUsernameTaken = &AuthError{
		Code:       ErrCodeUsernameTaken,
		Message:    "Username is already taken",
		HTTPStatus: 400,
	}

	ErrEmailExistsInTenant = &AuthError{
		Code:       ErrCodeEmailExistsInTenant,
		Message:    "Email already registered in this organization",
		HTTPStatus: 400,
	}

	ErrRateLimited = &AuthError{
		Code:       ErrCodeRateLimited,
		Message:    "Too many attempts. Please try again later",
		HTTPStatus: 429,
	}

	ErrInvalidResetToken = &AuthError{
		Code:       ErrCodeInvalidResetToken,
		Message:    "Password reset link is invalid or expired",
		HTTPStatus: 400,
	}
)

// NewAccountLockedError creates a locked error with remaining time
func NewAccountLockedError(remainingMinutes int) *AuthError {
	return &AuthError{
		Code:       ErrCodeAccountLocked,
		Message:    fmt.Sprintf("Account locked. Try again in %d minutes", remainingMinutes),
		HTTPStatus: 403,
	}
}
