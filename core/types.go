package core

import (
	"context"
	"lokstra/common/response"
	"lokstra/iface"
	"net/http"
)

// RequestContext is the context passed to each handler.
// It contains the HTTP request, response writer, app reference, and shared values.
type RequestContext struct {
	context.Context    // Inherits standard Go context for cancellation and deadlines.
	*response.Response // Embedded response helper for structured API output and fluent chaining (e.g., ctx.WithMessage(...).Ok(...))

	App     iface.App           // Reference to the app instance serving this request.
	Writer  http.ResponseWriter // HTTP response writer.
	Request *http.Request       // Original HTTP request.
	values  map[string]any      // Arbitrary data shared across middleware and handlers.
}

// RequestHandler defines the standard function signature for route handlers in Lokstra.
// It receives a RequestContext and returns an error if the request fails.
type RequestHandler func(ctx *RequestContext) error

// MiddlewareHandler wraps a RequestHandler, allowing for request preprocessing or postprocessing.
type MiddlewareHandler func(RequestHandler) RequestHandler

// MiddlewareFactory defines a function that creates a MiddlewareHandler from config parameters.
type MiddlewareFactory func(params MiddlewareConfig) (MiddlewareHandler, error)

// MiddlewareConfig provides a flexible map-based configuration for middleware.
type MiddlewareConfig = map[string]any

// HTTPMethod defines supported HTTP method strings.
type HTTPMethod string

const (
	GET    HTTPMethod = "GET"
	POST   HTTPMethod = "POST"
	PUT    HTTPMethod = "PUT"
	DELETE HTTPMethod = "DELETE"
	PATCH  HTTPMethod = "PATCH"
)

// RouteInfo stores metadata for a registered route in Lokstra.
type RouteInfo struct {
	FullPath   string              // Complete route path, including group prefixes (e.g. "/api/v1/users").
	Method     HTTPMethod          // HTTP method for this route (GET, POST, etc).
	Handler    RequestHandler      // The main request handler for the route.
	Middleware []MiddlewareHandler // List of middleware applied to this route, in order.
}
