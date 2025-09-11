package static_files

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/primadi/lokstra/common/json"
	"github.com/primadi/lokstra/core/request"
)

type StaticFiles struct {
	Sources []fs.FS
}

// New creates a new StaticFiles instance.
func New(sources ...fs.FS) *StaticFiles {
	return &StaticFiles{
		Sources: sources,
	}
}

func (sf *StaticFiles) Handler(spa bool) request.HandlerFunc {
	return func(ctx *request.Context) error {
		sf.RawHandler(false).ServeHTTP(ctx.Writer, ctx.Request)
		return nil
	}
}

func openFileAndStats(s fs.FS, path string) (fs.File, fs.FileInfo, error) {
	f, err := s.Open(path)
	if err != nil {
		return nil, nil, err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, nil, err
	}
	return f, stat, nil
}

func (sf *StaticFiles) RawHandler(spa bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// Remove leading slash for file system access
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}

		if path == "" {
			path = path + "index.html"
		}

		// Try each source in order
		for i, s := range sf.Sources {
			f, stat, err := openFileAndStats(s, path)
			if err != nil {
				continue
			}

			if stat.IsDir() {
				f.Close()
				indexPath := strings.TrimSuffix(path, "/") + "/index.html"
				r2 := *r
				r2.URL.Path = "/" + indexPath
				sf.RawHandler(false).ServeHTTP(w, &r2)
				return
			}

			if rs, ok := f.(io.ReadSeeker); ok {
				fmt.Printf("[DEBUG] Serving %s from FS source %d\n", path, i)
				http.ServeContent(w, r, path, stat.ModTime(), rs)
				f.Close()
			} else {
				f.Close()
				http.Error(w, "File does not support seeking", http.StatusInternalServerError)
			}
			return
		}

		// Not found in any source - SPA fallback to /index.html
		if spa && filepath.Ext(path) == "" {
			fmt.Printf("[DEBUG] SPA fallback: %s not found, load /index.html\n", path)
			r2 := *r
			r2.URL.Path = "/index.html"
			sf.RawHandler(false).ServeHTTP(w, &r2)
			return
		}

		fmt.Printf("[DEBUG] File %s not found in any source, returning 404\n", path)
		http.NotFound(w, r)
	})
}

func (sf *StaticFiles) ReadFile(file string) ([]byte, error) {
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

func (sf *StaticFiles) Open(file string) (fs.File, error) {
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

func (sf *StaticFiles) WithSourceDir(dir string) *StaticFiles {
	sf.Sources = append(sf.Sources, os.DirFS(dir))
	return sf
}

func (sf *StaticFiles) WithSourceFS(fs fs.FS) *StaticFiles {
	sf.Sources = append(sf.Sources, fs)
	return sf
}

func (sf *StaticFiles) WithEmbedFS(fsys embed.FS, subFS string) *StaticFiles {
	fs, err := fs.Sub(fsys, subFS)
	if err != nil {
		panic(fmt.Sprintf("Failed to create sub FS from %s: %v\n", subFS, err))
	}
	sf.Sources = append(sf.Sources, fs)
	return sf
}

// regex to detect layout directive in HTML comments
var layoutRegex = regexp.MustCompile(`<!--\s*layout:\s*([a-zA-Z0-9_\-./]+)\s*-->`)

// Assume sf.Sources has:
//   - "/static" for static assets (CSS, JS, images)
//   - "/layouts" for HTML layout templates
//   - "/pages" for HTML page templates
//
// All Request paths will be treated as page requests, except those starting with /static/
// which will be treated as static asset requests.
func (sf *StaticFiles) HtmxPageHandler(pageDataRouter http.Handler,
	staticFolders []string) http.Handler {
	normalizeStaticFolders := make([]string, 0, len(staticFolders))
	for _, f := range staticFolders {
		f = "/" + strings.Trim(f, "/") + "/"
		normalizeStaticFolders = append(normalizeStaticFolders, f)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 1. Serve static assets directly
		for _, folder := range normalizeStaticFolders {
			if strings.HasPrefix(path, folder) {
				sf.RawHandler(false).ServeHTTP(w, r)
				return
			}
		}

		// 2. Normalize path ke .html page
		pagePath := strings.TrimPrefix(path, "/")
		if pagePath == "" {
			pagePath = "index.html"
		} else if !strings.HasSuffix(pagePath, ".html") {
			pagePath += ".html"
		}
		pagePath = "pages/" + pagePath

		// 3. Load page file
		pageContent, err := sf.ReadFile(pagePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// 4. Extract layout name
		layoutName := "base.html" // default layout
		if m := layoutRegex.FindSubmatch(pageContent); len(m) > 1 {
			layoutName = string(m[1])
		}
		layoutPath := "layouts/" + layoutName

		// 5. Fetch page-data via internal call
		var data map[string]any
		var dataPath string
		if path == "/" || path == "" {
			dataPath = "/page-data"
		} else {
			dataPath = "/page-data" + path
		}
		req := httptest.NewRequest(http.MethodGet, dataPath, nil)
		req.Header = r.Header.Clone()
		req.URL.RawQuery = r.URL.RawQuery
		rr := httptest.NewRecorder()
		pageDataRouter.ServeHTTP(rr, req)
		res := rr.Result()
		defer res.Body.Close()

		switch res.StatusCode {
		case http.StatusOK:
			body, _ := io.ReadAll(res.Body)

			var pageData struct {
				Code string         `json:"code"`
				Data map[string]any `json:"data"`
			}

			if err := json.Unmarshal(body, &pageData); err != nil {
				// Handle error
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			data = pageData.Data
		case http.StatusNotFound:
			data = map[string]any{}
		default:
			// Propagate error to user
			w.WriteHeader(res.StatusCode)
			io.Copy(w, res.Body)
			return
		}

		// 6. Cek if partial render (HTMX) or full render (normal)
		isPartial := r.Header.Get("HX-Request") == "true" &&
			r.Header.Get("LS-Layout") == layoutName

		tmpl := template.New("")

		if isPartial {
			// Partial render → only page template
			_, err = tmpl.Parse(string(pageContent))
			if err != nil {
				http.Error(w, fmt.Sprintf("template parse error: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("HX-Partial", "true")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_ = tmpl.Execute(w, data)
			return
		}

		// Full render → load layout + page
		layoutContent, err := sf.ReadFile(layoutPath)
		if err != nil {
			http.Error(w, "layout not found: "+layoutPath, http.StatusInternalServerError)
			return
		}

		// Define layout & page templates
		_, err = tmpl.Parse(string(layoutContent))
		if err != nil {
			http.Error(w, fmt.Sprintf("layout parse error: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = tmpl.New("page").Parse(string(pageContent))
		if err != nil {
			http.Error(w, fmt.Sprintf("page parse error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tmpl.Execute(w, data)
	})
}
