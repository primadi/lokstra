package router_engine

import (
	"fmt"
	"lokstra/common/iface"
	"lokstra/serviceapi"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
)

func NewServeMuxEngine(_ any) (iface.Service, error) {
	return &ServeMuxEngine{mux: http.NewServeMux(), handlers: map[string]*handlerMethod{}}, nil
}

type ServeMuxEngine struct {
	mux      *http.ServeMux
	handlers map[string]*handlerMethod
}

// GetRouterEngineType implements RouterEngine.
func (m *ServeMuxEngine) GetRouterEngineType() string {
	return serviceapi.SERVEMUX_ROUTER_ENGINE_NAME
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
			if method == http.MethodOptions {
				w.Header().Set("Allow", hm.allowHeader)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// HEAD fallback to GET
			if method == http.MethodHead {
				if _, ok := hm.hm[http.MethodGet]; ok {
					method = http.MethodGet
					// replace writer to discard body
					w = headFallbackWriter{w}
				}
			}

			h, ok := hm.hm[method]
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
func (m *ServeMuxEngine) ServeReverseProxy(prefix string, target string) {
	targetURL, err := url.Parse(target)
	if err != nil {
		panic("invalid proxy target: " + err.Error())
	}

	cleanPrefix := cleanPrefix(prefix)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Update host to match target
			r.URL.Scheme = targetURL.Scheme
			r.URL.Host = targetURL.Host
			r.Host = targetURL.Host
			proxy.ServeHTTP(w, r)
		})

	if cleanPrefix == "/" {
		m.mux.Handle("/", handler)
	} else {
		m.mux.Handle(cleanPrefix+"/", http.StripPrefix(cleanPrefix, handler))
	}
}

// ServeSPA implements RouterEngine.
func (m *ServeMuxEngine) ServeSPA(prefix string, indexFile string) {
	rootDir := path.Dir(indexFile)
	fs := http.FileServer(http.Dir(rootDir))

	// Serve static files directly
	m.mux.HandleFunc(cleanPrefix(prefix), func(w http.ResponseWriter, r *http.Request) {
		requestPath := strings.TrimPrefix(r.URL.Path, prefix)
		if requestPath == "" || requestPath == "/" {
			http.ServeFile(w, r, indexFile)
			return
		}

		fullPath := path.Join(rootDir, requestPath)
		if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}

		if strings.Contains(path.Base(requestPath), ".") {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, indexFile)
	})
}

// ServeStatic implements RouterEngine.
func (m *ServeMuxEngine) ServeStatic(prefix string, folder http.Dir) {
	cleanPrefix := cleanPrefix(prefix)
	fs := http.StripPrefix(cleanPrefix, http.FileServer(folder))
	m.mux.Handle(cleanPrefix, fs)
}

var _ serviceapi.RouterEngine = (*ServeMuxEngine)(nil)
