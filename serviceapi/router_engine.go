package serviceapi

import (
	"net/http"

	"github.com/primadi/lokstra/core/request"
)

const HTTP_ROUTER_PREFIX string = "lokstra.http_router."

// RouterEngine defines the interface for a router engine that can handle HTTP methods,
// serve static files, single-page applications (SPA), and reverse proxies.
type RouterEngine interface {
	// HandleMethod registers a handler for a specific HTTP method and path.
	HandleMethod(method request.HTTPMethod, path string, handler http.Handler)

	ServeHTTP(w http.ResponseWriter, r *http.Request)
	ServeStatic(prefix string, folder http.Dir)
	ServeSPA(prefix string, indexFile string)
	ServeReverseProxy(prefix string, handler http.HandlerFunc)
}
