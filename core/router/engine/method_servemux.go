package engine

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

// findAllowHeaderForPath finds the Allow header for a given request path.
// This handles both exact matches and prefix patterns like "/say/{path...}".
func (m *MethodServeMux) findAllowHeaderForPath(requestPath string) string {
	// Try exact match first
	if allow, ok := m.allowCache[requestPath]; ok {
		return allow
	}

	// Try prefix patterns - check if any registered pattern could match this path
	for pattern, allow := range m.allowCache {
		if m.patternMatches(pattern, requestPath) {
			return allow
		}
	}

	return ""
} // patternMatches checks if a ServeMux pattern could match the given path.
// This is a simplified check for common patterns like "/prefix/{path...}".
func (m *MethodServeMux) patternMatches(pattern, requestPath string) bool {
	// Handle wildcard patterns like "/say/{path...}"
	if strings.HasSuffix(pattern, "/{path...}") {
		prefix := strings.TrimSuffix(pattern, "/{path...}")
		return strings.HasPrefix(requestPath, prefix+"/") || requestPath == prefix
	}

	// Handle other wildcard patterns if needed in the future
	// For now, just check exact match (already handled above)
	return false
}

func (m *MethodServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS automatically
	if r.Method == http.MethodOptions {
		if allow := m.findAllowHeaderForPath(r.URL.Path); allow != "" {
			w.Header().Set("Allow", allow)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// Normal dispatch - try method-specific handler first
	if mux := m.mux[r.Method]; mux != nil {
		recorder := &responseChecker{}
		mux.ServeHTTP(recorder, r)
		// If method-specific handler matched (didn't return 404), we're done
		if recorder.statusCode != http.StatusNotFound {
			recorder.writeTo(w)
			return
		}
		// If method-specific handler returned 404, try ANY handler
	}

	// If ANY handler exists, try it
	if m.any != nil {
		// Use a response recorder to check if ANY handler can handle this path
		recorder := &responseChecker{}
		m.any.ServeHTTP(recorder, r)
		// If ANY handler matched (didn't return 404), we're done
		if recorder.statusCode != http.StatusNotFound {
			recorder.writeTo(w)
			return
		}
		// If ANY handler returned 404, continue with normal 405/404 logic below
	}

	// If path known but method not allowed → 405
	if allow := m.findAllowHeaderForPath(r.URL.Path); allow != "" {
		w.Header().Set("Allow", allow)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Otherwise 404
	http.NotFound(w, r)
}

var _ RouterEngine = (*MethodServeMux)(nil)

// --------------------------
// responseDiscardBody (for HEAD auto-handler)
// --------------------------

type responseDiscardBody struct {
	w           http.ResponseWriter
	wroteHeader bool
	statusCode  int
}

func newResponseRecorder(w http.ResponseWriter) *responseDiscardBody {
	return &responseDiscardBody{w: w}
}

func (rr *responseDiscardBody) Header() http.Header {
	return rr.w.Header()
}

func (rr *responseDiscardBody) Write(b []byte) (int, error) {
	// discard body
	if !rr.wroteHeader {
		rr.WriteHeader(http.StatusOK)
	}
	return len(b), nil
}

func (rr *responseDiscardBody) WriteHeader(code int) {
	if !rr.wroteHeader {
		rr.statusCode = code
		rr.wroteHeader = true
	}
}

func (rr *responseDiscardBody) writeHeaderOnly() {
	if rr.wroteHeader {
		rr.w.WriteHeader(rr.statusCode)
	}
}

// --------------------------
// ResponseChecker (for checking if ANY handler matches)
// --------------------------

type responseChecker struct {
	statusCode    int
	headerWritten bool
	body          []byte
	headers       http.Header
}

func (rc *responseChecker) Header() http.Header {
	if rc.headers == nil {
		rc.headers = make(http.Header)
	}
	return rc.headers
}

func (rc *responseChecker) WriteHeader(code int) {
	if !rc.headerWritten {
		rc.statusCode = code
		rc.headerWritten = true
	}
}

func (rc *responseChecker) Write(b []byte) (int, error) {
	// If no status code set yet, assume 200
	if rc.statusCode == 0 {
		rc.statusCode = http.StatusOK
	}
	rc.body = append(rc.body, b...)
	return len(b), nil
}

func (rc *responseChecker) writeTo(w http.ResponseWriter) {
	// Copy headers
	for k, v := range rc.headers {
		w.Header()[k] = v
	}

	// Write status
	if rc.statusCode > 0 {
		w.WriteHeader(rc.statusCode)
	}

	// Write body
	if len(rc.body) > 0 {
		w.Write(rc.body)
	}
}
