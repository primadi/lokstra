package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router/engine"
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

	isBuilt      bool
	routerEngine engine.RouterEngine
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
	if r.nextChain == nil {
		return nil
	}
	return r.nextChain
}

// IsBuilt implements Router.
func (r *routerImpl) IsBuilt() bool {
	return r.isBuilt
}

// Guard: forbid adding routes after build
func (r *routerImpl) assertNotBuilt() {
	if r.isBuilt {
		panic("router: cannot register routes after Build()")
	}
}

// Build implements Router.
func (r *routerImpl) Build() {
	if r.isBuilt || r.isChained {
		return
	}
	if !r.isRoot {
		panic("router [" + r.name + "] is not root router, Build() can only be called on the root router")
	}

	r.routerEngine = engine.CreateEngine(r.engineType)
	r.walkBuildRecursive("", "", nil,
		func(rt *route.Route, fullName, fullPath string, fullMiddlewares []request.HandlerFunc) {
			rt.FullName = fullName
			rt.FullPath = fullPath
			if rt.Name == "" {
				pref := ""
				if rt.FullPath != "/" && strings.HasSuffix(rt.FullPath, "/") {
					pref = "PREF:"
				}
				nm := strings.ReplaceAll(strings.Trim(fullPath, "/"), "/", "_")
				if nm == "" {
					nm = "root"
				}
				rt.Name = strings.Join([]string{rt.Method, "[", pref, nm, "]"}, "")
				rt.FullName += rt.Name
			}
			var fullMw []request.HandlerFunc
			if rt.OverrideParentMw {
				fullMw = rt.Middleware
			} else {
				fullMw = append(fullMiddlewares, rt.Middleware...)
			}
			rt.FullMiddleware = fullMw
			r.routerEngine.Handle(rt.Method+" "+fullPath, request.NewHandler(
				rt.Handler, fullMw...))
		})
}

// ServeHTTP implements Router.
func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Build()
	r.routerEngine.ServeHTTP(w, req)
}

func (r *routerImpl) handle(method string, path string, h any, middleware []any) Router {
	r.assertNotBuilt()

	if path == "" {
		path = "/"
	}
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

	rt.Middleware = adaptMiddlewares(mws)
	rt.Handler = adaptHandler(path, h)
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

// Clone implements Router.
func (r *routerImpl) Clone() Router {
	return &routerImpl{
		name:             r.name,
		engineType:       r.engineType,
		pathPrefix:       r.pathPrefix,
		routes:           r.routes,
		middlewares:      r.middlewares,
		overrideParentMw: r.overrideParentMw,
		children:         r.children,
		isRoot:           true,
	}
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

// SetPathPrefix implements Router.
func (r *routerImpl) SetPathPrefix(prefix string) Router {
	r.pathPrefix = cleanPath(prefix)
	return r
}

// SetNextChain implements Router.
func (r *routerImpl) SetNextChain(next Router) Router {
	return r.SetNextChainWithPrefix(next, "")
}

// SetNextChain implements Router.
func (r *routerImpl) SetNextChainWithPrefix(next Router, prefix string) Router {
	curr := r
	for curr.nextChain != nil {
		curr = curr.nextChain
	}
	if nc, ok := next.(*routerImpl); ok {
		nc.isChained = true
		nc.pathPrefix = cleanPath(prefix) + nc.pathPrefix
		curr.nextChain = nc
	} else {
		panic("router: SetNextChain expects a *routerImpl")
	}
	return r
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
	baseName := fullName
	if r.isRoot {
		baseName += r.name + "."
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
	r.isBuilt = true
}

// Walk implements Router.
func (r *routerImpl) Walk(fn func(rt *route.Route)) {
	r.Build()
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
