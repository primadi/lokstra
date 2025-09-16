package router_engine

import (
	"io/fs"
	"net/http"

	"github.com/primadi/lokstra/common/static_files"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/julienschmidt/httprouter"
)

func NewHttpRouterEngine(_ any) (service.Service, error) {
	engine := &HttpRouterEngine{
		hr: httprouter.New(),
	}

	return engine, nil
}

type HttpRouterEngine struct {
	hr *httprouter.Router
	sm serviceapi.RouterEngine
}

func (h *HttpRouterEngine) getServeMux() serviceapi.RouterEngine {
	if h.sm == nil {
		sm, _ := NewServeMuxEngine(nil)
		h.sm = sm.(serviceapi.RouterEngine)
		h.hr.NotFound = h.sm
	}
	return h.sm
}

// ServeReverseProxy implements RouterEngine.
func (h *HttpRouterEngine) ServeReverseProxy(prefix string, handler http.HandlerFunc) {
	h.getServeMux().ServeReverseProxy(prefix, handler)
}

// RawHandle implements RouterEngine.
func (h *HttpRouterEngine) RawHandle(pattern string, handler http.Handler) {
	h.getServeMux().RawHandle(pattern, handler)
}

// RawHandleFunc implements RouterEngine.
func (h *HttpRouterEngine) RawHandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	h.getServeMux().RawHandleFunc(pattern, handlerFunc)
}

// ServeStatic implements RouterEngine.
func (h *HttpRouterEngine) ServeStatic(prefix string, spa bool, sources ...fs.FS) {
	h.getServeMux().ServeStatic(prefix, spa, sources...)
}

// ServeHtmxPage implements serviceapi.RouterEngine.
func (h *HttpRouterEngine) ServeHtmxPage(pageDataRouter http.Handler,
	prefix string, si *static_files.ScriptInjection, sources ...fs.FS) {
	h.getServeMux().ServeHtmxPage(pageDataRouter, prefix, si, sources...)
}

// ServeHTTP implements RouterEngine.
func (h *HttpRouterEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hr.ServeHTTP(w, r)
}

// Handle implements RouterEngine.
func (h *HttpRouterEngine) HandleMethod(method request.HTTPMethod, path string, handler http.Handler) {
	convertedPath := ConvertToHttpRouterParamPath(path)

	h.hr.Handle(method, convertedPath, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Set path parameters in request context
		for _, p := range ps {
			r.SetPathValue(p.Key, p.Value)
		}
		handler.ServeHTTP(w, r)
	})
}

var _ serviceapi.RouterEngine = (*HttpRouterEngine)(nil)
var _ service.Service = (*HttpRouterEngine)(nil)
