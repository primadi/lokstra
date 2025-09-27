package route

import "github.com/primadi/lokstra/core/request"

type Route struct {
	Name             string
	Description      string
	Method           string
	Path             string
	Handler          request.HandlerFunc
	Middleware       []request.HandlerFunc
	OverrideParentMw bool

	// populated during Build()
	FullPath       string
	FullName       string
	FullMiddleware []request.HandlerFunc
}

type RouteHandlerOption interface {
	Apply(*Route)
}
