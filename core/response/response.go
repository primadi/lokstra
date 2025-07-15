package response

import (
	"net/http"
)

// ResponseCode is a logical code representing the result of an operation.
// It is independent of HTTP status codes and intended to standardize internal response semantics.
//
// This type is designed to be extensible. Developers are encouraged to define their own
// ResponseCode constants in their own packages to represent application-specific outcomes.
//
// Example:
//
//	package user
//
//	import "lokstra/common/response"
//
//	const (
//	    CodeUserSuspended response.ResponseCode = "USER_SUSPENDED"
//	    CodeInvalidOTP    response.ResponseCode = "INVALID_OTP"
//	)
//
// These custom codes can then be used throughout the application for consistent error handling,
// logging, and localization. For instance, they can act as keys in a multi-language response system:
//
//	var ResponseMessages = map[response.ResponseCode]map[string]string{
//	    CodeUserSuspended: {
//	        "en": "Your account is suspended.",
//	        "id": "Akun Anda ditangguhkan.",
//	    },
//	    ...
//	}
//
// The built-in codes provided by Lokstra include:
//
//	CodeOK, CodeCreated, CodeUpdated,
//	CodeNotFound, CodeDuplicate,
//	CodeBadRequest, CodeInternal
//
// Developers are free to expand this set as needed.
type Response struct {
	StatusCode   int          `json:"-"`              // HTTP status
	ResponseCode ResponseCode `json:"code,omitempty"` // Logical code
	Success      bool         `json:"success"`        // true/false

	Message string      `json:"message,omitempty"` // Localizable message
	Data    any         `json:"data,omitempty"`    // Payload
	Meta    any         `json:"meta,omitempty"`    // List pagination info
	Headers http.Header `json:"-"`                 // HTTP only
	RawData []byte      `json:"-"`                 // Raw data for non-JSON responses

	FieldErrors map[string]string `json:"errors,omitempty"` // For Refine form field errors
}

func NewResponse() *Response {
	return &Response{}
}

// WithX helpers
func (r *Response) WithMessage(msg string) *Response {
	r.Message = msg
	return r
}

func (r *Response) WithData(data any) *Response {
	r.Data = data
	return r
}

func (r *Response) WithMeta(meta any) *Response {
	r.Meta = meta
	return r
}

func (r *Response) WithHeader(key, value string) *Response {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set(key, value)
	return r
}

func (r *Response) GetHeaders() http.Header {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	return r.Headers
}

func (r *Response) GetStatusCode() int {
	if r.StatusCode == 0 {
		r.StatusCode = http.StatusOK // Default to 200 OK if not set
	}
	return r.StatusCode
}

func (r *Response) SetStatusCode(code int) *Response {
	r.StatusCode = code
	return r
}
