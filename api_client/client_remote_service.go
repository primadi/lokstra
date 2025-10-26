package api_client

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
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
	client             *ClientRouter
	basePath           string                   // e.g., "/auth", "/users", "/orders"
	convention         string                   // e.g., "rest", "rpc" (default: "rest")
	resourceName       string                   // e.g., "user", "order"
	pluralResourceName string                   // e.g., "users", "orders"
	routeOverrides     map[string]string        // methodName -> custom path
	methodOverrides    map[string]string        // methodName -> HTTP method
	parser             *router.ConventionParser // Reuse server-side convention parser
}

// NewRemoteService creates a client for automatic remote service calls.
// Default convention is "rest".
func NewRemoteService(client *ClientRouter, basePath string) *RemoteService {
	rs := &RemoteService{
		client:          client,
		basePath:        basePath,
		convention:      "rest", // default convention
		routeOverrides:  make(map[string]string),
		methodOverrides: make(map[string]string),
	}
	// Initialize parser with empty resource names (will be set via With methods)
	rs.updateParser()
	return rs
}

// updateParser creates/updates the ConventionParser when resource names change
func (c *RemoteService) updateParser() {
	c.parser = router.NewConventionParser(c.resourceName, c.pluralResourceName)
}

// WithConvention sets the convention for path generation.
// Supported conventions: "rest", "kebab-case", "rpc"
func (c *RemoteService) WithConvention(convention string) *RemoteService {
	c.convention = convention
	return c
}

// WithResourceName sets the resource name (singular).
func (c *RemoteService) WithResourceName(resourceName string) *RemoteService {
	c.resourceName = resourceName
	c.updateParser()
	return c
}

// WithPluralResourceName sets the plural resource name override.
func (c *RemoteService) WithPluralResourceName(pluralResourceName string) *RemoteService {
	c.pluralResourceName = pluralResourceName
	c.updateParser()
	return c
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

// methodToHTTP converts method name to HTTP method and path using conventions.
// Detects:
//   - Path parameters from struct tags: `path:"dep"`, `path:"id"`
//   - Method override from struct tag: `method:"POST"`
//   - Custom path from struct tag: `route:"/custom/path"`
//   - Convention: REST (default), kebab-case, RPC
func (c *RemoteService) methodToHTTP(methodName string, req any) (httpMethod string, path string) {
	var pathParams []string
	var methodOverride string
	var pathOverride string

	// Step 1: Extract struct metadata (tags)
	if req != nil {
		pathParams, methodOverride, pathOverride = c.extractStructMetadata(req)
	}

	// Step 2: Check for configured route override (from auto-router config)
	if overridePath, exists := c.routeOverrides[methodName]; exists {
		pathOverride = overridePath
	}

	// Step 3: Check for configured method override (from auto-router config)
	if overrideMethod, exists := c.methodOverrides[methodName]; exists {
		methodOverride = overrideMethod
	}

	// Step 4: Check for path override (highest priority)
	if pathOverride != "" {
		path = c.applyBasePath(pathOverride)
		// Still need to determine HTTP method
		if methodOverride != "" {
			httpMethod = methodOverride
		} else {
			httpMethod = c.inferHTTPMethodFromMethodName(methodName, req)
		}
		return httpMethod, path
	}

	// Step 5: Check for method override
	if methodOverride != "" {
		httpMethod = methodOverride
	} else {
		httpMethod = c.inferHTTPMethodFromMethodName(methodName, req)
	}

	// Step 6: Build path using ConventionParser (REUSE SERVER LOGIC!)
	switch c.convention {
	case "rest":
		path = c.buildRESTPathUsingParser(methodName, pathParams)
	case "kebab-case":
		path = c.buildKebabPath(methodName)
	case "rpc":
		path = c.buildRPCPath(methodName)
	default:
		path = c.buildRESTPathUsingParser(methodName, pathParams)
	}

	return httpMethod, path
}

// buildRESTPathUsingParser uses the shared ConventionParser logic
func (c *RemoteService) buildRESTPathUsingParser(methodName string, pathParams []string) string {
	// Extract action from method name using parser
	action := c.parser.ExtractAction(methodName)
	if action == "" {
		// Fallback to simple path
		return c.basePath
	}

	var resourcePath string

	if len(pathParams) > 0 {
		// Has path params - need to generate path from struct
		// Since we don't have the actual struct type, we manually build the path
		// following the same logic as GeneratePathFromStruct
		resourcePath = c.buildPathWithParams(action, pathParams)
	} else {
		// No path params - use simple convention
		resourcePath = c.parser.GeneratePath(action)
	}

	return c.applyBasePath(resourcePath)
}

// buildPathWithParams builds path with parameters following GeneratePathFromStruct logic
func (c *RemoteService) buildPathWithParams(action string, pathParams []string) string {
	pluralName := c.pluralResourceName
	if pluralName == "" && c.resourceName != "" {
		pluralName = c.resourceName + "s"
	}

	if pluralName == "" {
		pluralName = "resources" // fallback
	}

	// Build path following GeneratePathFromStruct logic
	switch action {
	case "Get", "Update", "Replace", "Put", "Modify", "Patch", "Delete", "Remove":
		// Build path: /users/{dep}/{id}
		pathParts := make([]string, len(pathParams))
		for i, param := range pathParams {
			pathParts[i] = "{" + param + "}"
		}
		return "/" + pluralName + "/" + strings.Join(pathParts, "/")

	case "List", "Find", "Search", "Query":
		// List/Search with filters: /users/{dep}/search
		if len(pathParams) > 0 {
			pathParts := make([]string, len(pathParams))
			for i, param := range pathParams {
				pathParts[i] = "{" + param + "}"
			}
			suffix := ""
			if action == "Search" || action == "Query" {
				suffix = "/search"
			}
			return "/" + pluralName + "/" + strings.Join(pathParts, "/") + suffix
		}
		return "/" + pluralName

	case "Create", "Add", "Post":
		// Create with parent resource: /departments/{dep}/users
		if len(pathParams) > 0 {
			parentParams := pathParams
			if len(pathParams) > 1 {
				parentParams = pathParams[:len(pathParams)-1]
			}
			if len(parentParams) > 0 {
				pathParts := make([]string, len(parentParams))
				for i, param := range parentParams {
					pathParts[i] = "{" + param + "}"
				}
				return "/" + strings.Join(pathParts, "/") + "/" + pluralName
			}
		}
		return "/" + pluralName

	default:
		return "/" + pluralName
	}
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

// extractStructMetadata extracts path params, method override, and path override from struct tags.
// Returns:
//   - pathParams: []string - path parameter names from `path:"param"` tags
//   - methodOverride: string - HTTP method from `method:"POST"` tag on struct
//   - pathOverride: string - custom path from `route:"/custom/path"` tag on struct
func (c *RemoteService) extractStructMetadata(req any) (pathParams []string, methodOverride string, pathOverride string) {
	v := reflect.ValueOf(req)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, "", ""
	}

	t := v.Type()

	// Check struct-level tags (for method and route overrides)
	// Note: Go doesn't support struct-level tags directly, so we check the first field with special names
	// Alternative: use a special interface or method

	// Iterate through fields to find path tags
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Extract path parameters
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			pathParams = append(pathParams, pathTag)
		}

		// Check for method override (special field name convention)
		if field.Name == "_Method" || field.Name == "Method_" {
			if methodTag := field.Tag.Get("value"); methodTag != "" {
				methodOverride = methodTag
			}
		}

		// Check for route override (special field name convention)
		if field.Name == "_Route" || field.Name == "Route_" {
			if routeTag := field.Tag.Get("value"); routeTag != "" {
				pathOverride = routeTag
			}
		}
	}

	return pathParams, methodOverride, pathOverride
}

