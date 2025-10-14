package lokstra_handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

// ReverseProxyRewrite represents path rewrite configuration
type ReverseProxyRewrite struct {
	From string // Pattern to match (regex supported)
	To   string // Replacement pattern
}

// MountReverseProxy mounts a reverse proxy.
// If stripPrefix is not empty, it will be stripped automatically.
// If rewrite is provided, path rewriting will be applied after stripping.
// Example:
//
//	  MountReverseProxy("/api", "http://localhost:9000", nil)
//		   -> /api          -> http://localhost:9000/
//		   -> /api/users    -> http://localhost:9000/users
//		   -> /api/products -> http://localhost:9000/products
//
// Example with rewrite:
//
//	  MountReverseProxy("/api", "http://localhost:9000", &ReverseProxyRewrite{From: "^/v1", To: "/v2"})
//		   -> /api/v1/users -> http://localhost:9000/v2/users
//
// Note: to cover both "/api" and "/api/*", register using ANYAll or both paths in ServeMux.
func MountReverseProxy(stripPrefix string, target string, rewrite *ReverseProxyRewrite) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("Invalid target URL: " + err.Error())
	}

	var rewriteRegex *regexp.Regexp
	if rewrite != nil && rewrite.From != "" {
		rewriteRegex, err = regexp.Compile(rewrite.From)
		if err != nil {
			panic("Invalid rewrite pattern: " + err.Error())
		}
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// Add custom director to handle path rewriting
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)

		// Apply rewrite if configured
		if rewriteRegex != nil && rewrite != nil {
			req.URL.Path = rewriteRegex.ReplaceAllString(req.URL.Path, rewrite.To)
			// Update raw path as well
			req.URL.RawPath = rewriteRegex.ReplaceAllString(req.URL.RawPath, rewrite.To)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, "Reverse proxy error: "+e.Error(), http.StatusBadGateway)
	}

	if stripPrefix != "" {
		return http.StripPrefix(stripPrefix, proxy)
	}
	return proxy
}

// MountReverseProxySimple is a simplified version that uses string replacement for rewrite
func MountReverseProxySimple(stripPrefix string, target string, rewriteFrom, rewriteTo string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		panic("Invalid target URL: " + err.Error())
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// Add custom director to handle path rewriting
	if rewriteFrom != "" {
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)

			// Apply simple string replacement
			req.URL.Path = strings.Replace(req.URL.Path, rewriteFrom, rewriteTo, 1)
			req.URL.RawPath = strings.Replace(req.URL.RawPath, rewriteFrom, rewriteTo, 1)
		}
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, "Reverse proxy error: "+e.Error(), http.StatusBadGateway)
	}

	if stripPrefix != "" {
		return http.StripPrefix(stripPrefix, proxy)
	}
	return proxy
}
