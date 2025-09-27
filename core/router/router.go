package router

import (
	"net/http"

	"github.com/primadi/lokstra/core/route"
)

type Router interface {
	http.Handler

	// Router Name for identification
	Name() string
	// EngineType returns the underlying engine type, e.g. "default", "servemux", etc.
	EngineType() string
	// PathPrefix returns the path prefix of this router
	PathPrefix() string

	// route registration for GET method
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//
	// middleware can be:
	//  - [same as above]
	//  - route.HandlerOption (e.g. route.WithNameOption,
	//    route.WithDescriptionOption, route.WithOverrideParentMwOption)
	GET(path string, h any, middleware ...any) Router
	// route registration for POST method
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	POST(path string, h any, middleware ...any) Router
	// route registration for PUT method
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	PUT(path string, h any, middleware ...any) Router
	// route registration for DELETE metod
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	DELETE(path string, h any, middleware ...any) Router
	// route registration for PATCH method
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	PATCH(path string, h any, middleware ...any) Router
	// route registration for ANY method (all methods)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	ANY(path string, h any, middleware ...any) Router

	// route registration for GET method with prefix match
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	GETPrefix(prefix string, h any, middleware ...any) Router
	// route registration for POST method with prefix match
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, *T) error
	//  - route.HandlerOption
	POSTPrefix(prefix string, h any, middleware ...any) Router
	// route registration for PUT method with prefix match
	PUTPrefix(prefix string, h any, middleware ...any) Router
	// route registration for DELETE method with prefix match
	DELETEPrefix(prefix string, h any, middleware ...any) Router
	// route registration for PATCH method with prefix match
	PATCHPrefix(prefix string, h any, middleware ...any) Router
	// route registration for ANY method with prefix match
	ANYPrefix(prefix string, h any, middleware ...any) Router

	// create a sub- router with prefix, and call the fn to register routes on it
	// e.g. r.Group("/v1", func(g lokstra.Router) { ... })
	Group(prefix string, fn func(r Router)) Router
	// create a sub- router with prefix, and return it for further route registration
	// e.g. gv2 := r.AddGroup("/v2")
	AddGroup(prefix string) Router

	// add global middleware(s) to this router
	// e.g. r.Use(middleware...)
	Use(middleware ...any) Router

	// set whether this router should override parent middleware when adding routes
	WithOverrideParentMiddleware(override bool) Router

	// walk through all routes (including in child groups) and call fn for each route
	// fullPath is the complete path including all parent group prefixes
	// e.g. /v1/admin/stats
	Walk(fn func(rt *route.Route))
	// Print all routes to stdout for introspection
	PrintRoutes()

	// finalize the router and its children, building the underlying engine
	Build()
	// check if the router has been built
	IsBuilt() bool

	// check if the router is part of a chain
	IsChained() bool
	// get the next router in the chain, or nil if none
	GetNextChain() Router
}
