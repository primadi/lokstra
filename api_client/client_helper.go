package api_client

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/primadi/lokstra/common/cast"
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
// Returns ApiError on HTTP errors to preserve status code information for proper error handling.
//
// Note: Reflection overhead for type checking is minimal (~8ns per call) and caching adds more overhead
// than it saves. The code is kept simple and readable without premature optimization.
func FetchAndCast[T any](client *ClientRouter, path string, opts ...FetchOption) (T, error) {
	cfg := &fetchConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	method := cfg.Method
	if method == "" {
		method = "GET"
	}

	var zero T

	resp, err := client.Method(method, path, cfg.Body, cfg.Headers)
	if err != nil {
		return zero, fmt.Errorf("failed to fetch: %v", err)
	}

	formatter := cfg.Formatter
	if formatter == nil {
		formatter = api_formatter.GetGlobalFormatter()
	}

	clientResp := &api_formatter.ClientResponse{}
	if err := formatter.ParseClientResponse(resp, clientResp); err != nil {
		return zero, fmt.Errorf("failed to parse response: %v", err)
	}

	// If CustomFunc is provided, delegate all handling to it
	if cfg.CustomFunc != nil {
		customResult, err := cfg.CustomFunc(resp, clientResp)
		if err != nil {
			return zero, err
		}

		// If custom function returns nil, continue with default flow
		if customResult != nil {
			if result, ok := customResult.(T); ok {
				return result, nil
			}

			// Fallback: cast using reflection
			var result T
			resultType := reflect.TypeOf((*T)(nil)).Elem()

			if resultType.Kind() == reflect.Pointer {
				elemType := resultType.Elem()
				newValue := reflect.New(elemType)

				if err := cast.ToStruct(customResult, newValue.Interface(), true); err != nil {
					return zero, fmt.Errorf("failed to cast custom result: %v", err)
				}

				result = newValue.Interface().(T)
			} else {
				if err := cast.ToStruct(customResult, &result, true); err != nil {
					return zero, fmt.Errorf("failed to cast custom result: %v", err)
				}
			}

			return result, nil
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

		// Return ApiError based on HTTP status code to preserve error information
		return zero, &ApiError{
			StatusCode: resp.StatusCode,
			Code:       code,
			Message:    message,
		}
	}

	var result T

	// Check if T is interface{} or any - if so, return data directly
	resultType := reflect.TypeOf((*T)(nil)).Elem()

	// Special case: if T is interface{} or any, return clientResp.Data directly
	if resultType.Kind() == reflect.Interface && resultType.NumMethod() == 0 {
		// T is any/interface{}, return data as-is
		if data, ok := any(clientResp.Data).(T); ok {
			return data, nil
		}
		return zero, fmt.Errorf("failed to cast data to interface type")
	}

	// Check if T is a pointer type using reflection
	if resultType.Kind() == reflect.Pointer {
		elemType := resultType.Elem()
		newValue := reflect.New(elemType)

		if err := cast.ToStruct(clientResp.Data, newValue.Interface(), true); err != nil {
			return zero, fmt.Errorf("failed to cast data: %v", err)
		}

		result = newValue.Interface().(T)
	} else {
		if err := cast.ToStruct(clientResp.Data, &result, true); err != nil {
			return zero, fmt.Errorf("failed to cast data: %v", err)
		}
	}

	return result, nil
}
