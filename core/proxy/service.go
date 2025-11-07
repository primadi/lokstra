package proxy

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/primadi/lokstra/api_client"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router/autogen"
	"github.com/primadi/lokstra/core/router/convention"
)

// Service represents a remote service proxy
type Service struct {
	client         *api_client.ClientRouter
	baseURL        string
	conversionRule autogen.ConversionRule
	routeOverride  autogen.RouteOverride
}

// NewService creates a new proxy service
func NewService(baseURL string, rule autogen.ConversionRule, override autogen.RouteOverride) *Service {
	client := &api_client.ClientRouter{
		FullURL: baseURL,
		IsLocal: false,
		Timeout: 30 * time.Second,
	}

	// Log remote service creation
	log.Printf("🌐 Created remote service proxy: %s (resource: %s)", baseURL, rule.Resource)

	return &Service{
		client:         client,
		baseURL:        baseURL,
		conversionRule: rule,
		routeOverride:  override,
	}
}

// Call invokes a remote service method with automatic HTTP request building
// Supports handler signatures that return error only:
//   - func() error
//   - func(*Context) error
//   - func(*Struct) error
//   - func(*Context, *Struct) error
//
// Returns error
func Call(s *Service, methodName string, params ...any) error {
	// Build HTTP request based on method name and conversion rule
	httpMethod, pathTemplate, err := s.resolveMethodToHTTP(methodName)
	if err != nil {
		return err
	}

	// Extract parameters
	var ctx *request.Context
	var structParam any

	for _, param := range params {
		if c, ok := param.(*request.Context); ok {
			ctx = c
		} else {
			structParam = param
		}
	}

	// Replace path parameters from context
	path := s.replacePathParameters(pathTemplate, ctx, structParam)

	log.Printf("🌐 proxy.Call: %s → %s %s", methodName, httpMethod, s.baseURL+path)

	// Build request options
	opts := s.buildRequestOptions(httpMethod, structParam, ctx)

	// Make HTTP call - use empty response type for error-only handlers
	_, err = api_client.FetchAndCast[any](s.client, path, opts...)
	if err != nil {
		log.Printf("❌ proxy.Call error: %v", err)
		return err
	}

	log.Printf("✅ proxy.Call success")
	return nil
}

// CallWithData invokes a remote service method and returns typed data
// Supports handler signatures that return data:
//   - func() (T, error) or func() any
//   - func(*Context) (T, error) or func(*Context) any
//   - func(*Struct) (T, error) or func(*Struct) any
//   - func(*Context, *Struct) (T, error) or func(*Context, *Struct) any
//   - func() (*Response, error) or func() *Response
//   - func(*Context) (*Response, error) or func(*Context) *Response
//   - func(*Struct) (*Response, error) or func(*Struct) *Response
//   - func(*Context, *Struct) (*Response, error) or func(*Context, *Struct) *Response
//   - func() (*ApiHelper, error) or func() *ApiHelper
//   - func(*Context) (*ApiHelper, error) or func(*Context) *ApiHelper
//   - func(*Struct) (*ApiHelper, error) or func(*Struct) *ApiHelper
//   - func(*Context, *Struct) (*ApiHelper, error) or func(*Context, *Struct) *ApiHelper
//
// Returns (T, error)
// T will be extracted from:
//   - response.Data field for standard API responses
//   - Full Response struct for *Response returns
//   - Full ApiHelper for *ApiHelper returns
func CallWithData[T any](s *Service, methodName string, params ...any) (T, error) {
	var zero T

	// Build HTTP request based on method name and conversion rule
	httpMethod, pathTemplate, err := s.resolveMethodToHTTP(methodName)
	if err != nil {
		return zero, err
	}

	// Extract parameters
	var ctx *request.Context
	var structParam any

	for _, param := range params {
		if c, ok := param.(*request.Context); ok {
			ctx = c
		} else {
			structParam = param
		}
	}

	// Replace path parameters from context
	path := s.replacePathParameters(pathTemplate, ctx, structParam)

	log.Printf("🌐 proxy.CallWithData: %s → %s %s", methodName, httpMethod, s.baseURL+path)

	// Build request options
	opts := s.buildRequestOptions(httpMethod, structParam, ctx)

	// Make HTTP call and get typed response
	// FetchAndCast will automatically extract the Data field from standard API responses
	data, err := api_client.FetchAndCast[T](s.client, path, opts...)
	if err != nil {
		log.Printf("❌ proxy.CallWithData error: %v", err)
		return zero, err
	}

	log.Printf("✅ proxy.CallWithData success: %T", data)
	return data, nil
}

