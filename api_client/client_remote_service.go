package api_client

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

// ==============================================================================
// Remote Service Client
// Maps interface method calls to HTTP endpoints automatically
// ==============================================================================

// RemoteService provides automatic HTTP mapping for service method calls.
// It converts Go method calls to REST API calls using naming conventions.
//
// Usage:
//
//	client := api_client.NewRemoteService(clientRouter, "/auth")
//	response := api_client.CallRemoteService[LoginResponse](client, "Login", ctx, req)
//
// Method Mapping Rules:
//   - Create*, Add*, Process*, Generate*, Login*, Register* → POST
//   - Update*, Modify* → PUT
//   - Delete*, Remove* → DELETE
//   - Get*, Find*, List*, Validate* → GET (or POST with body)
//   - Method names convert to kebab-case paths: ValidateToken → /validate-token
//
// Strategy Detection:
//   - Detects path tags in request struct: `path:"dep"`, `path:"id"`
//   - Detects strategy tags: `method:"POST"`, `path:"/custom/path"`
//   - Supports multiple strategies: REST, custom paths, method overrides
type RemoteService struct {
	client          *ClientRouter
	basePath        string            // e.g., "/auth", "/users", "/orders"
	routeOverrides  map[string]string // methodName -> custom path
	methodOverrides map[string]string // methodName -> HTTP method
}

// NewRemoteService creates a client for remote service calls.
// Routes must be explicitly configured via WithRouteOverride.
func NewRemoteService(client *ClientRouter, basePath string) *RemoteService {
	return &RemoteService{
		client:          client,
		basePath:        basePath,
		routeOverrides:  make(map[string]string),
		methodOverrides: make(map[string]string),
	}
}

// WithRouteOverride adds a route path override for a specific method.
func (c *RemoteService) WithRouteOverride(methodName, path string) *RemoteService {
	c.routeOverrides[methodName] = path
	return c
}

// WithMethodOverride adds an HTTP method override for a specific method.
func (c *RemoteService) WithMethodOverride(methodName, httpMethod string) *RemoteService {
	c.methodOverrides[methodName] = httpMethod
	return c
}

// CallRemoteService makes a type-safe remote service call using method name conventions.
//
// Example:
//
//	response, err := api_client.CallRemoteService[LoginResponse](client, "Login", ctx, req)
//	// → POST /auth/login with req as body
//
// Method naming determines HTTP method and path:
//   - Login(req) → POST /auth/login
//   - GetUser(req) → GET /users/get-user (or POST if req has body)
//   - ValidateToken(req) → POST /auth/validate-token
func CallRemoteService[TResponse any](c *RemoteService, methodName string, ctx *request.Context, req any) (TResponse, error) {
	httpMethod, path := c.methodToHTTP(methodName, req)

	// Substitute path parameters with actual values from req
	path = c.substitutePathParameters(path, req)

	return FetchAndCast[TResponse](c.client, path,
		WithMethod(httpMethod),
		WithBody(req),
	)
}

// Call is the untyped version of CallRemoteService (returns any).
// Prefer CallRemoteService for type safety.
func (c *RemoteService) Call(methodName string, ctx *request.Context, req any) (any, error) {
	httpMethod, path := c.methodToHTTP(methodName, req)

	// Substitute path parameters with actual values from req
	path = c.substitutePathParameters(path, req)

	return FetchAndCast[any](c.client, path,
		WithMethod(httpMethod),
		WithBody(req),
	)
}

// methodToHTTP converts method name to HTTP method and path.
// REQUIRES explicit route override via WithRouteOverride.
// Returns error path if no override found.
func (c *RemoteService) methodToHTTP(methodName string, req any) (httpMethod string, path string) {
	// Check for configured route override (REQUIRED)
	path, exists := c.routeOverrides[methodName]
	if !exists {
		// No route override - return base path as fallback
		path = c.basePath
	}

	// Apply base path if path is relative
	if !strings.HasPrefix(path, "/") {
		path = c.applyBasePath(path)
	}

	// Check for method override
	httpMethod, exists = c.methodOverrides[methodName]
	if !exists {
		// Default to POST if has body, GET otherwise
		if c.hasBodyWithoutPathParams(req) {
			httpMethod = "POST"
		} else {
			httpMethod = "GET"
		}
	}

	return httpMethod, path
}

