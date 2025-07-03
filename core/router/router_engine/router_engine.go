package router_engine

import (
	"lokstra/common/iface"
	"net/http"
)

// RouterEngine defines the interface for a router engine that can handle HTTP methods,
// serve static files, single-page applications (SPA), and reverse proxies.
type RouterEngine interface {
	// HandleMethod registers a handler for a specific HTTP method and path.
	HandleMethod(method iface.HTTPMethod, path string, handler http.Handler)

	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeStatic(prefix string, folder http.Dir)
	ServeSPA(prefix string, indexFile string)
	ServeReverseProxy(prefix string, target string)
}
