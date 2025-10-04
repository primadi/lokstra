package api_client

import (
	"fmt"
	"net/http"

	"github.com/primadi/lokstra/common/cast"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

type FetchOption func(*fetchConfig)

type fetchConfig struct {
	Headers    map[string]string
	Formatter  api_formatter.ResponseFormatter
	Method     string
	CustomFunc func(*http.Response, *api_formatter.ClientResponse) (any, error)
	Body       any
}

// WithHeaders sets custom headers for the request
func WithHeaders(headers map[string]string) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.Headers = headers
	}
}

// WithFormatter sets a custom formatter for the response
func WithFormatter(formatter api_formatter.ResponseFormatter) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.Formatter = formatter
	}
}

// WithMethod sets the HTTP method (GET, POST, etc)
func WithMethod(method string) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.Method = method
	}
}

// WithCustomFunc allows custom handling of the response and parsed client response, returns result and error
func WithCustomFunc(fn func(*http.Response, *api_formatter.ClientResponse) (any, error)) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.CustomFunc = fn
	}
}

// WithBody sets the request body for POST, PUT, PATCH
func WithBody(body any) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.Body = body
	}
}

// FetchAndCast is a flexible fetch helper with options (headers, formatter, method, body, custom func, etc)
func FetchAndCast[T any](c *request.Context, client *ClientRouter, path string, opts ...FetchOption) (*T, error) {
	cfg := &fetchConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	resp, err := client.Method(method, path, cfg.Body, cfg.Headers)
	if err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to fetch: %v", err))
	}

	formatter := cfg.Formatter
	if formatter == nil {
		formatter = api_formatter.GetGlobalFormatter()
	}

	clientResp := &api_formatter.ClientResponse{}
	if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to parse response: %v", err))
	}

	// If CustomFunc is provided, delegate all handling to it
	if cfg.CustomFunc != nil {
		customResult, err := cfg.CustomFunc(resp, clientResp)
		if err != nil {
			return nil, err
		}

		// If custom function returns nil, continue with default flow
		if customResult != nil {
			// Try direct type assertion first
			if result, ok := customResult.(*T); ok {
				return result, nil
			}
			// Try value type assertion
			if result, ok := customResult.(T); ok {
				return &result, nil
			}
			// Fallback: try to cast using reflection
			var result T
			if err := cast.ToStruct(customResult, &result, true); err != nil {
				return nil, c.Api.InternalError(fmt.Sprintf("Failed to cast custom result: %v", err))
			}
			return &result, nil
		}
	}

	// Forward the exact error from downstream API if not successful
	if clientResp.Status != "success" {
		// Extract error code and message from response
		code := "API_ERROR"
		message := clientResp.Message
		if message == "" {
			message = "Downstream API returned error"
		}

		if clientResp.Error != nil {
			if clientResp.Error.Code != "" {
				code = clientResp.Error.Code
			}
			if clientResp.Error.Message != "" {
				message = clientResp.Error.Message
			}
		}

		// Return error based on HTTP status code
		switch resp.StatusCode {
		case http.StatusBadRequest:
			return nil, c.Api.BadRequest(code, message)
		case http.StatusUnauthorized:
			return nil, c.Api.Unauthorized(message)
		case http.StatusForbidden:
			return nil, c.Api.Forbidden(message)
		case http.StatusNotFound:
			return nil, c.Api.NotFound(message)
		default:
			if resp.StatusCode >= 500 {
				return nil, c.Api.InternalError(message)
			}
			// For other status codes, use BadRequest with code
			return nil, c.Api.BadRequest(code, message)
		}
	}

	var result T
	if err := cast.ToStruct(clientResp.Data, &result, true); err != nil {
		return nil, c.Api.InternalError(fmt.Sprintf("Failed to cast data: %v", err))
	}
	return &result, nil
}
