package response

import (
	"net/http"

	"github.com/primadi/lokstra/core/response/api_formatter"
)

// sends a successful response
func NewApiOk(data any) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().Success(data)
	a.resp.WithStatus(http.StatusOK).Json(formatted)
	return a
}

// sends a successful response with message
func NewApiOkWithMessage(data any, message string) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().Success(data, message)
	a.resp.WithStatus(http.StatusOK).Json(formatted)
	return a
}

// sends a 201 Created response
func NewApiCreated(data any, message string) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().Created(data, message)
	a.resp.WithStatus(http.StatusCreated).Json(formatted)
	return a
}

// sends a paginated list response
func NewApiOkList(data any, meta *api_formatter.ListMeta) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().List(data, meta)
	a.resp.WithStatus(http.StatusOK).Json(formatted)
	return a
}

// sends a paginated list response with full metadata
func NewApiOkListWithMeta(data any, meta *api_formatter.Meta) *ApiHelper {
	a := NewApiHelper()
	// Extract ListMeta from Meta if available
	var listMeta *api_formatter.ListMeta
	if meta != nil {
		listMeta = meta.ListMeta
	}
	formatted := api_formatter.GetGlobalFormatter().List(data, listMeta)
	a.resp.WithStatus(http.StatusOK).Json(formatted)
	return a
}

// sends an error response with code and message
func NewApiError(statusCode int, code, message string) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().Error(code, message)
	a.resp.WithStatus(statusCode).Json(formatted)
	return a
}

// sends an error response with additional details
func NewApiErrorWithDetails(statusCode int, code, message string,
	details map[string]any) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().Error(code, message, details)
	a.resp.WithStatus(statusCode).Json(formatted)
	return a
}

// sends a 400 validation error response
func NewApiValidationError(message string, fields []api_formatter.FieldError) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().ValidationError(message, fields)
	a.resp.WithStatus(http.StatusBadRequest).Json(formatted)
	return a
}

// sends a 400 bad request error
func NewApiBadRequest(code, message string) *ApiHelper {
	a := NewApiHelper()
	a.Error(http.StatusBadRequest, code, message)
	return a
}

// sends a 401 unauthorized error
func NewApiUnauthorized(message string) *ApiHelper {
	a := NewApiHelper()
	a.Error(http.StatusUnauthorized, "UNAUTHORIZED", message)
	return a
}

// sends a 403 forbidden error
func NewApiForbidden(message string) *ApiHelper {
	a := NewApiHelper()
	a.Error(http.StatusForbidden, "FORBIDDEN", message)
	return a
}

// NotFound sends a 404 not found error
func NewApiNotFound(message string) *ApiHelper {
	a := NewApiHelper()
	formatted := api_formatter.GetGlobalFormatter().NotFound(message)
	a.resp.WithStatus(http.StatusNotFound).Json(formatted)
	return a
}

// InternalError sends a 500 internal server error
func NewApiInternalError(message string) *ApiHelper {
	a := NewApiHelper()
	a.Error(http.StatusInternalServerError, "INTERNAL_ERROR", message)
	return a
}