// applyBasePath prepends basePath to a resource path
func (c *RemoteService) applyBasePath(resourcePath string) string {
	if c.basePath == "" {
		return resourcePath
	}
	// Ensure proper path concatenation
	base := strings.TrimSuffix(c.basePath, "/")
	resource := strings.TrimPrefix(resourcePath, "/")
	if resource == "" {
		return base
	}
	return base + "/" + resource
}

// hasBodyWithoutPathParams checks if request has meaningful body fields
// excluding path parameter fields (those with `path:"..."` tag).
func (c *RemoteService) hasBodyWithoutPathParams(req any) bool {
	if req == nil {
		return false
	}

	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	// If not a struct, assume no body
	if v.Kind() != reflect.Struct {
		return false
	}

	t := v.Type()
	hasNonPathFields := false

	// Check if has any exported fields that are NOT path parameters
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip path parameter fields
		if field.Tag.Get("path") != "" {
			continue
		}

		// Found a non-path field
		hasNonPathFields = true
		break
	}

	return hasNonPathFields
}

// substitutePathParameters replaces path parameter placeholders with actual values from req.
// Example: /api/v1/users/{id} + req.ID=123 → /api/v1/users/123
//
// Path parameters are identified by struct tags: `path:"id"`
// The function extracts values from these fields and substitutes them into the path.
func (c *RemoteService) substitutePathParameters(path string, req any) string {
	if req == nil || !strings.Contains(path, "{") {
		return path
	}

	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return path
	}

	t := v.Type()
	replacements := make(map[string]string)

	// Extract all path parameter values from struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		pathTag := field.Tag.Get("path")

		if pathTag != "" {
			// Get field value
			fieldValue := v.Field(i)

			// Convert field value to string
			var valueStr string
			switch fieldValue.Kind() {
			case reflect.String:
				valueStr = fieldValue.String()
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				valueStr = strings.TrimSpace(fmt.Sprintf("%d", fieldValue.Int()))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				valueStr = strings.TrimSpace(fmt.Sprintf("%d", fieldValue.Uint()))
			case reflect.Float32, reflect.Float64:
				valueStr = strings.TrimSpace(fmt.Sprintf("%g", fieldValue.Float()))
			case reflect.Bool:
				valueStr = strings.TrimSpace(fmt.Sprintf("%t", fieldValue.Bool()))
			default:
				// Try to convert to string using fmt
				valueStr = strings.TrimSpace(fmt.Sprintf("%v", fieldValue.Interface()))
			}

			// Store replacement: {id} → 123
			replacements["{"+pathTag+"}"] = valueStr
		}
	}

	// Apply all replacements to path
	result := path
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// ==============================================================================
// Example Usage:
// ==============================================================================
//
// type authServiceRemote struct {
//     client *api_client.ClientRemoteService
// }
//
// func (s *authServiceRemote) Login(ctx *request.Context, req *LoginRequest) (*LoginResponse, error) {
//     return api_client.CallRemoteService[LoginResponse](s.client, "Login", ctx, req)
// }
//
// func (s *authServiceRemote) ValidateToken(ctx *request.Context, req *ValidateTokenRequest) (*ValidateTokenResponse, error) {
//     return api_client.CallRemoteService[ValidateTokenResponse](s.client, "ValidateToken", ctx, req)
// }
//
// func CreateAuthServiceRemote(cfg map[string]any) any {
//     routerName := utils.GetValueFromMap(cfg, "router", "auth-service")
//     pathPrefix := utils.GetValueFromMap(cfg, "path-prefix", "/auth")
//
//     clientRouter := lokstra_registry.GetClientRouter(routerName)
//     client := api_client.NewClientRemoteService(clientRouter, pathPrefix)
//
//     return &authServiceRemote{
//         client: client,
//     }
// }
