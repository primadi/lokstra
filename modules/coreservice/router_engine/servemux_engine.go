package router_engine

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/primadi/lokstra/common/static_files"
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

// RawHandle implements RouterEngine.
func (m *ServeMuxEngine) RawHandle(pattern string, handler http.Handler) {
	m.mux.Handle(pattern, handler)
}

// RawHandleFunc implements RouterEngine.
func (m *ServeMuxEngine) RawHandleFunc(pattern string, handlerFunc http.HandlerFunc) {
	m.mux.HandleFunc(pattern, handlerFunc)
}

// ServeStatic implements RouterEngine.
func (m *ServeMuxEngine) ServeStatic(prefix string, spa bool, sources ...fs.FS) {
	cleanPrefixStr := cleanPrefix(prefix)

	staticServe := static_files.New(sources...)

	handler := staticServe.RawHandler(spa)
	// Strip prefix before passing to fallback handler
	if cleanPrefixStr != "/" {
		handler = http.StripPrefix(cleanPrefixStr, handler)
	}

	if cleanPrefixStr == "/" {
		m.mux.Handle("/", handler)
	} else {
		m.mux.Handle(cleanPrefixStr+"/", handler)
	}
}

// ServeHtmxPage implements serviceapi.RouterEngine.
func (m *ServeMuxEngine) ServeHtmxPage(pageDataRouter http.Handler,
	prefix string, si *static_files.ScriptInjection, sources ...fs.FS) {
	cleanPrefixStr := cleanPrefix(prefix)

	staticServe := static_files.New(sources...)

	handler := staticServe.HtmxPageHandlerWithScriptInjection(pageDataRouter, prefix, si)
	// Strip prefix before passing to fallback handler
	if cleanPrefixStr != "/" {
		handler = http.StripPrefix(cleanPrefixStr, handler)
	}

	if cleanPrefixStr == "/" {
		m.mux.Handle("/", handler)
	} else {
		m.mux.Handle(cleanPrefixStr, handler)
		m.mux.Handle(cleanPrefixStr+"/", handler)
	}
}

var _ serviceapi.RouterEngine = (*ServeMuxEngine)(nil)
var _ service.Service = (*ServeMuxEngine)(nil)
