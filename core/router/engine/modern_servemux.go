package engine

import (
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// ModernServeMux uses Go 1.22+ http.ServeMux with method-aware routing
// while still providing automatic HEAD/OPTIONS handling
type ModernServeMux struct {
	mux        *http.ServeMux
	allowCache map[string][]string // path → allowed methods
}

// NewModernServeMux creates a new ModernServeMux using Go 1.22+ features
func NewModernServeMux() RouterEngine {
	return &ModernServeMux{
		mux:        http.NewServeMux(),
		allowCache: make(map[string][]string),
	}
}

func splitMethodPath(pattern string) (method, path string) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		allowedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "ANY"}
		if slices.Contains(allowedMethods, parts[0]) {
			return parts[0], parts[1]
		}
		panic("Invalid method in pattern: " + pattern)
	}
	return "ANY", pattern
}

func (m *ModernServeMux) Handle(pattern string, h http.Handler) {
	method, path := splitMethodPath(pattern)

	// Special case: ANY method - use pattern without method prefix
	if method == "ANY" {
		m.mux.Handle(path, h) // Go 1.22+ supports method-less patterns
		// ANY allows all common HTTP methods
		m.allowCache[path] = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
		return
	}

	// Use Go 1.22+ method-aware pattern: "METHOD /path"
	if method == "ANY" {
		pattern = path
	}
	m.mux.Handle(pattern, h)

	// Update allowed methods for this path
	m.updateAllowedMethods(path, method)

	// Note: Go 1.22+ ServeMux automatically handles HEAD requests for GET routes
	// and ANY routes (method-less patterns) already handle all HTTP methods including HEAD
	// So we don't need to auto-generate HEAD handlers

	// Update allowed methods to include HEAD for GET routes
	if method == "GET" {
		m.updateAllowedMethods(path, "HEAD")
	}
}

func (m *ModernServeMux) updateAllowedMethods(path, method string) {
	methods := m.allowCache[path]

	// Add method if not already present
	found := false
	for _, m := range methods {
		if m == method {
			found = true
			break
		}
	}
	if !found {
		methods = append(methods, method)
	}

	// Always include OPTIONS
	optionsFound := false
	for _, m := range methods {
		if m == "OPTIONS" {
			optionsFound = true
			break
		}
	}
	if !optionsFound {
		methods = append(methods, "OPTIONS")
	}

	sort.Strings(methods)
	m.allowCache[path] = methods
}

func (m *ModernServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Handle OPTIONS automatically
	if r.Method == "OPTIONS" {
		if methods := m.findAllowedMethods(r.URL.Path); len(methods) > 0 {
			w.Header().Set("Allow", strings.Join(methods, ", "))
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	if r.Method == "HEAD" {
		dbw := &discardBodyWriter{ResponseWriter: w}
		m.mux.ServeHTTP(dbw, r)
		if dbw.code == 200 {
			w.Header().Set("Content-Length", strconv.Itoa(dbw.totalLength))
		}
		w.WriteHeader(dbw.code)
		return
	}

	// Normal handling for other methods
	m.mux.ServeHTTP(w, r)
}

// findAllowedMethods finds allowed methods for the given path
func (m *ModernServeMux) findAllowedMethods(requestPath string) []string {
	// Try exact match first
	if methods, ok := m.allowCache[requestPath]; ok {
		return methods
	}

	// Try pattern matching
	for pattern, methods := range m.allowCache {
		if m.patternMatches(pattern, requestPath) {
			return methods
		}
	}

	return nil
}

// patternMatches checks if a pattern could match the given path
func (m *ModernServeMux) patternMatches(pattern, requestPath string) bool {
	// Handle wildcard patterns like "/api/{path...}"
	if strings.HasSuffix(pattern, "/{path...}") {
		prefix := strings.TrimSuffix(pattern, "/{path...}")
		return strings.HasPrefix(requestPath, prefix+"/") || requestPath == prefix
	}

	// Handle other Go 1.22+ patterns if needed
	// For now, just exact match (already handled above)
	return false
}

var _ RouterEngine = (*ModernServeMux)(nil)

// discardBodyWriter discards any body written (for HEAD requests)
type discardBodyWriter struct {
	http.ResponseWriter
	totalLength int
	code        int
}

func (d *discardBodyWriter) Write(b []byte) (int, error) {
	d.totalLength += len(b)
	return len(b), nil
}

func (d *discardBodyWriter) WriteHeader(code int) {
	d.code = code
}
