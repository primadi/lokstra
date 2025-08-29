package router_engine

import (
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"

	"github.com/julienschmidt/httprouter"
)

func NewHttpRouterEngine(_ any) (service.Service, error) {
	engine := &HttpRouterEngine{
		hr:           httprouter.New(),
		customRoutes: make(map[string]map[string]http.Handler),
	}

	// Set custom NotFound handler to handle conflicting routes
	engine.hr.NotFound = http.HandlerFunc(engine.handleNotFound)

	return engine, nil
}

type HttpRouterEngine struct {
	hr           *httprouter.Router
	sm           serviceapi.RouterEngine
	customRoutes map[string]map[string]http.Handler // method -> path -> handler
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

// ServeSPA implements RouterEngine.
func (h *HttpRouterEngine) ServeSPA(prefix string, indexFile string) {
	h.getServeMux().ServeSPA(prefix, indexFile)
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
func (h *HttpRouterEngine) ServeStatic(prefix string, folder http.Dir) {
	h.getServeMux().ServeStatic(prefix, folder)
}

// ServeStaticWithFallback implements RouterEngine.
func (h *HttpRouterEngine) ServeStaticWithFallback(prefix string, sources ...any) {
	h.getServeMux().ServeStaticWithFallback(prefix, sources...)
}

// ServeHTTP implements RouterEngine.
func (h *HttpRouterEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.hr.ServeHTTP(w, r)
}

// Handle implements RouterEngine.
func (h *HttpRouterEngine) HandleMethod(method request.HTTPMethod, path string, handler http.Handler) {
	methodStr := string(method)

	// Initialize method map if needed
	if h.customRoutes[methodStr] == nil {
		h.customRoutes[methodStr] = make(map[string]http.Handler)
	}

	// Store the route in our custom registry
	h.customRoutes[methodStr][path] = handler

	// Try to register with httprouter, but handle conflicts gracefully
	defer func() {
		if r := recover(); r != nil {
			// If httprouter fails, we'll handle this route through custom matching
			// The route is already stored in customRoutes
		}
	}()

	h.hr.Handle(methodStr, path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Set path parameters in request context
		for _, p := range ps {
			r.SetPathValue(p.Key, p.Value)
		}
		handler.ServeHTTP(w, r)
	})
}

// handleNotFound handles routes that httprouter couldn't match
// This includes conflicting routes that were stored in customRoutes
func (h *HttpRouterEngine) handleNotFound(w http.ResponseWriter, r *http.Request) {
	methodRoutes := h.customRoutes[r.Method]
	if methodRoutes == nil {
		// Fall back to default 404
		http.NotFound(w, r)
		return
	}

	// Try to match custom routes
	for routePath, handler := range methodRoutes {
		if params, matches := h.matchRoute(routePath, r.URL.Path); matches {
			// Set path parameters
			for key, value := range params {
				r.SetPathValue(key, value)
			}
			handler.ServeHTTP(w, r)
			return
		}
	}

	// No match found
	http.NotFound(w, r)
}

// matchRoute matches a route pattern against a path and extracts parameters
func (h *HttpRouterEngine) matchRoute(pattern, path string) (map[string]string, bool) {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)

	for i, patternPart := range patternParts {
		pathPart := pathParts[i]

		if strings.HasPrefix(patternPart, ":") {
			// This is a parameter
			paramName := patternPart[1:] // Remove the ":"
			params[paramName] = pathPart
		} else if patternPart != pathPart {
			// Literal parts must match exactly
			return nil, false
		}
	}

	return params, true
}

var _ serviceapi.RouterEngine = (*HttpRouterEngine)(nil)
var _ service.Service = (*HttpRouterEngine)(nil)
