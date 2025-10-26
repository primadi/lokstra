package route

import "github.com/primadi/lokstra/core/request"

type Route struct {
	Name             string
	Description      string
	Method           string
	Path             string
	Handler          request.HandlerFunc
	Middleware       []any // Mixed: request.HandlerFunc or string (lazy)
	OverrideParentMw bool

	// populated during Build()
	RouterName     string // Name of the router this route belongs to
	FullPath       string
	FullName       string
	FullMiddleware []request.HandlerFunc
}

type RouteHandlerOption interface {
	Apply(*Route)
}
