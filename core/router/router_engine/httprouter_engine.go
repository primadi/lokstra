package router_engine

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func NewHTTPRouterEngine() RouterEngine {
	return HTTPRouterEngine{hr: httprouter.New()}
}

type HTTPRouterEngine struct {
	hr *httprouter.Router
	sm RouterEngine
}

// ServeReverseProxy implements RouterEngine.
func (h HTTPRouterEngine) ServeReverseProxy(prefix string, target string) {
	if h.sm == nil {
		h.sm = NewServeMuxEngine()
		h.hr.NotFound = h.sm
	}
	h.sm.ServeReverseProxy(prefix, target)
}

// ServeSPA implements RouterEngine.
func (h HTTPRouterEngine) ServeSPA(prefix string, indexFile string) {
	if h.sm == nil {
		h.sm = NewServeMuxEngine()
		h.hr.NotFound = h.sm
	}
	h.sm.ServeSPA(prefix, indexFile)
}

// ServeStatic implements RouterEngine.
func (h HTTPRouterEngine) ServeStatic(prefix string, folder http.Dir) {
	if h.sm == nil {
		h.sm = NewServeMuxEngine()
		h.hr.NotFound = h.sm
	}
	h.sm.ServeStatic(prefix, folder)
}

// ServeHTTP implements RouterEngine.
func (h HTTPRouterEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hr.ServeHTTP(w, r)
}

// Handle implements RouterEngine.
func (h HTTPRouterEngine) HandleMethod(method string, path string, handler http.Handler) {
	h.hr.Handle(method, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		for _, p := range ps {
			r.SetPathValue(p.Key, p.Value)
		}
		handler.ServeHTTP(w, r)
	})
}
