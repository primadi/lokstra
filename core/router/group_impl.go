package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/midware"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"

	"github.com/valyala/fasthttp"
)

type GroupImpl struct {
	parent   *RouterImpl
	meta     *RouterMeta
	mwLocked bool
}

// RawHandleStripPrefix implements Router.
func (g *GroupImpl) RawHandleStripPrefix(prefix string, handler http.Handler) Router {
	return g.RawHandle(prefix, http.StripPrefix(prefix, handler))
}

// RawHandle implements Router.
func (g *GroupImpl) RawHandle(prefix string, handler http.Handler) Router {
	g.parent.r_engine.RawHandle(prefix, handler)
	return g
}

// MountStaticWithFallback implements Router.
func (g *GroupImpl) MountStaticWithFallback(prefix string, sources ...any) Router {
	g.parent.r_engine.ServeStaticWithFallback(prefix, sources...)
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
	g.mwLocked = true

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
	g.mwLocked = true
	g.meta.Handle(method, g.cleanPrefix(path), handler, mw...)
	return g
}

// HandleOverrideMiddleware implements Router.
func (g *GroupImpl) HandleOverrideMiddleware(method request.HTTPMethod, path string, handler any, mw ...any) Router {
	g.mwLocked = true
	g.meta.HandleWithOverrideMiddleware(method, g.cleanPrefix(path), handler, mw...)
	return g
}

// LockMiddleware implements Router.
func (g *GroupImpl) LockMiddleware() {
	g.mwLocked = true
}

// MountReverseProxy implements Router.
func (g *GroupImpl) MountReverseProxy(prefix string, target string,
	overrideMiddleware bool, mw ...any) Router {
	g.mwLocked = true
	g.meta.MountReverseProxy(prefix, target, overrideMiddleware, mw...)
	return g
}

// MountSPA implements Router.
func (g *GroupImpl) MountSPA(prefix string, fallbackFile string) Router {
	g.parent.r_engine.ServeSPA(prefix, fallbackFile)
	return g
}

// MountStatic implements Router.
func (g *GroupImpl) MountStatic(prefix string, folder http.Dir) Router {
	g.parent.r_engine.ServeStatic(prefix, folder)
	return g
}

// MountRpcService implements Router.
func (g *GroupImpl) MountRpcService(path string, svc any, overrideMiddleware bool, mw ...any) Router {
	g.mwLocked = true

	cleanPath := g.cleanPrefix(path)
	if strings.HasSuffix(cleanPath, "/") {
		cleanPath += ":method"
	} else {
		cleanPath += "/:method"
	}

	rpcMeta := &service.RpcServiceMeta{
		MethodParam: "method",
	}
	switch s := svc.(type) {
	case string:
		rpcMeta.ServiceName = s
	case *service.RpcServiceMeta:
		rpcMeta = s
	case service.Service:
		rpcMeta.ServiceInst = s
	default:
		fmt.Printf("Service type: %T\n", svc)
		panic("Invalid service type, must be a string, *RpcServiceMeta, or iface.Service")
	}

	handlerMeta := &request.HandlerMeta{
		HandlerFunc: func(ctx *request.Context) error {
			return ctx.ErrorInternal("RpcService not yet resolved")
		},
		Extension: rpcMeta,
	}

	if overrideMiddleware {
		g.meta.HandleWithOverrideMiddleware("POST", cleanPath, handlerMeta, mw...)
	} else {
		g.meta.Handle("POST", cleanPath, handlerMeta, mw...)
	}
	return g
}

// OverrideMiddleware implements Router.
func (g *GroupImpl) OverrideMiddleware() Router {
	g.meta.OverrideMiddleware = true
	return g
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

// WithOverrideMiddleware implements Router.
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
