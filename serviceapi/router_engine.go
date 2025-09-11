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

	ServeStatic(prefix string, spa bool, sources ...fs.FS)
	ServeReverseProxy(prefix string, handler http.HandlerFunc)

	// Assume sources has:
	//   - "/layouts" for HTML layout templates
	//   - "/pages" for HTML page templates
	//
	// All Request paths will be treated as page requests,
	// except those starting with staticFolders list
	// which will be treated as static asset requests.
	ServeHtmxPage(pageDataRouter http.Handler, prefix string,
		staticFolders []string, sources ...fs.FS)
}
