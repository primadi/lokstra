package router

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/primadi/lokstra/core/request"
)

type StaticFallback struct {
	Sources []fs.FS
}

// NewStaticFallback creates a new StaticFallback instance.
func NewStaticFallback(sources ...fs.FS) *StaticFallback {
	return &StaticFallback{
		Sources: sources,
	}
}

func (sf *StaticFallback) Handler(spa bool) request.HandlerFunc {
	return func(ctx *request.Context) error {
		sf.RawHandler(false).ServeHTTP(ctx.Writer, ctx.Request)
		return nil
	}
}

func (sf *StaticFallback) RawHandler(spa bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Remove leading slash for file system access
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		if path == "" || strings.HasSuffix(path, "/") {
			path = path + "index.html"
		}

		// Try each source in order
		for i, s := range sf.Sources {
			f, err := s.Open(path)
			if err != nil {
				continue
			}

			stat, err := f.Stat()
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				f.Close()
				return
			}

			if stat.IsDir() {
				f.Close()
				indexPath := strings.TrimSuffix(path, "/") + "/index.html"
				r2 := *r
				r2.URL.Path = "/" + indexPath
				sf.RawHandler(false).ServeHTTP(w, &r2)
				return
			}

			if ext := filepath.Ext(path); ext != "" {
				if ctype := mime.TypeByExtension(ext); ctype != "" {
					w.Header().Set("Content-Type", ctype)
				}
			}

			if rs, ok := f.(io.ReadSeeker); ok {
				fmt.Printf("[DEBUG] Serving %s from FS source %d\n", path, i)
				http.ServeContent(w, r, path, stat.ModTime(), rs)
			} else {
				http.Error(w, "File does not support seeking", http.StatusInternalServerError)
			}
			f.Close()
			return
		}

		// Not found in any source
		if spa && filepath.Ext(path) == ".html" {
			// Special handling for SPA: fallback to index.html if HTML file not found
			fmt.Printf("[DEBUG] File %s not found, fallback ke index.html\n", path)
			r2 := *r
			r2.URL.Path = "/index.html"
			sf.RawHandler(false).ServeHTTP(w, &r2)
			return
		}

		fmt.Printf("[DEBUG] File %s not found in any source, returning 404\n", path)
		http.NotFound(w, r)
	})
}

func (sf *StaticFallback) ReadFile(file string) ([]byte, error) {
	var lastErr error
	for _, source := range sf.Sources {
		data, err := fs.ReadFile(source, file)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("file %s not found in any source: %w", file, lastErr)
}

func (sf *StaticFallback) Open(file string) (fs.File, error) {
	var lastErr error
	for _, source := range sf.Sources {
		file, err := source.Open(file)
		if err == nil {
			return file, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("file %s not found in any source: %w", file, lastErr)
}

func (sf *StaticFallback) WithSourceDir(dir string) *StaticFallback {
	sf.Sources = append(sf.Sources, os.DirFS(dir))
	return sf
}

func (sf *StaticFallback) WithSourceFS(fs fs.FS) *StaticFallback {
	sf.Sources = append(sf.Sources, fs)
	return sf
}

func (sf *StaticFallback) WithEmbedFS(fsys embed.FS, subFS string) *StaticFallback {
	fs, err := fs.Sub(fsys, subFS)
	if err != nil {
		panic(fmt.Sprintf("Failed to create sub FS from %s: %v\n", subFS, err))
	}
	sf.Sources = append(sf.Sources, fs)
	return sf
}
