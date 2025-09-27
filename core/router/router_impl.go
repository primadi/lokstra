package router

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
)

type routerImpl struct {
	name       string
	engineType string
	pathPrefix string

	routes           []*route.Route
	middlewares      []request.HandlerFunc
	overrideParentMw bool
	children         []*routerImpl

	isChained bool
	nextChain *routerImpl

	isRoot bool

	buildOnce    sync.Once
	isBuilt      atomic.Bool
	routerEngine RouterEngine
}

func New(name string) Router {
	return &routerImpl{
		name:       name,
		engineType: "default",
		pathPrefix: "",
		isRoot:     true,
	}
}

func NewWithEngine(name string, engineType string) Router {
	return &routerImpl{
		name:       name,
		engineType: engineType,
		pathPrefix: "",
		isRoot:     true,
	}
}

// IsChained implements Router.
func (r *routerImpl) IsChained() bool {
	return r.isChained
}

// GetNextChain implements Router.
func (r *routerImpl) GetNextChain() Router {
	return r.nextChain
}

// IsBuilt implements Router.
func (r *routerImpl) IsBuilt() bool {
	return r.isBuilt.Load()
}

// Guard: forbid adding routes after build
func (r *routerImpl) assertNotBuilt() {
	if r.IsBuilt() {
		panic("router: cannot register routes after Build()")
	}
}

// Build implements Router.
func (r *routerImpl) Build() {
	if !r.isRoot {
		panic("router: Build() can only be called on the root router")
	}
	r.buildOnce.Do(func() {
		if !r.isChained {
			r.routerEngine = createEngine(r.engineType)
			r.walkBuildRecursive("", "", nil,
				func(rt *route.Route, fullName, fullPath string, fullMiddlewares []request.HandlerFunc) {
					rt.FullName = fullName
					rt.FullPath = fullPath

					var fullMw []request.HandlerFunc
					if rt.OverrideParentMw {
						fullMw = rt.Middleware
					} else {
						fullMw = append(fullMiddlewares, rt.Middleware...)
					}
					rt.FullMiddleware = fullMw
					r.routerEngine.Handle(rt.Method, fullPath, request.NewHandler(
						rt.Handler, fullMw...))
				})
		}
		r.isBuilt.Store(true)
	})
}

// ServeHTTP implements Router.
func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Build()
	r.routerEngine.ServeHTTP(w, req)
}

func (r *routerImpl) handle(method string, path string, h any, middleware []any) Router {
	r.assertNotBuilt()

	rt := &route.Route{
		Method: method,
		Path:   path,
	}

	var mws []any
	// Remove RouteOption from middleware list
	for _, mw := range middleware {
		if n, ok := mw.(route.RouteHandlerOption); ok {
			n.Apply(rt)
			continue
		}
		mws = append(mws, mw)
	}

	rt.Name = normalizeName(rt.Name, path, method)
	rt.Middleware = adaptMiddlewares(mws)
	rt.Handler = adaptHandler(rt.Name, h)
	r.routes = append(r.routes, rt)
	return r
}

// ANY implements Router.
func (r *routerImpl) ANY(path string, h any, middleware ...any) Router {
	return r.handle("ANY", cleanPath(path), h, middleware)
}

// ANYPrefix implements Router.
func (r *routerImpl) ANYPrefix(prefix string, h any, middleware ...any) Router {
	return r.handle("ANY", cleanPrefix(prefix), h, middleware)
}

// AddGroup implements Router.
func (r *routerImpl) AddGroup(path string) Router {
	r.assertNotBuilt()
	path = cleanPath(path)
	child := &routerImpl{
		name:       normalizeGroupName("", path),
		pathPrefix: path,
	}
	r.children = append(r.children, child)
	return child
}

// DELETE implements Router.
func (r *routerImpl) DELETE(path string, h any, middleware ...any) Router {
	return r.handle("DELETE", cleanPath(path), h, middleware)
}

// DELETEPrefix implements Router.
func (r *routerImpl) DELETEPrefix(prefix string, h any, middleware ...any) Router {
	return r.handle("DELETE", cleanPrefix(prefix), h, middleware)
}

