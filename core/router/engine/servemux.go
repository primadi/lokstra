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
	var fullPath string
	if method == "ANY" {
		fullPath = smPath
	} else {
		fullPath = method + " " + smPath
	}
	s.mux.Handle(fullPath, h)
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
