package serviceapi

import (
	"io/fs"
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

	RawHandle(pattern string, handler http.Handler)
	RawHandleFunc(pattern string, handlerFunc http.HandlerFunc)

	ServeStatic(prefix string, folder http.Dir)

	// Sources can be http.Dir, fs.FS, or embed.FS
	ServeStaticWithFallback(prefix string, spa bool, sources ...fs.FS)

	ServeSPA(prefix string, indexFile string)
	ServeReverseProxy(prefix string, handler http.HandlerFunc)
}
