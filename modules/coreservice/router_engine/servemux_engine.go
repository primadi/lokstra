package router_engine

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/serviceapi"
)

func NewServeMuxEngine(_ any) (service.Service, error) {
	return &ServeMuxEngine{
		mux:      http.NewServeMux(),
		handlers: map[string]*handlerMethod{},
	}, nil
}

type ServeMuxEngine struct {
	mux      *http.ServeMux
	handlers map[string]*handlerMethod
}

type handlerMethod struct {
	allowHeader string
	hm          map[string]http.Handler
}

// HandleMethod implements RouterEngine.
func (m *ServeMuxEngine) HandleMethod(method string, path string, handler http.Handler) {
	hm, ok := m.handlers[path]
	if !ok {
		hm = &handlerMethod{hm: map[string]http.Handler{}}
		m.handlers[path] = hm

		convertedPath := ConvertToServeMuxParamPath(path)
		m.mux.HandleFunc(convertedPath, func(w http.ResponseWriter, r *http.Request) {
			requestMethod := r.Method

			if requestMethod == http.MethodOptions {
				w.Header().Set("Allow", hm.allowHeader)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// HEAD fallback to GET
			if requestMethod == http.MethodHead {
				if _, ok := hm.hm[http.MethodGet]; ok {
					requestMethod = http.MethodGet
					// replace writer to discard body
					w = headFallbackWriter{w}
				}
			}

			h, ok := hm.hm[requestMethod]
			if !ok {
				w.Header().Set("Allow", hm.allowHeader)
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	if _, ok := hm.hm[method]; ok {
		panic(fmt.Errorf("path %s already has method %s", path, method))
	}

	hm.hm[method] = handler

	if hm.allowHeader == "" {
		hm.allowHeader = string(method)
	} else {
		hm.allowHeader += ", " + string(method)
	}
}

// ServeHTTP implements RouterEngine.
func (m *ServeMuxEngine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

// ServeReverseProxy implements RouterEngine.
func (m *ServeMuxEngine) ServeReverseProxy(prefix string, handler http.HandlerFunc) {
	cleanPrefix := cleanPrefix(prefix)
	if cleanPrefix == "/" {
		m.mux.Handle("/", handler)
	} else {
		m.mux.Handle(cleanPrefix+"/", http.StripPrefix(cleanPrefix, handler))
	}
}

// ServeSPA implements RouterEngine.
func (m *ServeMuxEngine) ServeSPA(prefix string, indexFile string) {
	rootDir := path.Dir(indexFile)

	spaHandler := func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(r.URL.Path, prefix)
		if requestPath == "" || requestPath == "/" {
			http.ServeFile(w, r, indexFile)
			return
		}

		fullPath := path.Join(rootDir, requestPath)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			// Serve the static file directly
			http.ServeFile(w, r, fullPath)
			return
		}

		if strings.Contains(path.Base(requestPath), ".") {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, indexFile)
	}

	// Register both exact prefix and sub-paths for SPA routing
	cleanPrefixStr := cleanPrefix(prefix)
	m.mux.HandleFunc(cleanPrefixStr, spaHandler)
	if cleanPrefixStr != "/" {
		m.mux.HandleFunc(cleanPrefixStr+"/", spaHandler)
	}
}

// ServeStatic implements RouterEngine.
func (m *ServeMuxEngine) ServeStatic(prefix string, folder http.Dir) {
	cleanPrefixStr := cleanPrefix(prefix)
	fs := http.StripPrefix(cleanPrefixStr, http.FileServer(folder))

	// For static file serving, we need trailing slash pattern to match sub-paths
	if cleanPrefixStr == "/" {
		m.mux.Handle("/", fs)
	} else {
		m.mux.Handle(cleanPrefixStr+"/", fs)
	}
}

var _ serviceapi.RouterEngine = (*ServeMuxEngine)(nil)
var _ service.Service = (*ServeMuxEngine)(nil)
