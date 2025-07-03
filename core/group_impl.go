package core

import (
	"lokstra/common/iface"
	"lokstra/core/request"
	"lokstra/core/router"
	"net/http"

	"github.com/valyala/fasthttp"
)

type GroupImpl struct {
	parent             *RouterImpl
	prefix             string
	middleware         []iface.MiddlewareFunc
	overrideMiddleware bool
	middlewareLocked   bool
}

// DumpRoutes implements router.Router.
func (g *GroupImpl) DumpRoutes() {
	g.parent.DumpRoutes()
}

// DELETE implements router.Router.
func (g *GroupImpl) DELETE(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	return g.Handle("DELETE", path, handler, mw...)
}

// FastHttpHandler implements router.Router.
func (g *GroupImpl) FastHttpHandler() fasthttp.RequestHandler {
	panic("unimplemented")
}

// GET implements router.Router.
func (g *GroupImpl) GET(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// GetMiddleware implements router.Router.
func (g *GroupImpl) GetMiddleware() []iface.MiddlewareFunc {
	panic("unimplemented")
}

// Group implements router.Router.
func (g *GroupImpl) Group(prefix string, mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// GroupBlock implements router.Router.
func (g *GroupImpl) GroupBlock(prefix string, fn func(gr router.Router)) router.Router {
	panic("unimplemented")
}

// Handle implements router.Router.
func (g *GroupImpl) Handle(method iface.HTTPMethod, path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// HandleOverrideMiddleware implements router.Router.
func (g *GroupImpl) HandleOverrideMiddleware(method iface.HTTPMethod, path string,
	handler request.HandlerFunc, mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// LockMiddleware implements router.Router.
func (g *GroupImpl) LockMiddleware() {
	panic("unimplemented")
}

// MountReverseProxy implements router.Router.
func (g *GroupImpl) MountReverseProxy(prefix string, target string) router.Router {
	panic("unimplemented")
}

// MountSPA implements router.Router.
func (g *GroupImpl) MountSPA(prefix string, fallbackFile string) router.Router {
	panic("unimplemented")
}

// MountStatic implements router.Router.
func (g *GroupImpl) MountStatic(prefix string, folder http.Dir) router.Router {
	panic("unimplemented")
}

// OverrideMiddleware implements router.Router.
func (g *GroupImpl) OverrideMiddleware() router.Router {
	panic("unimplemented")
}

// PATCH implements router.Router.
func (g *GroupImpl) PATCH(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// POST implements router.Router.
func (g *GroupImpl) POST(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// PUT implements router.Router.
func (g *GroupImpl) PUT(path string, handler request.HandlerFunc,
	mw ...iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// Prefix implements router.Router.
func (g *GroupImpl) Prefix() string {
	panic("unimplemented")
}

// RecurseAllHandler implements router.Router.
func (g *GroupImpl) RecurseAllHandler(callback func(rt *router.RouteHandlerData)) {
	panic("unimplemented")
}

// ServeHTTP implements router.Router.
func (g *GroupImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Use implements router.Router.
func (g *GroupImpl) Use(iface.MiddlewareFunc) router.Router {
	panic("unimplemented")
}

// WithOverrideMiddleware implements router.Router.
func (g *GroupImpl) WithOverrideMiddleware(enable bool) router.Router {
	panic("unimplemented")
}

// WithPrefix implements router.Router.
func (g *GroupImpl) WithPrefix(prefix string) router.Router {
	panic("unimplemented")
}

var _ router.Router = (*GroupImpl)(nil)
