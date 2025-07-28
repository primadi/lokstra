package router_engine

import (
	"net/http"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/julienschmidt/httprouter"
)

func NewHttpRouterEngine(_ any) (service.Service, error) {
	return &HttpRouterEngine{
		hr: httprouter.New(),
	}, nil
}

type HttpRouterEngine struct {
	hr *httprouter.Router
	sm serviceapi.RouterEngine
}

// ServeReverseProxy implements RouterEngine.
func (h *HttpRouterEngine) ServeReverseProxy(prefix string, target string) {
	if h.sm == nil {
		sm, _ := NewServeMuxEngine(nil)
		h.sm = sm.(serviceapi.RouterEngine)
		h.hr.NotFound = h.sm
	}
	h.sm.ServeReverseProxy(prefix, target)
}

// ServeSPA implements RouterEngine.
func (h *HttpRouterEngine) ServeSPA(prefix string, indexFile string) {
	if h.sm == nil {
		sm, _ := NewServeMuxEngine(nil)
		h.sm = sm.(serviceapi.RouterEngine)
		h.hr.NotFound = h.sm
	}
	h.sm.ServeSPA(prefix, indexFile)
}

// ServeStatic implements RouterEngine.
func (h *HttpRouterEngine) ServeStatic(prefix string, folder http.Dir) {
	if h.sm == nil {
		sm, _ := NewServeMuxEngine(nil)
		h.sm = sm.(serviceapi.RouterEngine)
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
var _ service.Service = (*HttpRouterEngine)(nil)
