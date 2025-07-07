package router

import (
	"lokstra/common/iface"
	"lokstra/common/meta"
	"lokstra/common/utils"
	"lokstra/core/request"
	"net/http"
	"slices"
	"strings"

	"github.com/valyala/fasthttp"
)

type GroupImpl struct {
	parent   *RouterImpl
	meta     *meta.RouterMeta
	mwLocked bool
}

// DumpRoutes implements Router.
func (g *GroupImpl) DumpRoutes() {
	panic("unimplemented")
}

// DELETE implements Router.
func (g *GroupImpl) DELETE(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	return g.Handle("DELETE", path, handler, mw...)
}

// FastHttpHandler implements Router.
func (g *GroupImpl) FastHttpHandler() fasthttp.RequestHandler {
	panic("unimplemented")
}

// GET implements Router.
func (g *GroupImpl) GET(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	return g.Handle("GET", path, handler, mw...)
}

// GetMiddleware implements Router.
func (g *GroupImpl) GetMiddleware() []iface.MiddlewareFunc {
	mw := make([]iface.MiddlewareFunc, len(g.meta.Middleware))
	for i, m := range g.meta.Middleware {
		mw[i] = m.MiddlewareFunc
	}

	if g.meta.OverrideMiddleware {
		return mw
	}

	return slices.Concat(g.parent.GetMiddleware(), mw)
}

// Group implements Router.
func (g *GroupImpl) Group(prefix string, mw ...iface.MiddlewareFunc) Router {
	g.mwLocked = true

	rm := meta.NewRouter()
	rm.Prefix = g.cleanPrefix(prefix)

	for _, m := range mw {
		rm.UseMiddleware(m)
	}

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
func (g *GroupImpl) Handle(method iface.HTTPMethod, path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	g.mwLocked = true

	gmw := slices.Concat(g.GetMiddleware(), mw)

	g.meta.Handle(method, g.cleanPrefix(path), handler, utils.ToAnySlice(gmw)...)
	g.parent.handle(method, g.cleanPrefix(path), handler, true, false, gmw...)
	return g
}

// HandleOverrideMiddleware implements Router.
func (g *GroupImpl) HandleOverrideMiddleware(method iface.HTTPMethod, path string,
	handler request.HandlerFunc, mw ...iface.MiddlewareFunc) Router {
	g.mwLocked = true

	g.meta.Handle(method, g.cleanPrefix(path), handler, utils.ToAnySlice(mw)...)
	g.parent.handle(method, g.cleanPrefix(path), handler, true, false, mw...)
	return g
}

// LockMiddleware implements Router.
func (g *GroupImpl) LockMiddleware() {
	g.mwLocked = true
}

// MountReverseProxy implements Router.
func (g *GroupImpl) MountReverseProxy(prefix string, target string) Router {
	g.parent.r_engine.ServeReverseProxy(g.cleanPrefix(prefix), target)
	return g
}

// MountSPA implements Router.
func (g *GroupImpl) MountSPA(prefix string, fallbackFile string) Router {
	g.parent.r_engine.ServeSPA(g.cleanPrefix(prefix), fallbackFile)
	return g
}

// MountStatic implements Router.
func (g *GroupImpl) MountStatic(prefix string, folder http.Dir) Router {
	g.parent.r_engine.ServeStatic(g.cleanPrefix(prefix), folder)
	return g
}

// OverrideMiddleware implements Router.
func (g *GroupImpl) OverrideMiddleware() Router {
	g.meta.OverrideMiddleware = true
	return g
}

// PATCH implements Router.
func (g *GroupImpl) PATCH(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	return g.Handle("PATCH", path, handler, mw...)
}

// POST implements Router.
func (g *GroupImpl) POST(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	return g.Handle("POST", path, handler, mw...)
}

// PUT implements Router.
func (g *GroupImpl) PUT(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) Router {
	return g.Handle("PUT", path, handler, mw...)
}

// Prefix implements Router.
func (g *GroupImpl) Prefix() string {
	return g.meta.Prefix
}

// RecurseAllHandler implements Router.
func (g *GroupImpl) RecurseAllHandler(callback func(rt *meta.RouteMeta)) {
	panic("unimplemented")
}

// ServeHTTP implements Router.
func (g *GroupImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Use implements Router.
func (g *GroupImpl) Use(mw iface.MiddlewareFunc) Router {
	g.meta.UseMiddleware(mw)
	return g
}

// WithOverrideMiddleware implements Router.
func (g *GroupImpl) WithOverrideMiddleware(enable bool) Router {
	g.meta.OverrideMiddleware = enable
	return g
}

// WithPrefix implements Router.
func (g *GroupImpl) WithPrefix(prefix string) Router {
	g.meta.Prefix = g.cleanPrefix(prefix)
	return g
}

var _ Router = (*GroupImpl)(nil)

func (g *GroupImpl) cleanPrefix(prefix string) string {
	if prefix == "/" || prefix == "" {
		return g.meta.Prefix
	}

	if g.meta.Prefix == "/" {
		return "/" + strings.Trim(prefix, "/")
	}
	return g.meta.Prefix + "/" + strings.Trim(prefix, "/")
}
