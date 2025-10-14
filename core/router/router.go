package router

import (
	"net/http"

	"github.com/primadi/lokstra/core/route"
)

type Router interface {
	http.Handler

	// Router Name for identification
	Name() string
	// returns the underlying engine type, e.g. "default", "servemux", etc.
	EngineType() string
	// returns the path prefix of this router
	PathPrefix() string
	// sets the path prefix of this router
	SetPathPrefix(prefix string) Router
	// Create a shallow copy of this router (without routes and children)
	Clone() Router

	// route registration for GET method
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	GET(path string, h any, middleware ...any) Router
	// route registration for POST method
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	POST(path string, h any, middleware ...any) Router
	// route registration for PUT method
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	PUT(path string, h any, middleware ...any) Router
	// route registration for DELETE metod
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	DELETE(path string, h any, middleware ...any) Router
	// route registration for PATCH method
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	PATCH(path string, h any, middleware ...any) Router
	// route registration for ANY method (all methods)
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	ANY(path string, h any, middleware ...any) Router

	// route registration for GET method with prefix match
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	GETPrefix(prefix string, h any, middleware ...any) Router
	// route registration for POST method with prefix match
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	POSTPrefix(prefix string, h any, middleware ...any) Router
	// route registration for PUT method with prefix match
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	PUTPrefix(prefix string, h any, middleware ...any) Router
	// route registration for DELETE method with prefix match
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	DELETEPrefix(prefix string, h any, middleware ...any) Router
	// route registration for PATCH method with prefix match
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
	PATCHPrefix(prefix string, h any, middleware ...any) Router
	// route registration for ANY method with prefix match
	//
	// h can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - http.HandlerFunc
	//  - http.Handler
	//  - func(*lokstra.RequestContext, any) error
	//  - func() error
	//  - func() (any, error)
	//  - func(any) error
	//  - func(any) (any, error)
	//
	// middleware can be:
	//  - func(*lokstra.RequestContext) error
	//  - request.HandlerFunc
	//  - func(*lokstra.RequestContext, any) error
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
	// set the next router in the chain, returns the next router
	SetNextChain(next Router) Router
	// set the next router in the chain with prefix, returns the next router
	SetNextChainWithPrefix(next Router, prefix string) Router
}
