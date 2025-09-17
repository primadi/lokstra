# HTMX Integration

Lokstra provides first-class support for HTMX applications, allowing you to build modern, dynamic web interfaces without complex JavaScript frameworks. The HTMX integration includes automatic layout management, page data handling, script injection, and seamless partial rendering.

## HTMX Overview

HTMX extends HTML with attributes that enable:
- **AJAX requests** directly from HTML elements
- **Partial page updates** without full page reloads
- **WebSocket and SSE support** for real-time features
- **CSS transitions** for smooth user experiences
- **History management** with back/forward button support

Lokstra's HTMX integration provides:
- **Layout-based templating** with automatic inheritance
- **Page data endpoints** for dynamic content
- **Script injection** for enhanced functionality
- **Multi-source file systems** with override capabilities
- **Partial rendering** optimization for HTMX requests

## Basic HTMX Setup

### Mounting HTMX Pages

```go
import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/common/static_files"
    "embed"
)

//go:embed htmx_app/*
var htmxFS embed.FS

func setupHTMX(app *lokstra.App) {
    // Create static file handler with multiple sources
    sf := static_files.New().
        WithSourceDir("./htmx_override").   // Local overrides (highest priority)
        WithEmbedFS(htmxFS, "htmx_app")     // Embedded fallback

    // Optional script injection
    scriptInjection := static_files.NewScriptInjection().
        AddNamedScriptInjection("default")

    // Mount HTMX pages at root
    app.MountHtmx("/", scriptInjection, sf.Sources...)
    
    // Mount static assets
    app.MountStatic("/static/", false, sf.Sources...)
}
```

### Directory Structure

Organize your HTMX application files:

```
htmx_app/
├── layouts/
│   ├── base.html          # Default layout
│   ├── admin.html         # Admin layout
│   └── mobile.html        # Mobile layout
├── pages/
│   ├── index.html         # Home page
│   ├── about.html         # About page
│   ├── products.html      # Products page
│   └── admin/
│       └── dashboard.html # Admin pages
└── static/
    ├── css/
    │   └── style.css
    ├── js/
    │   └── app.js
    └── images/
        └── logo.png
```

## Layout System

### Creating Layouts

Layouts define the overall page structure and include the page content:

```html
<!-- layouts/base.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <!-- Title and meta tags are automatically injected -->
    <link rel="stylesheet" href="/static/css/style.css">
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
</head>
<body>
    <header>
        <div class="container">
            <div class="logo">My App</div>
            <nav>
                <ul>
                    <li><a href="/" hx-get="/" hx-target="main" hx-push-url="true">Home</a></li>
                    <li><a href="/about" hx-get="/about" hx-target="main" hx-push-url="true">About</a></li>
                    <li><a href="/products" hx-get="/products" hx-target="main" hx-push-url="true">Products</a></li>
                </ul>
            </nav>
        </div>
    </header>

    <main class="container">
        {{template "page" .}}  <!-- Page content is inserted here -->
    </main>

    <footer>
        <div class="container">
            <p>&copy; 2025 My App. Built with Lokstra Framework.</p>
        </div>
    </footer>

    <!-- Loading indicator for HTMX requests -->
    <div class="htmx-indicator" style="position: fixed; top: 10px; right: 10px;">
        Loading... <div class="spinner"></div>
    </div>
</body>
</html>
```

### Admin Layout

Create specialized layouts for different sections:

```html
<!-- layouts/admin.html -->
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="stylesheet" href="/static/css/style.css">
    <link rel="stylesheet" href="/static/css/admin.css">
    <script src="https://unpkg.com/htmx.org@1.9.6"></script>
</head>
<body class="admin-theme">
    <header class="admin-header">
        <div class="container">
            <div class="logo">🛡️ Admin Portal</div>
            <nav>
                <ul>
                    <li><a href="/admin" hx-get="/admin" hx-target="main" hx-push-url="true">Dashboard</a></li>
                    <li><a href="/admin/users" hx-get="/admin/users" hx-target="main" hx-push-url="true">Users</a></li>
                    <li><a href="/admin/settings" hx-get="/admin/settings" hx-target="main" hx-push-url="true">Settings</a></li>
                </ul>
            </nav>
        </div>
    </header>

    <main class="admin-content">
        {{template "page" .}}
    </main>

    <footer class="admin-footer">
        <div class="container">
            <p>&copy; 2025 Admin Portal. Secured interface.</p>
        </div>
    </footer>
</body>
</html>
```

