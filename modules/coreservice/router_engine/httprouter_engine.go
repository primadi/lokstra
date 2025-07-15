package router_engine

import (
	"lokstra/common/iface"
	"lokstra/serviceapi"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewHttpRouterEngine(_ any) (iface.Service, error) {
	return &HttpRouterEngine{hr: httprouter.New()}, nil
}

type HttpRouterEngine struct {
	hr *httprouter.Router
	sm serviceapi.RouterEngine
}

// GetRouterEngineType implements RouterEngine.
func (h *HttpRouterEngine) GetRouterEngineType() string {
	return serviceapi.HTTPROUTER_ROUTER_ENGINE_NAME
}

// ServeReverseProxy implements RouterEngine.
func (h *HttpRouterEngine) ServeReverseProxy(prefix string, target string) {
	if h.sm == nil {
		smAny, _ := NewServeMuxEngine(nil)
		h.sm = smAny.(serviceapi.RouterEngine)
		h.hr.NotFound = h.sm
	}
	h.sm.ServeReverseProxy(prefix, target)
}

// ServeSPA implements RouterEngine.
func (h *HttpRouterEngine) ServeSPA(prefix string, indexFile string) {
	if h.sm == nil {
		smAny, _ := NewServeMuxEngine(nil)
		h.sm = smAny.(serviceapi.RouterEngine)
	}
	h.sm.ServeSPA(prefix, indexFile)
}

// ServeStatic implements RouterEngine.
func (h *HttpRouterEngine) ServeStatic(prefix string, folder http.Dir) {
	if h.sm == nil {
		smAny, _ := NewServeMuxEngine(nil)
		h.sm = smAny.(serviceapi.RouterEngine)
		h.hr.NotFound = h.sm
	}
	h.sm.ServeStatic(prefix, folder)
}

// ServeHTTP implements RouterEngine.
func (h *HttpRouterEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hr.ServeHTTP(w, r)
}

// Handle implements RouterEngine.
func (h *HttpRouterEngine) HandleMethod(method string, path string, handler http.Handler) {
	h.hr.Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		for _, p := range ps {
			r.SetPathValue(p.Key, p.Value)
		}
		handler.ServeHTTP(w, r)
	})
}

var _ serviceapi.RouterEngine = (*HttpRouterEngine)(nil)
