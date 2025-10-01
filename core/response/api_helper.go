package response

import (
	"net/http"

	"github.com/primadi/lokstra/core/response/api_formatter"
) // SetApiResponseFormatter sets the global response formatter
func SetApiResponseFormatter(formatter api_formatter.ResponseFormatter) {
	api_formatter.SetGlobalFormatter(formatter)
}

// SetApiResponseFormatterByName sets the global response formatter by registered name
func SetApiResponseFormatterByName(name string) {
	api_formatter.SetGlobalFormatterByName(name)
}

// GetApiResponseFormatter returns the current global response formatter
func GetApiResponseFormatter() api_formatter.ResponseFormatter {
	return api_formatter.GetGlobalFormatter()
}

// ApiHelper provides opinionated API response helpers that wrap data in ApiResponse structure
type ApiHelper struct {
	resp *Response
}

// NewApiHelper creates a new API helper instance
func NewApiHelper(resp *Response) *ApiHelper {
	return &ApiHelper{resp: resp}
}

// Ok sends a successful response with data using configured formatter
func (a *ApiHelper) Ok(data any) error {
	formatted := api_formatter.GetGlobalFormatter().Success(data)
	return a.resp.WithStatus(http.StatusOK).Json(formatted)
}

// OkWithMessage sends a successful response with message and data using configured formatter
func (a *ApiHelper) OkWithMessage(data any, message string) error {
	formatted := api_formatter.GetGlobalFormatter().Success(data, message)
	return a.resp.WithStatus(http.StatusOK).Json(formatted)
}

// Created sends a 201 Created response with data using configured formatter
func (a *ApiHelper) Created(data any, message string) error {
	formatted := api_formatter.GetGlobalFormatter().Created(data, message)
	return a.resp.WithStatus(http.StatusCreated).Json(formatted)
}

// OkList sends a paginated list response using configured formatter
func (a *ApiHelper) OkList(data any, meta *api_formatter.ListMeta) error {
	formatted := api_formatter.GetGlobalFormatter().List(data, meta)
	return a.resp.WithStatus(http.StatusOK).Json(formatted)
}

// OkListWithMeta sends a paginated list response with full metadata
func (a *ApiHelper) OkListWithMeta(data any, meta *api_formatter.Meta) error {
	// Extract ListMeta from Meta if available
	var listMeta *api_formatter.ListMeta
	if meta != nil {
		listMeta = meta.ListMeta
	}
	formatted := api_formatter.GetGlobalFormatter().List(data, listMeta)
	return a.resp.WithStatus(http.StatusOK).Json(formatted)
}

// Error sends an error response with code and message
func (a *ApiHelper) Error(statusCode int, code, message string) error {
	formatted := api_formatter.GetGlobalFormatter().Error(code, message)
	return a.resp.WithStatus(statusCode).Json(formatted)
}

// ErrorWithDetails sends an error response with additional details
func (a *ApiHelper) ErrorWithDetails(statusCode int, code, message string, details map[string]any) error {
	formatted := api_formatter.GetGlobalFormatter().Error(code, message, details)
	return a.resp.WithStatus(statusCode).Json(formatted)
}

// ValidationError sends a 400 validation error response
func (a *ApiHelper) ValidationError(message string, fields []api_formatter.FieldError) error {
	formatted := api_formatter.GetGlobalFormatter().ValidationError(message, fields)
	return a.resp.WithStatus(http.StatusBadRequest).Json(formatted)
}

// BadRequest sends a 400 bad request error
func (a *ApiHelper) BadRequest(code, message string) error {
	return a.Error(http.StatusBadRequest, code, message)
}

// Unauthorized sends a 401 unauthorized error
func (a *ApiHelper) Unauthorized(message string) error {
	return a.Error(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 forbidden error
func (a *ApiHelper) Forbidden(message string) error {
	return a.Error(http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 not found error
func (a *ApiHelper) NotFound(message string) error {
	formatted := api_formatter.GetGlobalFormatter().NotFound(message)
	return a.resp.WithStatus(http.StatusNotFound).Json(formatted)
}

// InternalError sends a 500 internal server error
func (a *ApiHelper) InternalError(message string) error {
	return a.Error(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
