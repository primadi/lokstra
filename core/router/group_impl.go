package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/htmx_fsmanager"
	"github.com/primadi/lokstra/common/static_files"
	"github.com/primadi/lokstra/core/request"

	"github.com/valyala/fasthttp"
)

type GroupImpl struct {
	parent *RouterImpl
	meta   *RouterMeta
}

// SetHTMXLayoutScriptInjection implements Router.
func (g *GroupImpl) SetHTMXLayoutScriptInjection(si *htmx_fsmanager.ScriptInjection) Router {
	if g.parent.hfmContainer == nil {
		g.parent.hfmContainer = htmx_fsmanager.New()
	}
	g.parent.hfmContainer.SetScriptInjection(si)
	return g
}

// AddHtmxLayouts implements Router.
func (g *GroupImpl) AddHtmxLayouts(source fs.FS, dir ...string) Router {
	if g.parent.hfmContainer == nil {
		g.parent.hfmContainer = htmx_fsmanager.New()
	}
	g.parent.hfmContainer.AddLayoutFiles(source, dir...)
	return g
}

// AddHtmxPages implements Router.
func (g *GroupImpl) AddHtmxPages(source fs.FS, dir ...string) Router {
	if g.parent.hfmContainer == nil {
		g.parent.hfmContainer = htmx_fsmanager.New()
	}
	g.parent.hfmContainer.AddPageFiles(source, dir...)
	return g
}

// AddHtmxStatics implements Router.
func (g *GroupImpl) AddHtmxStatics(source fs.FS, dir ...string) Router {
	if g.parent.hfmContainer == nil {
		g.parent.hfmContainer = htmx_fsmanager.New()
	}
	g.parent.hfmContainer.AddStaticFiles(source, dir...)
	return g
}

// SetHtmxFSManager implements Router.
func (g *GroupImpl) SetHtmxFSManager(manager *htmx_fsmanager.HtmxFsManager) Router {
	g.parent.hfmContainer = manager
	return g
}

// RawHandle implements Router.
func (g *GroupImpl) RawHandle(prefix string, stripPrefix bool, handler http.Handler) Router {
	g.meta.RawHandles = append(g.meta.RawHandles, &RawHandleMeta{
		Prefix:  g.cleanPrefix(prefix),
		Handler: handler,
		Strip:   stripPrefix,
	})
	return g
}

// MountStatic implements Router.
func (g *GroupImpl) MountStatic(prefix string, spa bool, sources ...fs.FS) Router {
	g.meta.MountStatic(g.cleanPrefix(prefix), spa, sources...)
	return g
}

// MountHtmx implements Router.
func (g *GroupImpl) MountHtmx(prefix string, si *static_files.ScriptInjection,
	sources ...fs.FS) Router {
	g.meta.MountHtmx(g.cleanPrefix(prefix), si, sources...)
	return g
}

// GetMeta implements Router.
func (g *GroupImpl) GetMeta() *RouterMeta {
	return g.meta
}

// DumpRoutes implements Router.
func (g *GroupImpl) DumpRoutes() {
	panic("unimplemented")
}

// DELETE implements Router.
func (g *GroupImpl) DELETE(path string, handler any, mw ...any) Router {
	return g.Handle("DELETE", path, handler, mw...)
}

// FastHttpHandler implements Router.
func (g *GroupImpl) FastHttpHandler() fasthttp.RequestHandler {
	panic("unimplemented")
}

// GET implements Router.
func (g *GroupImpl) GET(path string, handler any, mw ...any) Router {
	return g.Handle("GET", path, handler, mw...)
}

// GETPrefix implements Router.
func (g *GroupImpl) GETPrefix(pathPrefix string, handler any, mw ...any) Router {
	return g.HandlePrefix("GET", pathPrefix, handler, mw...)
}

// Group implements Router.
func (g *GroupImpl) Group(prefix string, mw ...any) Router {
	rm := NewRouterMeta()
	rm.Prefix = g.cleanPrefix(prefix)

	for _, m := range mw {
		rm.UseMiddleware(m)
	}

	g.meta.Groups = append(g.meta.Groups, rm)
	return &GroupImpl{
		parent: g.parent,
		meta:   rm,
	}
}

// GroupBlock implements Router.
func (g *GroupImpl) GroupBlock(prefix string, fn func(gr Router)) Router {
	gr := g.Group(prefix)
	fn(gr)
	return g
}

// Handle implements Router.
func (g *GroupImpl) Handle(method request.HTTPMethod, path string, handler any, mw ...any) Router {
	prefixPath := g.cleanPrefix(path)
	if strings.HasSuffix(path, "/") {
		// remove trailing slash for exact match
		prefixPath, _ = strings.CutSuffix(prefixPath, "/")
	}
	g.meta.Handle(method, prefixPath, handler, false, mw...)
	return g
}

// HandleOverrideMiddleware implements Router.
func (g *GroupImpl) HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router {
	g.meta.Handle(method, g.cleanPrefix(path), handler, true, mw...)
	return g
}

