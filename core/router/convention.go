package router

import (
	"fmt"
	"reflect"
	"strings"
)

// ConventionParser parses method names using conventions
type ConventionParser struct {
	resourceName       string
	pluralResourceName string
}

// NewConventionParser creates a new parser
func NewConventionParser(resourceName, pluralResourceName string) *ConventionParser {
	return &ConventionParser{
		resourceName:       resourceName,
		pluralResourceName: pluralResourceName,
	}
}

// ParseMethodName parses a method name and returns HTTP method and path
//
// Convention Rules:
//
// GET methods:
//
//	Get{Resource}      → GET /{resource}s/{id}
//	List{Resource}s    → GET /{resource}s
//	Find{Resource}s    → GET /{resource}s
//	Search{Resource}s  → GET /{resource}s/search
//
// POST methods:
//
//	Create{Resource}   → POST /{resource}s
//	Add{Resource}      → POST /{resource}s
//
// PUT methods:
//
//	Update{Resource}   → PUT /{resource}s/{id}
//	Replace{Resource}  → PUT /{resource}s/{id}
//
// PATCH methods:
//
//	Modify{Resource}   → PATCH /{resource}s/{id}
//	Patch{Resource}    → PATCH /{resource}s/{id}
//
// DELETE methods:
//
//	Delete{Resource}   → DELETE /{resource}s/{id}
//	Remove{Resource}   → DELETE /{resource}s/{id}
func (p *ConventionParser) ParseMethodName(methodName string) (httpMethod, path string, err error) {
	// Extract action prefix
	action := p.extractAction(methodName)
	if action == "" {
		return "", "", fmt.Errorf("cannot determine HTTP method from method name: %s", methodName)
	}

	// Determine HTTP method from action
	httpMethod = p.actionToHTTPMethod(action)
	if httpMethod == "" {
		return "", "", fmt.Errorf("unknown action prefix: %s", action)
	}

	// Generate simple path (for methods without struct parameters)
	path = p.generatePath(action)

	return httpMethod, path, nil
}

// ExtractAction extracts the action prefix from method name (exported version)
// Example: "GetUser" → "Get", "ListUsers" → "List"
func (p *ConventionParser) ExtractAction(methodName string) string {
	return p.extractAction(methodName)
}

// extractAction extracts the action prefix from method name
// Example: "GetUser" → "Get", "ListUsers" → "List"
func (p *ConventionParser) extractAction(methodName string) string {
	actions := []string{
		// GET
		"Get", "List", "Find", "Search", "Query",
		// POST
		"Create", "Add", "Post",
		// PUT
		"Update", "Replace", "Put",
		// PATCH
		"Modify", "Patch",
		// DELETE
		"Delete", "Remove",
	}

	for _, action := range actions {
		if strings.HasPrefix(methodName, action) {
			return action
		}
	}

	return ""
}

// ActionToHTTPMethod converts action prefix to HTTP method (exported version)
func (p *ConventionParser) ActionToHTTPMethod(action string) string {
	return p.actionToHTTPMethod(action)
}

// actionToHTTPMethod converts action prefix to HTTP method
func (p *ConventionParser) actionToHTTPMethod(action string) string {
	switch action {
	case "Get", "List", "Find", "Search", "Query":
		return "GET"
	case "Create", "Add", "Post":
		return "POST"
	case "Update", "Replace", "Put":
		return "PUT"
	case "Modify", "Patch":
		return "PATCH"
	case "Delete", "Remove":
		return "DELETE"
	default:
		return ""
	}
}

// GeneratePath generates URL path based on action (exported version)
// Used only for simple methods without struct parameters
func (p *ConventionParser) GeneratePath(action string) string {
	return p.generatePath(action)
}

