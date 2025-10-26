package api_formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ApiResponseFormatter implements structured API response format
type ApiResponseFormatter struct{}

func NewApiResponseFormatter() ResponseFormatter {
	return &ApiResponseFormatter{}
}

func (f *ApiResponseFormatter) Success(data any, message ...string) any {
	resp := &ApiResponse{
		Status: "success",
		Data:   data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	return resp
}

func (f *ApiResponseFormatter) Created(data any, message ...string) any {
	resp := &ApiResponse{
		Status: "success",
		Data:   data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	} else {
		resp.Message = "Resource created successfully"
	}
	return resp
}

func (f *ApiResponseFormatter) Error(code string, message string, details ...map[string]any) any {
	errorObj := &Error{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		errorObj.Details = details[0]
	}
	return &ApiResponse{
		Status: "error",
		Error:  errorObj,
	}
}

func (f *ApiResponseFormatter) ValidationError(message string, fields []FieldError) any {
	return &ApiResponse{
		Status: "error",
		Error: &Error{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Fields:  fields,
		},
	}
}

func (f *ApiResponseFormatter) NotFound(message string) any {
	return f.Error("NOT_FOUND", message)
}

func (f *ApiResponseFormatter) List(data any, meta *ListMeta) any {
	resp := &ApiResponse{
		Status: "success",
		Data:   data,
	}
	if meta != nil {
		resp.Meta = &Meta{
			ListMeta: meta,
		}
	}
	return resp
}

func (f *ApiResponseFormatter) ParseClientResponse(resp *http.Response, cr *ClientResponse) error {
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Store raw body and status code
	cr.RawBody = body
	cr.StatusCode = resp.StatusCode

	// Parse headers (optional, can be useful)
	cr.Headers = make(map[string]any)
	for key, values := range resp.Header {
		if len(values) == 1 {
			cr.Headers[key] = values[0]
		} else {
			cr.Headers[key] = values
		}
	}

	// Try to parse as ApiResponse format
	var apiResp ApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// If parsing fails, treat it as raw data
		cr.Status = "unknown"
		cr.Data = string(body)
		return nil
	}

	// Map ApiResponse to ClientResponse
	cr.Status = apiResp.Status
	cr.Message = apiResp.Message
	cr.Data = apiResp.Data
	cr.Error = apiResp.Error
	cr.Meta = apiResp.Meta

	return nil
}

var _ ResponseFormatter = (*ApiResponseFormatter)(nil)
