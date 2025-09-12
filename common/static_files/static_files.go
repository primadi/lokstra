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
	"github.com/primadi/lokstra/core/response"
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
		for _, s := range sf.Sources {
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
				// fmt.Printf("[DEBUG] Serving %s from FS source %d\n", path, i)
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
			// fmt.Printf("[DEBUG] SPA fallback: %s not found, load /index.html\n", path)
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
var titleRegex = regexp.MustCompile(`(?i)<title>.*?</title>`)

// Assume sf.Sources has:
//   - "/layouts" for HTML layout templates
//   - "/pages" for HTML page templates
//
// All Request paths will be treated as page requests
func (sf *StaticFiles) HtmxPageHandler(pageDataRouter http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Printf("[DEBUG] Request: %s %s\n", r.Method, r.URL.Path)
		path := r.URL.Path

		// 1. Normalize path ke .html page
		pagePath := strings.TrimPrefix(path, "/")
		if pagePath == "" {
			pagePath = "index.html"
		} else if !strings.HasSuffix(pagePath, ".html") {
			pagePath += ".html"
		}
		pagePath = "pages/" + pagePath

		// 2. Load page file
		pageContent, err := sf.ReadFile(pagePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// 3. Extract layout name
		layoutName := "base.html" // default layout
		if m := layoutRegex.FindSubmatch(pageContent); len(m) > 1 {
			layoutName = string(m[1])
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("LS-Layout", layoutName)

		layoutPath := "layouts/" + layoutName

		// 4. Fetch page-data via internal call
		var data map[string]any
		var dataPath string

		cleanPrefix := "/" + strings.Trim(prefix, "/")
		if cleanPrefix == "/" {
			dataPath = "/page-data" + path
		} else {
			dataPath = "/page-data" + cleanPrefix + path
		}
		dataPath, _ = strings.CutSuffix(dataPath, "/")
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

			var responseData struct {
				Code string            `json:"code"`
				Data response.PageData `json:"data"`
			}

			if err := json.Unmarshal(body, &responseData); err != nil {
				// Handle error
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			data = responseData.Data.Data
			if data == nil {
				data = map[string]any{}
			}
			data["ls-title"] = responseData.Data.Title
			data["ls-description"] = responseData.Data.Description
		case http.StatusNotFound:
			data = map[string]any{}
		default:
			// Propagate error to user
			w.WriteHeader(res.StatusCode)
			io.Copy(w, res.Body)
			return
		}

		// 5. Check if partial render (HTMX) or full render (normal)
		isPartial := r.Header.Get("HX-Request") == "true" &&
			r.Header.Get("LS-Layout") == layoutName

		// 6. Set pageTitle and pageDesc variables for template
		var pageTitle, pageDesc string
		if title, ok := data["ls-title"].(string); ok {
			pageTitle = title
			// fmt.Printf("[DEBUG] PageTitle from data: %s\n", pageTitle)
		}
		if desc, ok := data["ls-description"].(string); ok {
			pageDesc = desc
		}

		tmpl := template.New("")

		if isPartial {
			// Partial render → only page template
			_, err = tmpl.Parse(string(pageContent))
			if err != nil {
				http.Error(w, fmt.Sprintf("template parse error: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("HX-Partial", "true")
			w.Header().Set("LS-Title", pageTitle)
			w.Header().Set("LS-Description", pageDesc)
			_ = tmpl.Execute(w, data)
			return
		}

		// 7. Full render → load layout + page
		layoutContent, err := sf.ReadFile(layoutPath)
		if err != nil {
			http.Error(w, "layout not found: "+layoutPath, http.StatusInternalServerError)
			return
		}

		strLayoutContent := string(layoutContent)

		var metaElements string
		var titleElement string

		if pageTitle != "" {
			// Always prioritize title injection at the very beginning
			if titleRegex.MatchString(strLayoutContent) {
				// Replace existing title
				strLayoutContent = titleRegex.ReplaceAllString(strLayoutContent, "")
			}
			titleElement = fmt.Sprintf(`<title>%s</title>`, pageTitle) + "\n"
		} else if !titleRegex.MatchString(strLayoutContent) {
			// Add default title if no pageTitle and no existing title
			titleElement = `<title>Lokstra App</title>` + "\n"
		}

		if pageDesc != "" {
			metaElements += fmt.Sprintf(`<meta name="description" content="%s">`, pageDesc) + "\n"
		}
		metaElements += fmt.Sprintf(`<meta name="ls-layout" content="%s">`, layoutName) + "\n"

		layoutSwitcherScript := `<script>
			// This script is automatically injected by Lokstra HTMX Mount.
			document.addEventListener("DOMContentLoaded", function () {
				// Inject LS-Layout header for every htmx request
				document.body.addEventListener("htmx:configRequest", function (evt) {
					var layoutMeta = document.querySelector('meta[name="ls-layout"]')
					var layoutName = layoutMeta ? layoutMeta.content : "base.html"
					evt.detail.headers["LS-Layout"] = layoutName
				})

				// Handle layout changes by full page reload, if layout differs
				document.body.addEventListener("htmx:beforeSwap", function (evt) {
					var layoutMeta = document.querySelector('meta[name="ls-layout"]')
					var currentLayout = layoutMeta ? layoutMeta.content : "base.html"
					var responseLayout = evt.detail.xhr.getResponseHeader("LS-Layout")
					if (responseLayout && responseLayout !== currentLayout) {
						evt.preventDefault()
						window.location.href = evt.detail.pathInfo.finalRequestPath || 
							window.location.pathname
					}
				})

				document.body.addEventListener("htmx:afterSwap", function () {
					var xhr = window.event.detail.xhr
					var titleMeta = xhr.getResponseHeader("LS-Title")
					var descMeta =  xhr.getResponseHeader("LS-Description")

					if (titleMeta) {
						document.title = titleMeta;
					}
					if (descMeta) {
						// update or replace meta[name="description"] in <head>
						var currentDescMeta = document.head.querySelector('meta[name="description"]')
						if (currentDescMeta) {
							currentDescMeta.content = descMeta
						} else {
							var newDescMeta = document.createElement('meta')
							newDescMeta.name = "description"
							newDescMeta.content = descMeta
							document.head.appendChild(newDescMeta)
						}
					}
				})
			});
		</script>`

		// Inject title FIRST, then other meta elements
		var headInjection string
		if titleElement != "" {
			headInjection = titleElement
		}
		headInjection += metaElements

		strLayoutContent = strings.Replace(strLayoutContent, "<head>",
			"<head>\n"+headInjection, 1)

		strLayoutContent = strings.Replace(strLayoutContent, "</body>",
			layoutSwitcherScript+"\n</body>", 1)

		// 8. Define layout & page templates
		_, err = tmpl.Parse(strLayoutContent)
		if err != nil {
			http.Error(w, fmt.Sprintf("layout parse error: %v", err),
				http.StatusInternalServerError)
			return
		}
		_, err = tmpl.New("page").Parse(string(pageContent))
		if err != nil {
			http.Error(w, fmt.Sprintf("page parse error: %v", err),
				http.StatusInternalServerError)
			return
		}

		// Buffer the response to send it atomically with proper Content-Length
		var buf strings.Builder
		err = tmpl.Execute(&buf, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("template execute error: %v", err),
				http.StatusInternalServerError)
			return
		}

		// Send complete response at once
		html := buf.String()
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(html)))
		w.Write([]byte(html))
	})
}
