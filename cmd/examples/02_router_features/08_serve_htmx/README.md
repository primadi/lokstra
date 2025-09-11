# HTMX Pages Example - Lokstra Framework

This example demonstrates how to serve HTMX pages with layout support using Lokstra's `MountHtmx` feature.

## Features Demonstrated

### 🎨 HTMX Page Serving
- **Layout-based rendering** with `<!-- layout: layout_name.html -->` directive
- **Partial rendering** for HTMX requests (when `HX-Request` header is present)
- **Page data injection** from `/page-data/*` API endpoints
- **Static asset serving** with proper MIME types and fallback chain

### 🗂️ Directory Structure
```
htmx_app/
├── static/           # Static assets (CSS, JS, images)
│   ├── style.css     # Main stylesheet
│   └── app.js        # JavaScript functionality
├── layouts/          # HTML layout templates
│   └── base.html     # Default layout template
└── pages/            # HTML page templates
    ├── index.html    # Home page
    ├── about.html    # About page
    ├── products.html # Products page
    └── contact.html  # Contact form page
```

### 🔄 How It Works

1. **Route Handling**: All requests are handled by the HTMX page handler
2. **Static Assets**: Requests to `/static/*` are served directly from the filesystem
3. **Page Rendering**: Other requests are treated as page requests:
   - Path `/about` → loads `pages/about.html`
   - Path `/` → loads `pages/index.html`
   - Path `/products` → loads `pages/products.html`

4. **Layout System**: 
   - Pages can specify layout with `<!-- layout: base.html -->` comment
   - Default layout is `base.html` if not specified
   - Layouts are loaded from `layouts/` directory

5. **Data Injection**:
   - Page data is fetched from `/page-data{path}` endpoints
   - Data is injected as `{{.PageData}}` in templates
   - JSON response is available in templates

6. **Partial Rendering**:
   - Full page load: Layout + Page content
   - HTMX request: Only page content (if `HX-Request` header present and layout matches)

## Running the Example

```bash
# Navigate to the example directory
cd cmd/examples/02_router_features/08_serve_htmx

# Run the application
go run main.go
```

The server will start on `:8080` with the following endpoints:

### 📄 Page Endpoints
- `http://localhost:8080/` - Home page
- `http://localhost:8080/about` - About page  
- `http://localhost:8080/products` - Products page
- `http://localhost:8080/contact` - Contact form page

### 📁 Static Assets
- `http://localhost:8080/static/style.css` - Stylesheet
- `http://localhost:8080/static/app.js` - JavaScript

### 🔌 API Endpoints
- `/page-data/*` - Page data endpoints (called internally)
- `/api/contact` - Contact form submission
- `/api/products/{id}` - Product details
- `/api/info` - Server information
- `/health` - Health check

## Page Template Format

### Page Template Example
```html
<!-- layout: base.html -->
<div class="page-content fade-in">
    <h1>{{.title}}</h1>
    <p>{{.message}}</p>
    
    <!-- HTMX interactions -->
    <button hx-get="/api/products/1" hx-target="#content">
        Load Content
    </button>
    <div id="content"></div>
</div>
```

### Layout Template Example
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.title}}</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
</head>
<body>
    <header>Navigation here</header>
    <main>{{template "page" .}}</main>
    <footer>Footer here</footer>
</body>
</html>
```

### Page Data API Example
```go
app.GET("/page-data/about", func(ctx *lokstra.Context) error {
    return ctx.Ok(map[string]any{
        "title": "About Us",
        "team": []map[string]string{
            {"name": "Alice", "role": "Developer"},
            {"name": "Bob", "role": "Designer"},
        },
    })
})
```

## Multiple Source Fallback

The example demonstrates fallback between multiple sources:

```go
app.MountHtmx("/", 
    os.DirFS("./htmx_content"),  // Project overrides (highest priority)
    subProjectHtmx,              // Project HTMX content  
    subHtmxApp,                  // Embedded HTMX app (lowest priority)
)
```

**Fallback Order:**
1. `./htmx_content/` - Local project overrides
2. `project/htmx/` - Embedded project content
3. `htmx_app/` - Default embedded content

## HTMX Features Used

### Navigation with History
```html
<a href="/about" hx-get="/about" hx-target="main" hx-push-url="true">About</a>
```

### Form Submission
```html
<form hx-post="/api/contact" hx-target="#response">
    <input name="email" type="email" required>
    <button type="submit">Send</button>
</form>
```

### Dynamic Content Loading
```html
<button hx-get="/api/products/1" hx-target="#details">
    Load Details
</button>
<div id="details"></div>
```

### Loading Indicators
```html
<button hx-get="/api/data">
    Load Data
    <span class="htmx-indicator">Loading...</span>
</button>
```

## Customization

### Adding New Pages
1. Create page template in `htmx_app/pages/newpage.html`
2. Add page data endpoint: `/page-data/newpage`
3. Optionally create custom layout in `htmx_app/layouts/`

### Custom Layouts
1. Create layout file in `htmx_app/layouts/custom.html`
2. Use in page: `<!-- layout: custom.html -->`
3. Include `{{template "page" .}}` in layout where page content should appear

### Static Assets
- Add CSS/JS files to `htmx_app/static/`
- Reference in layouts: `/static/filename.css`
- Automatic MIME type detection and caching headers

## Architecture Benefits

- **SEO Friendly**: Full page loads work without JavaScript
- **Fast Navigation**: HTMX provides SPA-like experience
- **Progressive Enhancement**: Works with and without JavaScript
- **Server-Side Rendering**: All templates rendered on server
- **Flexible Layouts**: Multiple layouts supported per application
- **Fallback Support**: Multiple content sources with priority
