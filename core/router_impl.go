package core

import "net/http"

type RouterImpl struct {
}

// DELETE implements Router.
func (r *RouterImpl) DELETE(path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// GET implements Router.
func (r *RouterImpl) GET(path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// GetMiddleware implements Router.
func (r *RouterImpl) GetMiddleware() []MiddlewareHandler {
	panic("unimplemented")
}

// Group implements Router.
func (r *RouterImpl) Group(prefix string) Router {
	panic("unimplemented")
}

// GroupBlock implements Router.
func (r *RouterImpl) GroupBlock(prefix string, fn func(gr Router)) Router {
	panic("unimplemented")
}

// Handle implements Router.
func (r *RouterImpl) Handle(method HTTPMethod, path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// HandleOverrideMiddleware implements Router.
func (r *RouterImpl) HandleOverrideMiddleware(method HTTPMethod, path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// LockMiddleware implements Router.
func (r *RouterImpl) LockMiddleware() {
	panic("unimplemented")
}

// MountReverseProxy implements Router.
func (r *RouterImpl) MountReverseProxy(prefix string, target string) Router {
	panic("unimplemented")
}

// MountSPA implements Router.
func (r *RouterImpl) MountSPA(prefix string, fallbackFile string) Router {
	panic("unimplemented")
}

// MountStatic implements Router.
func (r *RouterImpl) MountStatic(prefix string, folder http.Dir) Router {
	panic("unimplemented")
}

// OverrideMiddleware implements Router.
func (r *RouterImpl) OverrideMiddleware() Router {
	panic("unimplemented")
}

// PATCH implements Router.
func (r *RouterImpl) PATCH(path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// POST implements Router.
func (r *RouterImpl) POST(path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// Prefix implements Router.
func (r *RouterImpl) Prefix() string {
	panic("unimplemented")
}

// PUT implements Router.
func (r *RouterImpl) PUT(path string, handler RequestHandler, mw ...MiddlewareHandler) Router {
	panic("unimplemented")
}

// RecurseAllHandler implements Router.
func (r *RouterImpl) RecurseAllHandler(callback func(info RouteInfo)) {
	panic("unimplemented")
}

// ServeHTTP implements Router.
func (*RouterImpl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	panic("unimplemented")
}

// Use implements Router.
func (r *RouterImpl) Use(MiddlewareHandler) Router {
	panic("unimplemented")
}

// UseNamedMiddleware implements Router.
func (r *RouterImpl) UseNamedMiddleware(mwname string, params MiddlewareConfig) error {
	panic("unimplemented")
}

func (r *RouterImpl) DumpRoutes() {

}

func NewRouter() Router {
	return &RouterImpl{}
}