// generatePath generates URL path based on action
// Used only for simple methods without struct parameters
func (p *ConventionParser) generatePath(action string) string {
	plural := p.pluralResourceName
	if plural == "" {
		plural = p.resourceName + "s"
	}

	switch action {
	case "Get":
		// GetUser → /users/{id}
		return fmt.Sprintf("/%s/{id}", plural)

	case "List", "Find":
		// ListUsers → /users
		return fmt.Sprintf("/%s", plural)

	case "Search", "Query":
		// SearchUsers → /users/search
		return fmt.Sprintf("/%s/search", plural)

	case "Create", "Add", "Post":
		// CreateUser → /users
		return fmt.Sprintf("/%s", plural)

	case "Update", "Replace", "Put", "Modify", "Patch":
		// UpdateUser → /users/{id}
		return fmt.Sprintf("/%s/{id}", plural)

	case "Delete", "Remove":
		// DeleteUser → /users/{id}
		return fmt.Sprintf("/%s/{id}", plural)

	default:
		// Fallback: use plural resource
		return fmt.Sprintf("/%s", plural)
	}
}

// ExtractPathParamsFromStruct extracts path parameter names from struct tags
// Returns path param names in order they appear in the struct
//
// Example:
//
//	type GetUserRequest struct {
//	    DepartmentID string `path:"dep"`
//	    UserID       string `path:"id"`
//	    Query        string `query:"q"`
//	}
//	ExtractPathParamsFromStruct(GetUserRequest) → ["dep", "id"]
func ExtractPathParamsFromStruct(structType reflect.Type) []string {
	// Handle pointer to struct
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	// Must be a struct
	if structType.Kind() != reflect.Struct {
		return nil
	}

	var pathParams []string

	// Iterate through all fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Check for path tag
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			pathParams = append(pathParams, pathTag)
		}
	}

	return pathParams
}

// GeneratePathFromStruct generates URL path with parameters from struct tags
// Returns path like: /users/{dep}/{id} based on struct tags
//
// Example:
//
//	type GetUserRequest struct {
//	    DepartmentID string `path:"dep"`
//	    UserID       string `path:"id"`
//	}
//	GeneratePathFromStruct("Get", GetUserRequest, "users") → /users/{dep}/{id}
func (p *ConventionParser) GeneratePathFromStruct(action string, structType reflect.Type) string {
	pathParams := ExtractPathParamsFromStruct(structType)

	plural := p.pluralResourceName
	if plural == "" {
		plural = p.resourceName + "s"
	}

	// No path params - use standard convention
	if len(pathParams) == 0 {
		return p.generatePath(action)
	}

	// Build path with struct tag names
	// Example: ["dep", "id"] → /{plural}/{dep}/{id}
	switch action {
	case "Get", "Update", "Replace", "Put", "Modify", "Patch", "Delete", "Remove":
		// Build path: /users/{dep}/{id}
		pathParts := make([]string, len(pathParams))
		for i, param := range pathParams {
			pathParts[i] = fmt.Sprintf("{%s}", param)
		}
		return fmt.Sprintf("/%s/%s", plural, strings.Join(pathParts, "/"))

	case "List", "Find", "Search", "Query":
		// List/Search with filters: /users/{dep}/search
		if len(pathParams) > 0 {
			pathParts := make([]string, len(pathParams))
			for i, param := range pathParams {
				pathParts[i] = fmt.Sprintf("{%s}", param)
			}
			suffix := ""
			if action == "Search" || action == "Query" {
				suffix = "/search"
			}
			return fmt.Sprintf("/%s/%s%s", plural, strings.Join(pathParts, "/"), suffix)
		}
		return p.generatePath(action)

	case "Create", "Add", "Post":
		// Create with parent resource: /departments/{dep}/users
		if len(pathParams) > 0 {
			// Remove last param (that's for the resource being created)
			parentParams := pathParams
			if len(pathParams) > 1 {
				parentParams = pathParams[:len(pathParams)-1]
			}
			if len(parentParams) > 0 {
				pathParts := make([]string, len(parentParams))
				for i, param := range parentParams {
					pathParts[i] = fmt.Sprintf("{%s}", param)
				}
				return fmt.Sprintf("/%s/%s", strings.Join(pathParts, "/"), plural)
			}
		}
		return fmt.Sprintf("/%s", plural)

	default:
		// Fallback
		return p.generatePath(action)
	}
}