// HandlePrefix implements prefix-based routing (catch-all)
func (g *GroupImpl) HandlePrefix(method request.HTTPMethod, prefix string, handler any, mw ...any) Router {
	// For prefix routing, ensure the path ends with "/" for ServeMux prefix matching
	prefixPath := g.cleanPrefix(prefix)
	if !strings.HasSuffix(prefixPath, "/") {
		prefixPath += "/"
	}

	g.meta.Handle(method, prefixPath, handler, false, mw...)
	return g
}

// MountReverseProxy implements Router.
func (g *GroupImpl) MountReverseProxy(prefix string, target string,
	overrideMiddleware bool, mw ...any) Router {
	g.meta.MountReverseProxy(g.cleanPrefix(prefix), target, overrideMiddleware, mw...)
	return g
}

// MountRpcService implements Router.
func (g *GroupImpl) MountRpcService(path string, svc any, overrideMiddleware bool, mw ...any) Router {
	g.meta.MountRpcService(g.cleanPrefix(path), svc, overrideMiddleware, mw...)
	return g
}

// OverrideMiddleware implements Router.
func (g *GroupImpl) OverrideMiddleware() bool {
	return g.meta.OverrideMiddleware
}

// PATCH implements Router.
func (g *GroupImpl) PATCH(path string, handler any, mw ...any) Router {
	return g.Handle("PATCH", path, handler, mw...)
}

// POST implements Router.
func (g *GroupImpl) POST(path string, handler any, mw ...any) Router {
	return g.Handle("POST", path, handler, mw...)
}

// PUT implements Router.
func (g *GroupImpl) PUT(path string, handler any, mw ...any) Router {
	return g.Handle("PUT", path, handler, mw...)
}

// POSTPrefix implements Router.
func (g *GroupImpl) POSTPrefix(pathPrefix string, handler any, mw ...any) Router {
	return g.HandlePrefix("POST", pathPrefix, handler, mw...)
}

// PUTPrefix implements Router.
func (g *GroupImpl) PUTPrefix(pathPrefix string, handler any, mw ...any) Router {
	return g.HandlePrefix("PUT", pathPrefix, handler, mw...)
}

// PATCHPrefix implements Router.
func (g *GroupImpl) PATCHPrefix(pathPrefix string, handler any, mw ...any) Router {
	return g.HandlePrefix("PATCH", pathPrefix, handler, mw...)
}

// DELETEPrefix implements Router.
func (g *GroupImpl) DELETEPrefix(pathPrefix string, handler any, mw ...any) Router {
	return g.HandlePrefix("DELETE", pathPrefix, handler, mw...)
}

// Prefix implements Router.
func (g *GroupImpl) Prefix() string {
	return g.meta.Prefix
}

// RecurseAllHandler implements Router.
func (g *GroupImpl) RecurseAllHandler(callback func(rt *RouteMeta)) {
	panic("unimplemented")
}

// ServeHTTP implements Router.
func (g *GroupImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Use implements Router.
func (g *GroupImpl) Use(mw any) Router {
	g.meta.UseMiddleware(mw)
	return g
}

// AddRouter implements Router.
func (g *GroupImpl) AddRouter(other Router) Router {
	otherMeta := other.GetMeta()

	// Merge all routes
	g.meta.Routes = append(g.meta.Routes, otherMeta.Routes...)

	// Merge all groups
	g.meta.Groups = append(g.meta.Groups, otherMeta.Groups...)

	// Merge static mounts
	g.meta.StaticMounts = append(g.meta.StaticMounts, otherMeta.StaticMounts...)

	// Merge reverse proxies
	g.meta.ReverseProxies = append(g.meta.ReverseProxies, otherMeta.ReverseProxies...)

	// Merge raw handles
	g.meta.RawHandles = append(g.meta.RawHandles, otherMeta.RawHandles...)

	// Merge RPC handles
	g.meta.RPCHandles = append(g.meta.RPCHandles, otherMeta.RPCHandles...)

	// Middleware: TIDAK di-merge secara otomatis untuk menghindari efek samping
	// Middleware dari router lain tetap melekat pada route/group asalnya
	// Jika ingin merge middleware, gunakan Use() secara eksplisit setelah AddRouter()

	return g
} // WithOverrideMiddleware implements Router.
func (g *GroupImpl) WithOverrideMiddleware(enable bool) Router {
	g.meta.OverrideMiddleware = enable
	return g
}

// WithPrefix implements Router.
func (g *GroupImpl) WithPrefix(prefix string) Router {
	if prefix == "/" || prefix == "" {
		return g
	}

	if strings.HasPrefix(prefix, "/") {
		g.meta.Prefix = "/" + strings.Trim(prefix, "/") // replace absolute prefix
	} else {
		g.meta.Prefix = g.cleanPrefix(prefix) // add relative prefix
	}
	return g
}

var _ Router = (*GroupImpl)(nil)

func (g *GroupImpl) cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return g.meta.Prefix
	}

	cleaned := strings.Trim(prefix, "/")

	var result string
	if strings.HasSuffix(g.meta.Prefix, "/") {
		result = g.meta.Prefix + cleaned
	} else {
		result = g.meta.Prefix + "/" + cleaned
	}

	if strings.HasSuffix(prefix, "/") {
		result += "/"
	}

	return result
}
