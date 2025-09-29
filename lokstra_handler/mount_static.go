package lokstra_handler

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// MountStatic serves static files with index.html fallback.
// Behavior:
//
//	/about     -> serve "about/index.html"
//	/about/    -> serve "about/index.html"
//	/logo.png  -> serve "logo.png"
//	not found  -> 404
func MountStatic(stripPrefix string, fsys fs.FS) http.Handler {
	if stripPrefix != "" {
		stripPrefix = "/" + strings.Trim(stripPrefix, "/")
	}
	fileHandler := http.StripPrefix(stripPrefix, http.FileServer(http.FS(fsys)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If base path has no extension and doesn't end with '/', append '/index.html'
		if !strings.ContainsRune(path.Base(r.URL.Path), '.') && !strings.HasSuffix(r.URL.Path, "/") {
			r.URL.Path += "/index.html"
		}
		fileHandler.ServeHTTP(w, r)
	})
}