## Page Templates

### Basic Page Template

Pages contain the content specific to each route:

```html
<!-- pages/index.html -->
<!-- layout: base.html -->
<div class="hero-section">
    <h1>Welcome to {{.title}}</h1>
    <p class="subtitle">{{.subtitle}}</p>
    <p class="description">{{.description}}</p>
    
    <div class="quick-actions">
        <button class="btn primary" hx-get="/page-data/featured-products" hx-target="#dynamic-content">
            Load Featured Products
        </button>
        <button class="btn secondary" hx-get="/page-data/latest-news" hx-target="#dynamic-content">
            Load Latest News
        </button>
    </div>
    
    <div id="dynamic-content" class="dynamic-section">
        <p class="placeholder">Click a button above to load dynamic content!</p>
    </div>
</div>

<div class="features-grid">
    <div class="feature-card">
        <h3>🚀 Fast Performance</h3>
        <p>Built with Go for lightning-fast response times.</p>
    </div>
    <div class="feature-card">
        <h3>🔧 Flexible Routing</h3>
        <p>Powerful routing system with middleware support.</p>
    </div>
    <div class="feature-card">
        <h3>🎨 HTMX Integration</h3>
        <p>Modern, dynamic web applications without complex JavaScript.</p>
    </div>
</div>
```

### Dynamic Page with Data

Pages can access data passed from page data handlers:

```html
<!-- pages/products.html -->
<!-- layout: base.html -->
<div class="page-content">
    <h1>{{.title}}</h1>
    <p>{{.description}}</p>
    
    <div class="product-grid">
        {{range .products}}
        <div class="product-card">
            <h3>{{.name}}</h3>
            <div class="product-price">${{.price}}</div>
            <p>{{.description}}</p>
            <button 
                class="btn" 
                hx-get="/api/products/{{.id}}" 
                hx-target="#product-details-{{.id}}"
                hx-swap="innerHTML">
                View Details
            </button>
            <div id="product-details-{{.id}}" class="product-details hidden">
                <!-- Details loaded via HTMX -->
            </div>
        </div>
        {{end}}
    </div>
    
    <div class="pagination">
        <button 
            class="btn" 
            hx-get="/page-data/products?page={{.prev_page}}" 
            hx-target="main"
            hx-push-url="true"
            {{if not .has_prev}}disabled{{end}}>
            Previous
        </button>
        <span>Page {{.current_page}} of {{.total_pages}}</span>
        <button 
            class="btn" 
            hx-get="/page-data/products?page={{.next_page}}" 
            hx-target="main"
            hx-push-url="true"
            {{if not .has_next}}disabled{{end}}>
            Next
        </button>
    </div>
</div>
```

### Admin Page Template

```html
<!-- pages/admin/dashboard.html -->
<!-- layout: admin.html -->
<div class="admin-dashboard">
    <h1>Admin Dashboard</h1>
    <p class="subtitle">{{.description}}</p>
    
    <div class="admin-grid">
        <div class="admin-card">
            <h3>📊 Server Stats</h3>
            <div id="server-stats">
                <button class="btn primary" hx-get="/page-data/admin/server-stats" hx-target="#server-stats">
                    Load Statistics
                </button>
            </div>
        </div>
        
        <div class="admin-card">
            <h3>👥 User Management</h3>
            <div id="user-management">
                <button class="btn secondary" hx-get="/page-data/admin/users" hx-target="#user-management">
                    Load User List
                </button>
            </div>
        </div>
        
        <div class="admin-card">
            <h3>🔧 System Config</h3>
            <div id="system-config">
                <p>Current status: {{.system_status}}</p>
                <button class="btn" hx-post="/api/admin/refresh-config" hx-target="#system-config">
                    Refresh Configuration
                </button>
            </div>
        </div>
    </div>
</div>
```