// EngineType implements Router.
func (r *routerImpl) EngineType() string {
	return r.engineType
}

// GET implements Router.
func (r *routerImpl) GET(path string, h any, middleware ...any) Router {
	return r.handle("GET", cleanPath(path), h, middleware)
}

// GETPrefix implements Router.
func (r *routerImpl) GETPrefix(prefix string, h any, middleware ...any) Router {
	return r.handle("GET", cleanPrefix(prefix), h, middleware)
}

// Group implements Router.
func (r *routerImpl) Group(path string, fn func(r Router)) Router {
	fn(r.AddGroup(path))
	return r
}

// Name implements Router.
func (r *routerImpl) Name() string {
	return r.name
}

// PATCH implements Router.
func (r *routerImpl) PATCH(path string, h any, middleware ...any) Router {
	return r.handle("PATCH", cleanPath(path), h, middleware)
}

// PATCHPrefix implements Router.
func (r *routerImpl) PATCHPrefix(prefix string, h any, middleware ...any) Router {
	return r.handle("PATCH", cleanPrefix(prefix), h, middleware)
}

// POST implements Router.
func (r *routerImpl) POST(path string, h any, middleware ...any) Router {
	return r.handle("POST", cleanPath(path), h, middleware)
}

// POSTPrefix implements Router.
func (r *routerImpl) POSTPrefix(prefix string, h any, middleware ...any) Router {
	return r.handle("POST", cleanPrefix(prefix), h, middleware)
}

// PUT implements Router.
func (r *routerImpl) PUT(path string, h any, middleware ...any) Router {
	return r.handle("PUT", cleanPath(path), h, middleware)
}

// PUTPrefix implements Router.
func (r *routerImpl) PUTPrefix(prefix string, h any, middleware ...any) Router {
	return r.handle("PUT", cleanPrefix(prefix), h, middleware)
}

// PathPrefix implements Router.
func (r *routerImpl) PathPrefix() string {
	return r.pathPrefix
}

// Use implements Router.
func (r *routerImpl) Use(middleware ...any) Router {
	r.middlewares = append(r.middlewares, adaptMiddlewares(middleware)...)
	return r
}

// WithOverrideParentMiddleware implements Router.
func (r *routerImpl) WithOverrideParentMiddleware(override bool) Router {
	r.overrideParentMw = override
	return r
}

func (r *routerImpl) walkBuildRecursive(fullName, fullPrefix string, fullMw []request.HandlerFunc,
	fn func(*route.Route, string, string, []request.HandlerFunc)) {
	baseName := ""
	if !r.isRoot {
		baseName = fullName + r.name + "."
	}
	basePrefix := fullPrefix + r.pathPrefix
	var baseMw []request.HandlerFunc
	if r.overrideParentMw {
		baseMw = r.middlewares
	} else {
		baseMw = append(fullMw, r.middlewares...)
	}
	for _, rt := range r.routes {
		fn(rt, baseName+rt.Name, basePrefix+rt.Path, baseMw)
	}
	for _, child := range r.children {
		child.walkBuildRecursive(baseName, basePrefix, baseMw, fn)
	}
	if r.nextChain != nil {
		r.nextChain.walkBuildRecursive(fullName, fullPrefix, fullMw, fn)
	}
}

// Walk implements Router.
func (r *routerImpl) Walk(fn func(rt *route.Route)) {
	for _, rt := range r.routes {
		fn(rt)
	}
	for _, child := range r.children {
		child.Walk(fn)
	}
	if r.nextChain != nil {
		r.nextChain.Walk(fn)
	}
}

func (r *routerImpl) PrintRoutes() {
	r.Build()
	r.Walk(func(rt *route.Route) {
		var mwDescr string
		switch mwLen := len(rt.FullMiddleware); mwLen {
		case 0:
			mwDescr = ""
		case 1:
			mwDescr = " [with 1 mw]"
		default:
			mwDescr = fmt.Sprintf(" [with %d mw(s)]", mwLen)
		}
		fmt.Printf("[%s] %s %s -> %s%s\n", r.name, rt.Method, rt.FullPath, rt.FullName, mwDescr)
	})
}

var _ Router = (*routerImpl)(nil)
