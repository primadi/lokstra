
# Static Files

**Source-driven** for `lokstra-0.2.1` (from `core/router/*`, `modules/coreservice/router_engine/*`, and `common/static_files/*`).  
Lokstra can serve **static assets** and **SPA** (single-page app) bundles efficiently with minimal setup.

---

## API

```go
// Router (core/router/router.go)
MountStatic(prefix string, spa bool, sources ...fs.FS) Router

	// MountHtmx ser sources ...fs.FS) Router
```

- `prefix`: URL mount point (e.g., `/static/`, `/`)
- `spa`: if **true**, missing paths **fall back** to `index.html` (client-side routing)
- `sources ...fs.FS`: ordered file systems to read from; first match wins (e.g., `os.DirFS("./public")`, `embed.FS` via `fs.Sub`)

There is also a lower-level helper:

```go
// common/static_files/static_files.go
type StaticFiles struct {
    Sources []fs.FS
}

func New(sources ...fs.FS) *StaticFiles
func (sf *StaticFiles) RawHandler(spa bool) http.Handler
func (sf *StaticFiles) Handler(spa bool) request.HandlerFunc  // wraps RawHandler for Lokstra ctx
```

---

## Build & Runtime Wiring

During router build, each static mount is passed **directly** to the router engine:

```go
// core/router/router_impl.go
for _, sdf := range router.StaticMounts {
    r.r_engine.ServeStatic(sdf.Prefix, sdf.Spa, sdf.Sources...)
}
```

Both built-in engines implement `ServeStatic`:

```go
// modules/coreservice/router_engine/servemux_engine.go
ServeStatic(prefix string, spa bool, sources ...fs.FS) {
	cleanPrefixStr := cleanPrefix(prefix)

	staticServe := static_files.New(sources...)

	handler := staticServe.RawHandler(spa)
	// Strip prefix before passing to fallback handler
	if cleanPrefixStr != "/" {
		handler = http.StripPrefix(cleanPrefixStr, handler)
	}

	if cleanPrefixStr == "/" {
		m.mux.Handle("/", handler)
	} else {
		m.mux.Handle(cleanPrefixStr+"/", handler)
	}
}

// ServeHtmxPage implements serviceapi.RouterEngine.
func (m *ServeMuxEngine) ServeHtmxPage(pageDataRouter http.Handler,
	prefix string, si *static_files.ScriptInjection, sources ...fs.FS) {
	cleanPrefixStr := cleanPrefix(prefix)

	staticServe := static_files.New(sources...)

	handler := staticServe.HtmxPageHandlerWithScriptInjection(pageDataRouter, prefix, si
```

Key details:
- For non-root prefixes, the engine wraps the handler with `http.StripPrefix(prefix, handler)`.
- The engine registers **both** `prefix` and `prefix/` to catch directory-style URLs.

---

## File Resolution & SPA Fallback

The raw handler reads from each source FS in order:

- Request path `"/"` → **serves** `index.html` by default.
- If the requested asset is **found** in any FS → serves that file.
- If **not found**:
  - `spa == false` → returns **404** (`"404 page not found"`).
  - `spa == true` → **falls back** to `index.html` to support client-side routing (React/Vue/etc.).

This is covered by tests in `common/static_files/static_files_test.go`:
- Static mode missing route → **404**.
- SPA mode missing route → serves **index.html**.
- Existing assets like `/assets/app.js` or `/favicon.ico` are served directly in both modes.

---

## YAML (Config)

You can declare mounts in YAML via `MountStaticConfig`:

```go
type MountStaticConfig struct {
	Prefix string   `yaml:"prefix"`
	Spa    bool     `yaml:"spa,omitempty"`
	Folder []string `yaml:"folder"`
}
```

Example (app-level):

```yaml
apps:
  - name: web
    address: ":8080"
    mount_static:
      - prefix: /static/
        spa: false
        folder: ["./static", "./public"]
      - prefix: /
        spa: true
        folder: ["./spa"]       # SPA bundle
```

> `folder` entries are resolved to `os.DirFS` in your bootstrap code before mounting. For embedded assets, use `embed.FS` + `fs.Sub` and call `MountStatic` in code.

---

## Examples

### Static directories

```go
app.MountStatic("/static/", false, os.DirFS("./static"))
app.MountStatic("/assets/", false, os.DirFS("./assets"))
```

### SPA at root

```go
// Fallback to index.html for unknown routes
app.MountStatic("/", true, os.DirFS("./spa"))
```

### Using embed.FS

```go
//go:embed web/*
var webFS embed.FS

app.MountStatic("/web/", false, mustSub(webFS, "web"))

func mustSub(fsys embed.FS, dir string) fs.FS {
    s, err := fs.Sub(fsys, dir)
    if err != nil { panic(err) }
    return s
}
```

---

## Tips & Gotchas

- **Prefix:** include a trailing slash (e.g., `/static/`) for directory mounts. The engine also registers the non-slashed variant for convenience.
- **Ordering:** pass sources **highest priority first**; the first FS containing the file wins.
- **Coexist with APIs:** static mounts can live alongside API routes (`/api/*`). They don’t use Lokstra middleware; if you need auth or other checks, place assets behind routes instead of static mounts.
- **Caching:** rely on your reverse proxy/CDN for aggressive caching. The raw handler serves files directly without additional cache headers unless present in the file system or set upstream.
- **Security:** ensure your FS roots are scoped to the intended directories (use `fs.Sub` with `embed.FS` or `os.DirFS` to avoid directory traversal issues).