## Page Data Handlers

### Basic Page Data

Create handlers that provide data for your pages:

```go
func setupPageData(app *lokstra.App) {
    // Create page data routes
    pageData := app.Group("/page-data")
    
    // Home page data
    pageData.GET("/", func(ctx *lokstra.Context) error {
        return ctx.HtmxPageData(
            "Welcome to My App",                    // Page title
            "The best web framework for Go",       // Description
            map[string]any{                        // Page data
                "subtitle": "Build amazing web applications",
                "features": []string{
                    "Fast performance",
                    "HTMX integration", 
                    "Easy to use",
                },
            },
        )
    })
    
    // About page data
    pageData.GET("/about", func(ctx *lokstra.Context) error {
        return ctx.HtmxPageData(
            "About Us",
            "Learn about our company and team",
            map[string]any{
                "team": []map[string]string{
                    {"name": "Alice Johnson", "role": "Lead Developer"},
                    {"name": "Bob Smith", "role": "UI Designer"},
                    {"name": "Carol White", "role": "DevOps Engineer"},
                },
                "founded": 2020,
                "location": "San Francisco, CA",
            },
        )
    })
}
```

### Dynamic Page Data

Handle query parameters and dynamic content:

```go
func setupDynamicPageData(app *lokstra.App) {
    pageData := app.Group("/page-data")
    
    // Products with pagination
    pageData.GET("/products", func(ctx *lokstra.Context) error {
        page := ctx.GetQueryParam("page", "1")
        pageNum, _ := strconv.Atoi(page)
        
        // Get products from database
        products, total, err := productService.GetPaginated(ctx.Context, pageNum, 10)
        if err != nil {
            return err
        }
        
        totalPages := (total + 9) / 10
        
        return ctx.HtmxPageData(
            "Our Products",
            "Discover our amazing product lineup",
            map[string]any{
                "products":     products,
                "current_page": pageNum,
                "total_pages":  totalPages,
                "has_prev":     pageNum > 1,
                "has_next":     pageNum < totalPages,
                "prev_page":    pageNum - 1,
                "next_page":    pageNum + 1,
            },
        )
    })
    
    // Featured products (for HTMX requests)
    pageData.GET("/featured-products", func(ctx *lokstra.Context) error {
        products, err := productService.GetFeatured(ctx.Context, 6)
        if err != nil {
            return err
        }
        
        return ctx.HtmxPageData(
            "Featured Products",
            "",
            map[string]any{
                "products": products,
            },
        )
    })
}
```

### Admin Page Data

Secure admin pages with middleware:

```go
func setupAdminPageData(app *lokstra.App) {
    // Admin page data with authentication
    adminData := app.Group("/page-data/admin", "auth", "admin-only")
    
    // Dashboard data
    adminData.GET("/", func(ctx *lokstra.Context) error {
        stats, err := adminService.GetServerStats(ctx.Context)
        if err != nil {
            return err
        }
        
        return ctx.HtmxPageData(
            "Admin Dashboard",
            "Server administration and monitoring",
            map[string]any{
                "system_status": "Healthy",
                "uptime":        stats.Uptime,
                "memory_usage":  stats.MemoryUsage,
                "cpu_usage":     stats.CPUUsage,
                "active_users":  stats.ActiveUsers,
            },
        )
    })
    
    // User management data
    adminData.GET("/users", func(ctx *lokstra.Context) error {
        users, err := userService.GetAllUsers(ctx.Context)
        if err != nil {
            return err
        }
        
        return ctx.HtmxPageData(
            "User Management",
            "Manage application users",
            map[string]any{
                "users": users,
                "total": len(users),
            },
        )
    })
    
    // Server statistics (for HTMX partial updates)
    adminData.GET("/server-stats", func(ctx *lokstra.Context) error {
        stats, err := adminService.GetServerStats(ctx.Context)
        if err != nil {
            return err
        }
        
        // Return partial data for HTMX update
        return ctx.Ok(map[string]any{
            "uptime":        stats.Uptime,
            "memory_usage":  stats.MemoryUsage,
            "cpu_usage":     stats.CPUUsage,
            "request_count": stats.RequestCount,
            "last_updated":  time.Now().Format("15:04:05"),
        })
    })
}
```

