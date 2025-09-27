package router

import (
	"net/http"
	"sort"
	"strings"
)

// MethodServeMux is a method-aware multiplexer compatible with Lokstra Router.
type MethodServeMux struct {
	mux        map[string]*http.ServeMux // method → mux
	any        *http.ServeMux            // fallback for ANY
	allowCache map[string]string         // path → precomputed Allow header
}

// Creates a new MethodServeMux.
func NewMethodServeMux() RouterEngine {
	return &MethodServeMux{
		mux:        make(map[string]*http.ServeMux),
		any:        http.NewServeMux(),
		allowCache: make(map[string]string),
	}
}

func (m *MethodServeMux) Handle(method, path string, h http.Handler) {
	// Special case: ANY
	if method == "ANY" {
		m.any.Handle(path, h)
		m.allowCache[path] = "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD"
		return
	}

	// Register handler
	if m.mux[method] == nil {
		m.mux[method] = http.NewServeMux()
	}
	m.mux[method].Handle(path, h)

	// HEAD auto-mapping for GET
	if method == http.MethodGet {
		if m.mux[http.MethodHead] == nil {
			m.mux[http.MethodHead] = http.NewServeMux()
		}
		m.mux[http.MethodHead].Handle(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// call GET but discard body
			r2 := r.Clone(r.Context())
			r2.Method = http.MethodGet
			rr := newResponseRecorder(w)
			m.mux[http.MethodGet].ServeHTTP(rr, r2)
			rr.writeHeaderOnly()
		}))
	}

	// Update Allow header
	m.updateAllow(path, method)
}

func (m *MethodServeMux) updateAllow(path, method string) {
	current := m.allowCache[path]
	methods := strings.Split(current, ", ")
	if current == "" {
		methods = []string{}
	}

	// Add new method if not exists
	found := false
	for _, mth := range methods {
		if mth == method {
			found = true
			break
		}
	}
	if !found {
		methods = append(methods, method)
	}

	// OPTIONS always included
	hasOpt := false
	for _, mth := range methods {
		if mth == http.MethodOptions {
			hasOpt = true
			break
		}
	}
	if !hasOpt {
		methods = append(methods, http.MethodOptions)
	}

	// HEAD always included if GET exists
	if method == http.MethodGet {
		methods = append(methods, http.MethodHead)
	}

	// Normalize
	sort.Strings(methods)
	m.allowCache[path] = strings.Join(methods, ", ")
}

func (m *MethodServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS automatically
	if r.Method == http.MethodOptions {
		if allow := m.allowCache[r.URL.Path]; allow != "" {
			w.Header().Set("Allow", allow)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Normal dispatch
	if mux := m.mux[r.Method]; mux != nil {
		mux.ServeHTTP(w, r)
		return
	}

	// If ANY handler exists, run it
	if m.any != nil {
		// But only if path matches
		if _, ok := m.allowCache[r.URL.Path]; ok {
			m.any.ServeHTTP(w, r)
			return
		}
	}

	// If path known but method not allowed → 405
	if allow := m.allowCache[r.URL.Path]; allow != "" {
		w.Header().Set("Allow", allow)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Otherwise 404
	http.NotFound(w, r)
}

var _ RouterEngine = (*MethodServeMux)(nil)

// --------------------------
// ResponseRecorder (for HEAD auto-handler)
// --------------------------

type responseRecorder struct {
	w           http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{w: w}
}

func (rr *responseRecorder) Header() http.Header {
	return rr.w.Header()
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	// discard body
	if !rr.wroteHeader {
		rr.WriteHeader(http.StatusOK)
	}
	return len(b), nil
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.wroteHeader {
		rr.statusCode = code
		rr.wroteHeader = true
	}
}

func (rr *responseRecorder) writeHeaderOnly() {
	if rr.wroteHeader {
		rr.w.WriteHeader(rr.statusCode)
	}
}
