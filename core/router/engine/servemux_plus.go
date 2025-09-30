package engine

import (
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/primadi/lokstra/common/response_writer"
)

// Difference with Service Mux:
//   - HEAD method does not return BODY, only update Content-Length
//   - automatic insert OPTIONS in Allow Header when needed
type ServeMuxPlus struct {
	mux *http.ServeMux
}

// Creates a new ServeMuxPlus using Go 1.22+ features
func NewServeMuxPlus() RouterEngine {
	return &ServeMuxPlus{
		mux: http.NewServeMux(),
	}
}

func splitMethodPath(pattern string) (method, path string) {
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 {
		allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodPatch, http.MethodDelete, "ANY"}
		if slices.Contains(allowedMethods, parts[0]) {
			return parts[0], parts[1]
		}
		panic("Invalid method in pattern: " + pattern)
	}
	return "ANY", pattern
}

func (m *ServeMuxPlus) Handle(pattern string, h http.Handler) {
	method, path := splitMethodPath(pattern)

	if method == "ANY" {
		pattern = path
	}
	m.mux.Handle(pattern, h)
}

func (m *ServeMuxPlus) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbw := response_writer.NewBufferedBodyWriter(w)
	m.mux.ServeHTTP(dbw, r)

	// Ensure Allow header includes OPTIONS if applicable
	allow := w.Header().Get("Allow")
	if allow != "" {
		if !strings.Contains(allow, http.MethodOptions) {
			w.Header().Set("Allow", allow+", OPTIONS")
		}
	}

	// Auto Handling for OPTIONS
	if r.Method == http.MethodOptions {
		if allowCred := w.Header().Get("Access-Control-Allow-Methods"); allowCred != "" {
			// replace allow cred with allow
			w.Header().Set("Access-Control-Allow-Methods", allow)
		}

		// change not found to no content for OPTIONS
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Auto Handling for HEAD, discard body but set Content-Length
	if r.Method == http.MethodHead {
		// servemux auto call GET for HEAD method
		if dbw.Code == http.StatusOK {
			w.Header().Set("Content-Length", strconv.Itoa(dbw.Buf.Len()))
		}

		w.WriteHeader(dbw.Code)
		// not write BODY for HEAD request
		return
	}

	// normal methods
	w.WriteHeader(dbw.Code)
	w.Write(dbw.Buf.Bytes())
}

var _ RouterEngine = (*ServeMuxPlus)(nil)