// inferHTTPMethodFromMethodName infers HTTP method from method name prefix.
// Uses the shared ConventionParser logic.
func (c *RemoteService) inferHTTPMethodFromMethodName(methodName string, req any) string {
	// Extract action using parser
	action := c.parser.ExtractAction(methodName)
	if action != "" {
		// Use parser to convert action to HTTP method
		return c.parser.ActionToHTTPMethod(action)
	}

	// Unknown action - check if has meaningful body (non-path fields)
	if c.hasBodyWithoutPathParams(req) {
		return "POST"
	}
	return "GET"
}

// buildKebabPath builds kebab-case style path (legacy).
// Example: ValidateToken → /auth/validate-token
func (c *RemoteService) buildKebabPath(methodName string) string {
	pathSegment := c.methodNameToPath(methodName)
	return c.applyBasePath(pathSegment)
}

// buildRPCPath builds RPC-style path.
// Example: ValidateToken → /auth/ValidateToken or /auth/validate_token
func (c *RemoteService) buildRPCPath(methodName string) string {
	// RPC style: use method name as-is or with underscores
	return c.applyBasePath("/" + camelToSnake(methodName))
}

// camelToSnake converts camelCase to snake_case.
// Example: ValidateToken → validate_token
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// methodNameToPath converts method name to URL path segment.
//
// Examples:
//   - Login → /login
//   - CreateUser → /users (creates resource)
//   - GetUser → /users/:id (get specific)
//   - ListUsers → /users (list all)
//   - ValidateToken → /validate-token
func (c *RemoteService) methodNameToPath(methodName string) string {
	// Remove common prefixes
	name := methodName
	for _, prefix := range []string{"Create", "Get", "Update", "Delete", "Add", "Remove", "List", "Find", "Process", "Generate"} {
		if after, ok := strings.CutPrefix(name, prefix); ok {
			name = after
			break
		}
	}

	// Convert to lowercase and add hyphens
	// Example: ValidateToken → validate-token
	path := camelToKebab(name)
	if path == "" {
		path = camelToKebab(methodName)
	}

	return "/" + path
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

// camelToKebab converts camelCase to kebab-case.
// Example: ValidateToken → validate-token
func camelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
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