// resolveMethodToHTTP converts a method name to HTTP method and path
// using the conversion rule and route override
// Returns (httpMethod, path, error)
//
// Resolution order (same as autogen.NewFromService):
//  1. Check if method is hidden → return error
//  2. Check for custom route override → use custom route
//  3. Resolve using convention → use convention-based route
//  4. Method not found → return error
func (s *Service) resolveMethodToHTTP(methodName string) (httpMethod string, path string, err error) {
	// Step 1: Check if hidden
	for _, hiddenMethod := range s.routeOverride.Hidden {
		if hiddenMethod == methodName {
			return "", "", fmt.Errorf("method %s is hidden", methodName)
		}
	}

	// Step 2: Check for custom route override
	if customRoute, ok := s.routeOverride.Custom[methodName]; ok {
		path = customRoute.Path
		if s.routeOverride.PathPrefix != "" {
			path = s.routeOverride.PathPrefix + path
		}
		return customRoute.Method, path, nil
	}

	// Step 3: Use convention registry to resolve method
	conv := convention.MustGet(s.conversionRule.Convention)

	// Calculate resource plural
	resourcePlural := s.conversionRule.ResourcePlural
	if resourcePlural == "" {
		resourcePlural = s.conversionRule.Resource + "s"
	}

	// Resolve method to HTTP method and path template
	httpMethod, pathTemplate, found := conv.ResolveMethod(
		methodName,
		s.conversionRule.Resource,
		resourcePlural,
	)

	if !found {
		return "", "", fmt.Errorf("method %s not found in convention %s", methodName, s.conversionRule.Convention)
	}

	// Apply path prefix if specified
	path = pathTemplate
	if s.routeOverride.PathPrefix != "" {
		path = s.routeOverride.PathPrefix + path
	}

	return httpMethod, path, nil
}

// replacePathParameters replaces path parameter placeholders with actual values
// Priority: 1. struct path tags, 2. context path params
func (s *Service) replacePathParameters(pathTemplate string, ctx *request.Context, structParam any) string {
	path := pathTemplate
	replacements := make(map[string]string)

	// Step 1: Extract values from struct path tags
	if structParam != nil {
		val := reflect.ValueOf(structParam)
		if val.Kind() == reflect.Pointer {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			typ := val.Type()
			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i)
				fieldValue := val.Field(i)

				// Check for path tag
				if pathTag := field.Tag.Get("path"); pathTag != "" {
					replacements[pathTag] = fmt.Sprintf("%v", fieldValue.Interface())
				}
			}
		}
	}

	// Step 2: Apply replacements from struct
	for paramName, paramValue := range replacements {
		path = replacePathParam(path, paramName, paramValue)
	}

	// Step 3: Check if there are still unresolved parameters and get them from context
	if ctx != nil && strings.Contains(path, "{") {
		// Extract all remaining placeholders
		unresolvedParams := extractPathPlaceholders(path)
		for _, paramName := range unresolvedParams {
			paramValue := ctx.Req.PathParam(paramName, "")
			if paramValue != "" {
				path = replacePathParam(path, paramName, paramValue)
			}
		}
	}

	return path
}

// extractPathPlaceholders extracts all {paramName} placeholders from a path
func extractPathPlaceholders(path string) []string {
	var params []string
	start := -1
	for i, ch := range path {
		if ch == '{' {
			start = i
		} else if ch == '}' && start >= 0 {
			paramName := path[start+1 : i]
			params = append(params, paramName)
			start = -1
		}
	}
	return params
}

// replacePathParam replaces {paramName} with value in path
func replacePathParam(path, paramName, value string) string {
	placeholder := "{" + paramName + "}"
	return strings.ReplaceAll(path, placeholder, value)
}

// buildRequestOptions builds fetch options based on HTTP method and parameters
func (s *Service) buildRequestOptions(httpMethod string, structParam any, ctx *request.Context) []api_client.FetchOption {
	var opts []api_client.FetchOption

	// Set HTTP method
	opts = append(opts, api_client.WithMethod(httpMethod))

	// Handle struct parameter
	if structParam != nil {
		// Analyze struct to extract path params, query params, and body
		opts = append(opts, s.extractParamsFromStruct(structParam, httpMethod)...)
	}

	// Copy headers from context if available
	if ctx != nil && ctx.R != nil {
		headers := make(map[string]string)
		for key, values := range ctx.R.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		if len(headers) > 0 {
			opts = append(opts, api_client.WithHeaders(headers))
		}
	}

	return opts
}

// extractParamsFromStruct extracts path params, query params, and body from struct
func (s *Service) extractParamsFromStruct(structParam any, httpMethod string) []api_client.FetchOption {
	var opts []api_client.FetchOption

	val := reflect.ValueOf(structParam)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return opts
	}

	typ := val.Type()
	var bodyData any

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Check for json tag (body)
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// For POST/PUT/PATCH, send as body
			if httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH" {
				if bodyData == nil {
					bodyData = make(map[string]any)
				}
				bodyData.(map[string]any)[jsonTag] = fieldValue.Interface()
			}
		}
	}

	// Add body data
	if bodyData != nil {
		opts = append(opts, api_client.WithBody(bodyData))
	}

	return opts
}
