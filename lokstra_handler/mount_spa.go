package lokstra_handler

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// MountSpa serves a Single Page Application (SPA).
// Behavior:
//
//	/about       -> serve "index.html" (from root)
//	/users/123   -> serve "index.html" (from root)
//	/logo.png    -> serve "logo.png"
//	not found    -> 404 (unless fallback is index.html for no-ext paths)
func MountSpa(stripPrefix string, fsys fs.FS) http.Handler {
	if stripPrefix != "" {
		stripPrefix = "/" + strings.Trim(stripPrefix, "/")
	}
	fileHandler := http.StripPrefix(stripPrefix, http.FileServer(http.FS(fsys)))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if base path has no extension, serve "/index.html"
		if !strings.ContainsRune(path.Base(r.URL.Path), '.') {
			r.URL.Path = "/index.html"
		}
		fileHandler.ServeHTTP(w, r)
	})
}
