package engine

import (
	"net/http"
)

type ServeMux struct {
	mux *http.ServeMux
}

// Handle implements RouterEngine.
func (s *ServeMux) Handle(pattern string, h http.Handler) {
	method, path := splitMethodPath(pattern)

	smPath := convertToServeMuxPattern(path)
	if method == "ANY" {
		pattern = smPath
	}
	s.mux.Handle(method+" "+smPath, h)
}

// ServeHTTP implements RouterEngine.
func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func NewServeMux() RouterEngine {
	return &ServeMux{
		mux: http.NewServeMux(),
	}
}

var _ RouterEngine = (*ServeMux)(nil)