## Script Injection

### Default Script Injection

Automatically inject common JavaScript functionality:

```go
func setupHTMXWithScripts(app *lokstra.App) {
    // Create script injection with default scripts
    scriptInjection := static_files.NewScriptInjection().
        AddNamedScriptInjection("default")
    
    // Mount with script injection
    sf := static_files.New().WithSourceDir("./htmx_app")
    app.MountHtmx("/", scriptInjection, sf.Sources...)
}
```

The default script injection includes:
- **Layout switching** - Automatic full page reload when layouts change
- **HTMX headers** - Automatic `LS-Layout` header injection
- **Error handling** - Basic HTMX error response handling

### Custom Script Injection

Add custom JavaScript functionality:

```go
func setupCustomScripts(app *lokstra.App) {
    scriptInjection := static_files.NewScriptInjection()
    
    // Add custom head scripts
    scriptInjection.AddHeadScript(`
        <!-- Analytics -->
        <script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
        <script>
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', 'GA_MEASUREMENT_ID');
        </script>
    `)
    
    // Add custom body-end scripts
    scriptInjection.AddBodyEndScript(`
        <script>
            // Custom HTMX event handlers
            document.body.addEventListener('htmx:beforeRequest', function(evt) {
                console.log('HTMX request starting:', evt.detail.pathInfo.requestPath);
                
                // Show loading indicator
                document.querySelector('.htmx-indicator').style.display = 'block';
            });
            
            document.body.addEventListener('htmx:afterRequest', function(evt) {
                console.log('HTMX request completed:', evt.detail.pathInfo.requestPath);
                
                // Hide loading indicator
                document.querySelector('.htmx-indicator').style.display = 'none';
                
                // Handle errors
                if (!evt.detail.successful) {
                    showErrorMessage('Request failed. Please try again.');
                }
            });
            
            function showErrorMessage(message) {
                const errorDiv = document.createElement('div');
                errorDiv.className = 'error-message';
                errorDiv.textContent = message;
                document.body.appendChild(errorDiv);
                
                setTimeout(() => errorDiv.remove(), 5000);
            }
        </script>
    `)
    
    sf := static_files.New().WithSourceDir("./htmx_app")
    app.MountHtmx("/", scriptInjection, sf.Sources...)
}
```

## HTMX Best Practices

### Navigation with HTMX

Create smooth navigation using HTMX attributes:

```html
<!-- Navigation links with HTMX -->
<nav>
    <ul>
        <li>
            <a href="/" 
               hx-get="/" 
               hx-target="main" 
               hx-push-url="true"
               hx-indicator=".htmx-indicator">
                Home
            </a>
        </li>
        <li>
            <a href="/products" 
               hx-get="/products" 
               hx-target="main" 
               hx-push-url="true"
               hx-indicator=".htmx-indicator">
                Products
            </a>
        </li>
    </ul>
</nav>
```

### Form Handling

Handle forms with HTMX for seamless user experience:

```html
<!-- Contact form with HTMX -->
<form hx-post="/api/contact" 
      hx-target="#form-result" 
      hx-swap="innerHTML"
      hx-indicator=".form-loading">
    <div class="form-group">
        <label for="name">Name:</label>
        <input type="text" id="name" name="name" required>
    </div>
    
    <div class="form-group">
        <label for="email">Email:</label>
        <input type="email" id="email" name="email" required>
    </div>
    
    <div class="form-group">
        <label for="message">Message:</label>
        <textarea id="message" name="message" required></textarea>
    </div>
    
    <button type="submit" class="btn primary">
        Send Message
        <div class="form-loading htmx-indicator">
            <div class="spinner"></div>
        </div>
    </button>
</form>

<div id="form-result"></div>
```

### Dynamic Content Loading

Load content dynamically with smooth transitions:

