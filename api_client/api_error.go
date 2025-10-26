package api_client

import (
	"net/http"
)

// ApiError represents an error from a remote API call with HTTP status information.
// This allows callers to distinguish between different types of errors (BadRequest, NotFound, etc.)
// and handle them appropriately.
//
// Usage in service layer:
//
//	user, err := api_client.FetchAndCast[*User](client, "/users/123")
//	if err != nil {
//	    if apiErr, ok := err.(*api_client.ApiError); ok {
//	        // Handle specific error types based on status code
//	        switch apiErr.StatusCode {
//	        case 404:
//	            return nil, fmt.Errorf("user not found")
//	        case 401:
//	            return nil, fmt.Errorf("unauthorized")
//	        default:
//	            return nil, apiErr
//	        }
//	    }
//	    return nil, err // Other error types
//	}
//
// Usage in HTTP handler:
//
//	user, err := userService.GetUser(ctx, req)
//	if err != nil {
//	    if apiErr, ok := err.(*api_client.ApiError); ok {
//	        return ctx.Api.Error(apiErr.StatusCode, apiErr.Code, apiErr.Message)
//	    }
//	    return ctx.Api.InternalError(err.Error())
//	}
type ApiError struct {
	StatusCode int            // HTTP status code (400, 401, 404, 500, etc.)
	Code       string         // Error code (e.g., "VALIDATION_ERROR", "NOT_FOUND")
	Message    string         // Human-readable error message
	Details    map[string]any // Optional additional error details
}

// Error implements the error interface
func (e *ApiError) Error() string {
	return e.Message
	// if e.Code != "" {
	// 	return fmt.Sprintf("[%d %s] %s", e.StatusCode, e.Code, e.Message)
	// }
	// return fmt.Sprintf("[%d] %s", e.StatusCode, e.Message)
}

// IsClientError returns true if the error is a 4xx client error
func (e *ApiError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsServerError returns true if the error is a 5xx server error
func (e *ApiError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsBadRequest returns true if the error is a 400 Bad Request
func (e *ApiError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsUnauthorized returns true if the error is a 401 Unauthorized
func (e *ApiError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden
func (e *ApiError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsNotFound returns true if the error is a 404 Not Found
func (e *ApiError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// NewApiError creates a new ApiError
func NewApiError(statusCode int, code, message string) *ApiError {
	return &ApiError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// NewApiErrorWithDetails creates a new ApiError with additional details
func NewApiErrorWithDetails(statusCode int, code, message string, details map[string]any) *ApiError {
	return &ApiError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Details:    details,
	}
}
