package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/route"
	"github.com/primadi/lokstra/core/router/engine"
)

type routerImpl struct {
	name       string
	engineType string
	pathPrefix string

	routes           []*route.Route
	middlewares      []any // Mixed: request.HandlerFunc or string (lazy)
	overrideParentMw bool
	children         []*routerImpl

	isChained bool
	nextChain *routerImpl

	isRoot bool

	isBuilt      bool
	routerEngine engine.RouterEngine
	startServe   sync.Once

	// Path rewrite rules (pattern, replacement)
	pathRewrites []pathRewrite
}

type pathRewrite struct {
	pattern     string
	replacement string
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
	r.walkBuildRecursive("", "", nil, r.name,
		func(rt *route.Route, fullName, fullPath string, fullMiddlewares []request.HandlerFunc, routerName string) {
			rt.RouterName = routerName // Set the router name for this route
			rt.FullName = fullName
			rt.FullPath = fullPath
			if rt.Name == "" {
				pref := ""
				if strings.HasSuffix(rt.FullPath, "/") {
					pref = "Prefix"
				}
				nm := fullPath
				if nm == "" {
					nm = "/"
				}
				rt.Name = strings.Join([]string{rt.Method, pref, "_", nm}, "")
				rt.FullName += rt.Name
			}

			// Resolve route-level lazy middlewares
			resolvedRouteMw := resolveMiddlewares(rt.Middleware)

			var fullMw []request.HandlerFunc
			if rt.OverrideParentMw {
				fullMw = resolvedRouteMw
			} else {
				fullMw = append(fullMiddlewares, resolvedRouteMw...)
			}
			rt.FullMiddleware = fullMw

			// Apply path rewrites (regex-based)
			rewrittenPath := fullPath
			for _, rw := range r.pathRewrites {
				re := regexp.MustCompile(rw.pattern)
				if re.MatchString(rewrittenPath) {
					rewrittenPath = re.ReplaceAllString(rewrittenPath, rw.replacement)
					break // Apply only first matching rule
				}
			}

			// Update route with rewritten path
			if rewrittenPath != fullPath {
				rt.FullPath = rewrittenPath
			}

			r.routerEngine.Handle(rt.Method+" "+rewrittenPath, request.NewHandler(
				rt.Handler, fullMw...))
		})
}

// ServeHTTP implements Router.
func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.startServe.Do(func() {
		// build router on first serve, do only once
		r.Build()
	})
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
		pathRewrites:     r.pathRewrites,
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

// SetPathRewrites sets regex-based path rewrite rules
func (r *routerImpl) SetPathRewrites(rewrites map[string]string) Router {
	r.pathRewrites = make([]pathRewrite, 0, len(rewrites))
	for pattern, replacement := range rewrites {
		r.pathRewrites = append(r.pathRewrites, pathRewrite{
			pattern:     pattern,
			replacement: replacement,
		})
	}
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

// UpdateRoute implements Router.
func (r *routerImpl) UpdateRoute(name string, options ...any) error {
	r.assertNotBuilt()

	// Find route by name
	var targetRoute *route.Route
	for _, rt := range r.routes {
		if rt.Name == name {
			targetRoute = rt
			break
		}
	}

	// If not found in this router, search in children
	if targetRoute == nil {
		for _, child := range r.children {
			if err := child.UpdateRoute(name, options...); err == nil {
				return nil // Found and updated in child
			}
		}
		// If still not found, search in chain
		if r.nextChain != nil {
			return r.nextChain.UpdateRoute(name, options...)
		}
		return fmt.Errorf("route '%s' not found in router '%s'", name, r.name)
	}

	// Process options
	var mws []any
	for _, opt := range options {
		// Apply RouteOption to the route
		if routeOpt, ok := opt.(route.RouteHandlerOption); ok {
			routeOpt.Apply(targetRoute)
			continue
		}
		// Collect middlewares
		mws = append(mws, opt)
	}

	// Append middlewares (same logic as handle method)
	if len(mws) > 0 {
		adaptedMws := adaptMiddlewares(mws)
		targetRoute.Middleware = append(targetRoute.Middleware, adaptedMws...)
	}

	return nil
}

// WithOverrideParentMiddleware implements Router.
func (r *routerImpl) WithOverrideParentMiddleware(override bool) Router {
	r.overrideParentMw = override
	return r
}

func (r *routerImpl) walkBuildRecursive(fullName, fullPrefix string, fullMw []request.HandlerFunc, routerName string,
	fn func(*route.Route, string, string, []request.HandlerFunc, string)) {
	baseName := fullName
	if r.isRoot {
		baseName += r.name + "."
	}
	basePrefix := fullPrefix + r.pathPrefix

	// Resolve lazy middlewares at this level
	var baseMw []request.HandlerFunc
	if r.overrideParentMw {
		baseMw = resolveMiddlewares(r.middlewares)
	} else {
		baseMw = append(fullMw, resolveMiddlewares(r.middlewares)...)
	} // Use current router name for routes directly in this router
	currentRouterName := r.name
	if currentRouterName == "" {
		currentRouterName = routerName
	}
	for _, rt := range r.routes {
		// Fix: Don't add trailing slash when path is "/"
		fullPath := basePrefix + rt.Path
		if rt.Path == "/" && basePrefix != "" {
			fullPath = basePrefix
		}
		fn(rt, baseName+rt.Name, fullPath, baseMw, currentRouterName)
	}
	for _, child := range r.children {
		child.walkBuildRecursive(baseName, basePrefix, baseMw, currentRouterName, fn)
	}
	if r.nextChain != nil {
		r.nextChain.walkBuildRecursive(fullName, fullPrefix, fullMw, routerName, fn)
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
		routerNameDisplay := rt.RouterName
		if routerNameDisplay == "" {
			routerNameDisplay = r.name
		}
		logger.LogInfo("[%s] %s %s -> %s%s", routerNameDisplay, rt.Method, rt.FullPath, rt.Name, mwDescr)
	})
}

var _ Router = (*routerImpl)(nil)
