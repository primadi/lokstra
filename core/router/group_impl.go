package router

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"

	"github.com/valyala/fasthttp"
)

type GroupImpl struct {
	parent *RouterImpl
	meta   *RouterMeta
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
func (g *GroupImpl) MountHtmx(prefix string, sources ...fs.FS) Router {
	g.meta.MountHtmx(g.cleanPrefix(prefix), sources...)
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

// GetMiddleware implements Router.
func (g *GroupImpl) GetMiddleware() []*midware.Execution {
	mw := make([]*midware.Execution, len(g.meta.Middleware))
	copy(mw, g.meta.Middleware)

	if g.meta.OverrideMiddleware {
		return mw
	}

	if len(g.parent.GetMiddleware())+len(mw) == 0 {
		return []*midware.Execution{}
	}

	// Slices.Concat return 0 ?
	return utils.SlicesConcat(g.parent.GetMiddleware(), mw)
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
	g.meta.Handle(method, g.cleanPrefix(path), handler, false, mw...)
	return g
}

// HandleOverrideMiddleware implements Router.
func (g *GroupImpl) HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router {
	g.meta.Handle(method, g.cleanPrefix(path), handler, true, mw...)
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
