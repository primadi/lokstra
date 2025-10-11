package router

import (
	"fmt"
	"reflect"
	"strings"
)

// RESTConvention implements standard REST-like service-to-router mapping
type RESTConvention struct{}

// Name returns the convention name
func (c *RESTConvention) Name() string {
	return "rest"
}

// GenerateRoutes generates REST-like routes from service methods
func (c *RESTConvention) GenerateRoutes(serviceType reflect.Type, options ServiceRouterOptions) (map[string]RouteMeta, error) {
	routes := make(map[string]RouteMeta)

	// Determine resource name
	resourceName := options.ResourceName
	if resourceName == "" {
		resourceName = extractResourceName(serviceType.Name())
	}

	pluralName := options.PluralResourceName
	if pluralName == "" {
		pluralName = pluralize(resourceName)
	}

	// Iterate through all methods
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)
		methodName := method.Name

		// Check if there's an override
		if override, exists := options.RouteOverrides[methodName]; exists {
			routes[methodName] = override
			continue
		}

		// Skip if conventions are disabled
		if options.DisableConventions {
			continue
		}

		// Generate route from method name
		routeMeta := c.generateRouteFromMethodName(methodName, resourceName, pluralName)
		if routeMeta.HTTPMethod != "" { // Valid route generated
			routes[methodName] = routeMeta
		}
	}

	return routes, nil
}

// GenerateClientMethod generates client method metadata
func (c *RESTConvention) GenerateClientMethod(method ServiceMethodInfo, options ServiceRouterOptions) (ClientMethodMeta, error) {
	resourceName := options.ResourceName
	if resourceName == "" {
		// Try to extract from method name or use default
		resourceName = "resource"
	}

	pluralName := options.PluralResourceName
	if pluralName == "" {
		pluralName = pluralize(resourceName)
	}

	routeMeta := c.generateRouteFromMethodName(method.Name, resourceName, pluralName)

	clientMeta := ClientMethodMeta{
		MethodName: method.Name,
		HTTPMethod: routeMeta.HTTPMethod,
		Path:       routeMeta.Path,
		PathParams: extractPathParams(routeMeta.Path),
		HasBody:    isBodyMethod(routeMeta.HTTPMethod),
	}

	return clientMeta, nil
}

// generateRouteFromMethodName generates route metadata from method name
func (c *RESTConvention) generateRouteFromMethodName(methodName, resourceName, pluralName string) RouteMeta {
	meta := RouteMeta{
		MethodName: methodName,
	}

	_ = resourceName // Currently unused, but can be used for more complex conventions

	// REST convention patterns:
	// Get{Resource} -> GET /{resources}/{id}
	// List{Resource}s -> GET /{resources}
	// Create{Resource} -> POST /{resources}
	// Update{Resource} -> PUT /{resources}/{id}
	// Delete{Resource} -> DELETE /{resources}/{id}
	// Patch{Resource} -> PATCH /{resources}/{id}

	lower := strings.ToLower(methodName)

	switch {
	case strings.HasPrefix(lower, "get"):
		meta.HTTPMethod = "GET"
		meta.Path = fmt.Sprintf("/%s/{id}", pluralName)
	case strings.HasPrefix(lower, "list"):
		meta.HTTPMethod = "GET"
		meta.Path = fmt.Sprintf("/%s", pluralName)
	case strings.HasPrefix(lower, "create"):
		meta.HTTPMethod = "POST"
		meta.Path = fmt.Sprintf("/%s", pluralName)
	case strings.HasPrefix(lower, "update"):
		meta.HTTPMethod = "PUT"
		meta.Path = fmt.Sprintf("/%s/{id}", pluralName)
	case strings.HasPrefix(lower, "delete"):
		meta.HTTPMethod = "DELETE"
		meta.Path = fmt.Sprintf("/%s/{id}", pluralName)
	case strings.HasPrefix(lower, "patch"):
		meta.HTTPMethod = "PATCH"
		meta.Path = fmt.Sprintf("/%s/{id}", pluralName)
	default:
		// Default to POST with method name as path
		meta.HTTPMethod = "POST"
		meta.Path = fmt.Sprintf("/%s/%s", pluralName, camelToKebab(methodName))
	}

	return meta
}

// extractResourceName extracts resource name from service name
// Example: "UserService" -> "user"
func extractResourceName(serviceName string) string {
	name := strings.TrimSuffix(serviceName, "Service")
	name = strings.TrimSuffix(name, "API")
	return strings.ToLower(name)
}

// pluralize adds 's' to make plural (simple implementation)
func pluralize(word string) string {
	if word == "" {
		return word
	}
	// Simple pluralization (can be improved)
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "ch") {
		return word + "es"
	}
	if strings.HasSuffix(word, "y") {
		return word[:len(word)-1] + "ies"
	}
	return word + "s"
}

// camelToKebab converts CamelCase to kebab-case
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

// extractPathParams extracts parameter names from path
// Example: "/users/{id}/posts/{postId}" -> ["id", "postId"]
func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			param := strings.TrimPrefix(part, "{")
			param = strings.TrimSuffix(param, "}")
			params = append(params, param)
		}
	}
	return params
}

// isBodyMethod returns true if the HTTP method typically has a body
func isBodyMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "POST", "PUT", "PATCH":
		return true
	default:
		return false
	}
}

// Register the REST convention during init
func init() {
	MustRegisterConvention(&RESTConvention{})
}