```html
<!-- Product catalog with dynamic loading -->
<div class="product-filters">
    <button class="btn" 
            hx-get="/api/products?category=electronics" 
            hx-target="#product-list"
            hx-swap="innerHTML transition:true">
        Electronics
    </button>
    <button class="btn" 
            hx-get="/api/products?category=clothing" 
            hx-target="#product-list"
            hx-swap="innerHTML transition:true">
        Clothing
    </button>
</div>

<div id="product-list" class="product-grid">
    <!-- Products will be loaded here -->
</div>

<!-- Infinite scroll -->
<div hx-get="/api/products?page=2" 
     hx-trigger="revealed" 
     hx-target="#product-list"
     hx-swap="beforeend">
    <div class="loading-more">Loading more products...</div>
</div>
```

### Real-time Updates

Implement live updates using HTMX polling:

```html
<!-- Live dashboard with auto-refresh -->
<div class="dashboard-stats" 
     hx-get="/api/dashboard/stats" 
     hx-trigger="every 30s"
     hx-swap="innerHTML">
    <div class="stat-card">
        <h3>Active Users</h3>
        <div class="stat-value">{{.active_users}}</div>
    </div>
    <div class="stat-card">
        <h3>Server Load</h3>
        <div class="stat-value">{{.server_load}}%</div>
    </div>
    <div class="stat-card">
        <h3>Memory Usage</h3>
        <div class="stat-value">{{.memory_usage}}%</div>
    </div>
</div>
```

## File System Override

### Multi-layer File System

Lokstra supports multiple file sources with priority:

```go
func setupMultiLayerHTMX(app *lokstra.App) {
    sf := static_files.New().
        WithSourceDir("./custom_overrides").     // Highest priority
        WithSourceDir("./theme_files").          // Medium priority
        WithEmbedFS(defaultThemeFS, "theme").    // Lowest priority (fallback)
    
    app.MountHtmx("/", nil, sf.Sources...)
}
```

Priority order (highest to lowest):
1. **Local directories** (first added)
2. **Additional directories** (in order added)
3. **Embedded filesystems** (fallback)

### Environment-Specific Overrides

Use different file sources for different environments:

```go
func setupEnvironmentHTMX(app *lokstra.App, env string) {
    sf := static_files.New()
    
    // Add environment-specific overrides
    if env == "development" {
        sf.WithSourceDir("./dev_overrides")
    } else if env == "staging" {
        sf.WithSourceDir("./staging_overrides")
    }
    
    // Common files
    sf.WithSourceDir("./htmx_common")
    
    // Embedded fallback
    sf.WithEmbedFS(htmxFS, "htmx_app")
    
    app.MountHtmx("/", nil, sf.Sources...)
}
```

## Advanced HTMX Features

### WebSocket Integration

Combine HTMX with WebSockets for real-time features:

```html
<!-- Real-time chat with WebSocket -->
<div id="chat-messages" 
     hx-ext="ws" 
     ws-connect="/ws/chat">
    <div class="message-list">
        <!-- Messages appear here -->
    </div>
</div>

<form ws-send>
    <input type="text" name="message" placeholder="Type a message...">
    <button type="submit">Send</button>
</form>
```

### Server-Sent Events

Use SSE for live updates:

```html
<!-- Live notifications -->
<div hx-ext="sse" 
     sse-connect="/events/notifications" 
     sse-swap="message">
    <div id="notifications">
        <!-- Notifications appear here -->
    </div>
</div>
```

### Progressive Enhancement

Build pages that work without JavaScript:

```html
<!-- Form that works with and without HTMX -->
<form action="/contact" 
      method="POST"
      hx-post="/api/contact" 
      hx-target="#result"
      hx-swap="innerHTML">
    <!-- Form fields -->
    <button type="submit">Send Message</button>
</form>

<div id="result">
    <!-- HTMX result appears here, or page redirects on normal submit -->
</div>
```

## Configuration Integration

### YAML Configuration for HTMX

Configure HTMX mounting via YAML:

```yaml
# config.yaml
apps:
  - name: web-app
    address: ":8080"
    mount_htmx:
      - prefix: "/"
        sources:
          - ./htmx_app
          - embedded://htmx_default
        script_injection: default
    mount_static:
      - prefix: /static/
        sources:
          - ./htmx_app/static
          - embedded://static_default
```

### Dynamic Configuration

Configure HTMX behavior dynamically:

```go
func configureHTMX(app *lokstra.App, config map[string]any) {
    // Parse configuration
    htmxConfig := config["htmx"].(map[string]any)
    
    sources := []fs.FS{}
    for _, source := range htmxConfig["sources"].([]string) {
        if strings.HasPrefix(source, "embedded://") {
            // Add embedded FS
            sources = append(sources, embeddedFS)
        } else {
            // Add directory FS
            sources = append(sources, os.DirFS(source))
        }
    }
    
    // Setup script injection
    var scriptInjection *static_files.ScriptInjection
    if inject := htmxConfig["script_injection"]; inject != nil {
        scriptInjection = static_files.NewScriptInjection().
            AddNamedScriptInjection(inject.(string))
    }
    
    app.MountHtmx("/", scriptInjection, sources...)
}
```

## Testing HTMX Applications

### Page Data Testing

```go
func TestPageData(t *testing.T) {
    regCtx := lokstra.NewRegistrationContext()
    app := lokstra.NewApp(regCtx, "test-app", ":0")
    
    // Setup page data
    app.GET("/page-data/test", func(ctx *lokstra.Context) error {
        return ctx.HtmxPageData(
            "Test Page",
            "Test description",
            map[string]any{"test": "data"},
        )
    })
    
    // Test the endpoint
    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/page-data/test", nil)
    
    app.ServeHTTP(w, r)
    
    assert.Equal(t, 200, w.Code)
    
    var response map[string]any
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Equal(t, "Test Page", response["data"].(map[string]any)["title"])
}
```

### HTMX Request Testing

```go
func TestHTMXPartialRender(t *testing.T) {
    // Test HTMX partial rendering
    w := httptest.NewRecorder()
    r := httptest.NewRequest("GET", "/", nil)
    r.Header.Set("HX-Request", "true")
    r.Header.Set("LS-Layout", "base.html")
    
    app.ServeHTTP(w, r)
    
    assert.Equal(t, 200, w.Code)
    assert.Equal(t, "true", w.Header().Get("HX-Partial"))
    assert.Contains(t, w.Body.String(), "page content")
}
```

## Troubleshooting

### Common Issues

1. **Layout not found**
   ```
   layout not found: layouts/custom.html
   ```
   - Ensure layout file exists in file sources
   - Check file path and naming
   - Verify file source priorities

2. **Page data not loading**
   ```
   404 Not Found for /page-data/mypage
   ```
   - Ensure page data handler is registered
   - Check route path matches page request
   - Verify handler registration order

3. **Partial rendering not working**
   - Check `HX-Request` and `LS-Layout` headers
   - Verify HTMX script is loaded
   - Ensure proper HTMX attributes on elements

### Debug HTMX Requests

```html
<!-- Debug HTMX in browser -->
<script>
    // Log all HTMX events
    document.body.addEventListener('htmx:beforeRequest', function(evt) {
        console.log('HTMX Before Request:', evt.detail);
    });
    
    document.body.addEventListener('htmx:afterRequest', function(evt) {
        console.log('HTMX After Request:', evt.detail);
    });
    
    document.body.addEventListener('htmx:responseError', function(evt) {
        console.error('HTMX Error:', evt.detail);
    });
</script>
```

## Next Steps


- [Routing](./routing.md) - Learn advanced routing for HTMX apps
- [Middleware](./middleware.md) - Add authentication and security
- [Services](./services.md) - Integrate with backend services
- [Advanced Features](./advanced-features.md) - Performance optimization and deployment

## Demo Project

You can find a demo HTMX application built with Lokstra at:

- [lokstra_demo_htmxApp (GitHub)](https://github.com/primadi/lokstra_demo_htmxApp)

---

*HTMX integration in Lokstra enables you to build modern, dynamic web applications with the simplicity of traditional server-rendered pages and the interactivity of single-page applications.*