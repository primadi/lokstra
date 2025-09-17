
# HTMX Integration

**Source-driven** for `lokstra-0.2.1` (derived from `common/static_files/*`, `serviceapi/router_engine.go`, `core/router/*`, `core/request/context.go`, `core/response/helper.go`, and examples under `cmd/examples/02_router_features/*`).

Lokstra provides a high-level way to serve **HTML pages** with **layouts**, **page-data** endpoints, and **HTMX partial updates**. You mount an HTMX site at a URL prefix, point it to one or more static template sources, and define page-data handlers that return structured data. Lokstra does the rest: layout resolution, data fetch, script injection, and partial vs. full rendering.

---

## Mental Model

- **Pages & Layouts** come from your static FS sources:
  - `pages/*.html` files are *content pages*
  - `layouts/*.html` files are *master layouts*
  - A page can pick its layout via a comment like: `<!-- layout: base.html -->` (default: `base.html`)
  - A page can set a title via: `<!-- title: My Page -->`

- **Page-data** endpoints are your JSON-ish providers:
  - For each incoming page request, Lokstra internally calls an HTTP endpoint to fetch `PageData`.
  - Your handler returns `response.PageData` (Title, Description, Data) via `ctx.HtmxPageData(...)`.

- **Partial vs Full Render** is automatic:
  - If the request is an **HTMX** request (`HX-Request: true`) **and** the request’s `LS-Layout` header matches the page’s layout, Lokstra renders **page only** and returns headers: `HX-Partial: true`, `LS-Title`, `LS-Description`.
  - Otherwise, Lokstra renders **layout + page**, injects scripts, `<title>`, `<meta name="description">`, and `<meta name="ls-layout">`.

- **Client runtime** is handled by an auto-injected JS snippet:
  - Adds `LS-Layout` header to every HTMX request (so server can decide partial vs full).
  - If server responds with a **different** `LS-Layout`, the client forces a **full page reload**.
  - Optional animation hooks are included by default (fade-in on swap) — can be customized/disabled.

---

## Server API (what you write)

### Mounting an HTMX site

```go
// Router interface (core/router/router.go)
MountHtmx(prefix string, si *static_files.ScriptInjection, sources ...fs.FS) Router
```

- `prefix` — where the site lives (e.g. `/`, `/admin`)
- `si` — optional script injection; pass `nil` to use **default** (`default` + `animation`)
- `sources ...fs.FS` — ordered list of file systems (first one wins); each must contain:
  - `/layouts/*.html`
  - `/pages/*.html`

**Example (from examples/06_mount_htmx):**

```go
app.MountHtmx("/", nil, webAppFS)      // main site at /
app.MountHtmx("/admin", nil, adminFS)  // admin site at /admin
```

### Providing page-data

Create a group at `/page-data` and define handlers that return `PageData`:

```go
pageData := app.Group("/page-data")

pageData.GET("/", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("Home Page", "", map[string]any{
        "message": "Welcome to Lokstra HTMX Demo",
    })
})

pageData.GET("/about", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("About Us", "About page with dynamic content", map[string]any{
        "team": []map[string]string{{ {"name":"Alice"}, {"name":"Bob"} }},
    })
})
```

> **How the URL is computed:** For a request to `{prefix}/foo/bar`,
> Lokstra internally calls:
> - `/page-data{prefix}/foo/bar`  (if `prefix != "/"`)
> - `/page-data/foo/bar`          (if `prefix == "/"`)

So for `prefix="/admin"` and request `/admin/users`, the page-data endpoint must be `/page-data/admin/users`.

### Returning raw HTML snippets

For ad-hoc HTMX endpoints (not tied to the page/layout pipeline), return HTML directly:

```go
func (ctx *lokstra.Context) HTMX(html string) error  // sets "Vary: HX-Request" and renders HTML
```

### Response helpers you will use

```go
// response.PageData (core/response/helper.go)
type PageData struct {
    Title       string            `json:"title,omitempty"`
    Description string            `json:"description,omitempty"`
    Data        map[string]any    `json:"data,omitempty"`
}

// Send page-data
func (r *Response) HtmxPageData(title, description string, data map[string]any) error

// Send partial or full HTML directly
func (ctx *request.Context) HTMX(html string) error               // partial snippet
func (r *Response) HTML(html string) error                        // full HTML (200 OK)
func (r *Response) ErrorHTML(status int, html string) error       // HTML error
```

---

## Template & Filesystem Rules

- **Where to put files** (per source FS):
  - `layouts/base.html`, `layouts/admin.html`, etc.
  - `pages/index.html`, `pages/about.html`, `pages/products.html`, etc.
- **Choose layout per page** (optional):
  ```html
  <!-- layout: base.html -->
  <!-- title: Products -->
  <div id="content">
    <!-- your page body; use {{ .foo }} from PageData.Data -->
  </div>
  ```
- **Fallback order**: multiple FS sources are searched **in order**; first match wins.
  Useful to overlay a project-specific FS over a shared embedded FS.

- **SPA static**: for static mounts (`MountStatic(prefix, spa, ...sources)`), SPA mode falls back to `index.html` when a file is missing; **HTMX mount** uses the page/layout mechanism instead.

---

## How the Runtime Works (Step-by-step)

What happens when the browser requests `GET {prefix}/path`:

1) **Normalize page path** → `pages/{path}.html` (empty path → `pages/index.html`).
2) **Read page file** (searching ordered sources). If not found → `404`.
3) **Extract layout** from `<!-- layout: ... -->` (default `base.html`) and **title** from `<!-- title: ... -->`.
4) **Fetch page-data** by internally calling an HTTP endpoint:
   - URL: `/page-data{prefix}{path}` (or `/page-data{path}` when `prefix="/"`)
   - Lokstra makes an in-memory HTTP request to the *same router engine* to run your handlers.
   - On `200`: parse JSON `{ code, data: PageData }` and keep `Title`, `Description`, and `Data`.
   - On `404`: treat as `Data = {}`.
   - On other errors: propagate status/body to client.
5) **Partial vs Full decision**:
   - Partial if `HX-Request: true` **and** `LS-Layout` (request header) equals the page’s layout.
   - **Partial render**:
     - Parse the **page** template only, execute with `Data`.
     - Set headers: `HX-Partial: true`, `LS-Title`, `LS-Description`.
   - **Full render**:
     - Load `layouts/{layoutName}` and run **script injection** (see next section).
     - Ensure `<title>` and `<meta name="description">` and add `<meta name="ls-layout" content="{layoutName}">`.
     - Parse layout, then parse page as `tmpl.New("page")`, execute with `Data`.
6) **Respond** with `text/html; charset=utf-8`. For full render Lokstra writes a `Content-Length` header.

---

## Script Injection (defaults & customization)

Lokstra injects small scripts to enable HTMX conventions and UX niceties.

- **Default bundle** (when `si == nil`):
  - `scripts/default/body_end_htmx.js`:
    - Adds `LS-Layout` header on every HTMX request (`htmx:configRequest` listener).
    - If a response header `LS-Layout` is **different** from current layout, forces a **full reload**.
    - Handles OOB swap animation hooks.
  - `scripts/animation/body_end_fade_in.js`:
    - Adds a fade-in animation on swaps.

- **Customize** with `*static_files.ScriptInjection`:
  ```go
  si := static_files.NewScriptInjection().
        AddNamedScriptInjection("default").        // include default htmx helpers
        AddHeadEndScript(`<script>console.log("custom head")</script>`).
        AddBodyEndScript(`<script src="/static/extra.js"></script>`)

  app.MountHtmx("/admin", si, adminFS)
  ```

  Helpers available:
  - `NewScriptInjection()`
  - `AddNamedScriptInjection(name string)` → loads embedded scripts under `common/static_files/scripts/{name}/`
  - `AddHeadStartScript(string)`, `AddHeadEndScript(string)`, `AddBodyEndScript(string)`
  - `NewDefaultScriptInjection(enableAnimation bool)` → default + (optional) animation

---

## YAML Configuration

You can wire HTMX mounts via YAML as part of apps and groups (see `core/config/types.go`).

```yaml
apps:
  - name: web
    address: ":8080"
    mount_htmx:
      - prefix: /
        sources:            # ordered priority, first wins
          - ./web_htmx      # e.g., OS dir
          - embed://htmx    # e.g., module-provided embed.FS (resolved in code)
    routes: []              # your /page-data routes go in code (see below)
```

> The **page-data** endpoints are regular routes you register (usually under `/page-data`). They are not declared in YAML; you implement them in Go.

Static files & reverse proxy YAML (for completeness):
```yaml
mount_static:
  - prefix: /static/
    spa: false
    folder: [./static, ./public]

mount_reverse_proxy:
  - prefix: /api/
    target: http://localhost:9000
```

---

## Example Project Layout

```
web_htmx/
  layouts/
    base.html
    admin.html
  pages/
    index.html          <!-- layout: base.html -->
    about.html          <!-- layout: base.html -->
    products.html       <!-- layout: base.html -->
admin_htmx/
  layouts/
    admin.html
  pages/
    index.html          <!-- layout: admin.html -->
```

---

## Gotchas & Tips

- **Define `/page-data` routes** for every page path you serve (including `/` → `/page-data/`). For nested HTMX mounts, include the mount prefix: `/page-data/admin/...`.
- **First source wins**: pass sources in priority order to `MountHtmx` (project FS first; embedded FS fallback).
- **Partial rendering requires matching layout**: if you change layout via navigation, HTMX will trigger a full reload automatically (via `LS-Layout` check in the client script).
- **Embedding data**: your templates can render `{{ .key }}` from `PageData.Data` and can read title/description injected into `<head>`.
- **Raw fragment endpoints**: use `ctx.HTMX(html)` when you’re not using the page/layout pipeline.
- **Performance**: the server constructs templates per request; consider caching if your layouts/pages are large and stable.

---

## Minimal End-to-End Sample

```go
// 1) Mount site
app.MountHtmx("/", nil, os.DirFS("./web_htmx"))

// 2) Page-data
pd := app.Group("/page-data")
pd.GET("/", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("Home", "Welcome", map[string]any{
        "now": time.Now().Format(time.RFC3339),
    })
})
pd.GET("/about", func(ctx *lokstra.Context) error {
    return ctx.HtmxPageData("About", "", map[string]any{"team": []string{"Alice","Bob"}})
})
```

```html
<!-- web_htmx/layouts/base.html -->
<!doctype html><html><head></head><body>
  <header>My Site</header>
  <main hx-boost="true">
    {{ template "page" . }}
  </main>
</body></html>
```

```html
<!-- web_htmx/pages/about.html -->
<!-- layout: base.html -->
<!-- title: About -->
<section>
  <h1>About</h1>
  <ul>
    {{ range .team }}<li>{{ . }}</li>{{ end }}
  </ul>
</section>
```

That’s it: full renders for navigation that changes layout; partial updates when HTMX swaps within the same layout.
