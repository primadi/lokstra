package lokstra_handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// MountReverseProxy mounts a reverse proxy.
// If stripPrefix is not empty, it will be stripped automatically.
// Example:
//
//	  MountReverseProxy("/api", "http://localhost:9000")
//		   -> /api          -> http://localhost:9000/
//		   -> /api/users    -> http://localhost:9000/users
//		   -> /api/products -> http://localhost:9000/products
//
// Note: to cover both "/api" and "/api/*", register using ANYAll or both paths in ServeMux.
func MountReverseProxy(stripPrefix string, target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("Invalid target URL: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, "Reverse proxy error: "+e.Error(), http.StatusBadGateway)
	}

	if stripPrefix != "" {
		return http.StripPrefix(stripPrefix, proxy)
	}
	return proxy
}
