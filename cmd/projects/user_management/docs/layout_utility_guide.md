# Lokstra Layout Utility System

Sistem layout utility untuk Lokstra framework yang memungkinkan developer fokus pada konten halaman tanpa perlu menulis ulang struktur layout yang sama berulang kali.

## Konsep

Sebelumnya setiap handler harus menulis complete HTML dengan struktur:
- `<!DOCTYPE html>`
- `<head>` dengan meta tags, scripts, styles
- `<body>` dengan sidebar, header, main content
- Logic untuk detect HTMX vs direct access

Sekarang dengan layout utility, developer cukup:
1. **Tulis content saja** dalam bentuk HTML fragment
2. **Configure layout** (title, scripts, styles)
3. **Gunakan SmartPageHandler** untuk auto-handle full page vs HTMX partial

## File Structure

```
handlers/
├── layout.go          # Layout utility system
├── demo_handlers.go   # Contoh penggunaan
├── ui_handlers.go     # Handler existing (sudah direfactor)
└── ...
```

## Core Components

### 1. LayoutConfig

```go
type LayoutConfig struct {
    Title       string            // Page title
    CurrentPage string            // For sidebar navigation (dashboard/users/roles/settings)
    MetaTags    map[string]string // Custom meta tags
    Scripts     []string          // External script URLs
    Styles      []string          // External stylesheet URLs
    CustomCSS   string            // Inline custom CSS
}
```

### 2. SmartPageHandler

Handler yang intelligent - auto-detect request type:
- **Direct access** → Return full HTML page dengan layout
- **HTMX request** → Return content saja untuk partial update

```go
func SmartPageHandler(
    contentFunc func(*lokstra.Context) (string, error), 
    config LayoutConfig
) lokstra.HandlerFunc
```

### 3. SimpleContentHandler

Untuk endpoints yang hanya perlu return content (HTMX only):

```go
func SimpleContentHandler(
    contentFunc func(*lokstra.Context) (string, error)
) lokstra.HandlerFunc
```

## Cara Penggunaan

### Basic Usage (Simple Page)

```go
func CreateMyPageHandler() lokstra.HandlerFunc {
    config := DashboardLayout  // Predefined layout
    config.Title = "My Page"
    
    return SmartPageHandler(func(c *lokstra.Context) (string, error) {
        content := `
            <div class="bg-gray-800 rounded-lg p-6">
                <h2 class="text-2xl font-bold text-gray-100">Hello World</h2>
                <p class="text-gray-300">This is my content!</p>
            </div>
        `
        return content, nil
    }, config)
}
```

### Advanced Usage (Custom Assets)

```go
func CreateAdvancedPageHandler() lokstra.HandlerFunc {
    config := LayoutConfig{
        Title:       "Advanced Page",
        CurrentPage: "dashboard",
        MetaTags: map[string]string{
            "description": "My advanced page",
            "keywords":    "demo, advanced",
        },
        Scripts: []string{
            "https://cdn.jsdelivr.net/npm/chart.js",
        },
        Styles: []string{
            "https://cdnjs.cloudflare.com/ajax/libs/animate.css/4.1.1/animate.min.css",
        },
        CustomCSS: `
            .my-custom-class {
                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            }
        `,
    }
    
    return SmartPageHandler(func(c *lokstra.Context) (string, error) {
        content := `
            <div class="my-custom-class p-6 rounded-lg">
                <h2 class="text-white text-2xl">Advanced Page</h2>
                <canvas id="myChart"></canvas>
            </div>
            
            <script>
                // Chart.js code here
                new Chart(document.getElementById('myChart'), {
                    // chart config
                });
            </script>
        `
        return content, nil
    }, config)
}
```

### Content-Only Handler (HTMX Endpoints)

```go
func CreateAPIContentHandler() lokstra.HandlerFunc {
    return SimpleContentHandler(func(c *lokstra.Context) (string, error) {
        content := `
            <div class="alert bg-green-600 text-white p-4 rounded">
                API response content here
            </div>
        `
        return content, nil
    })
}
```

## Predefined Layouts

Layout umum yang sudah tersedia:

```go
var (
    DashboardLayout = LayoutConfig{...}  // For dashboard pages
    UsersLayout     = LayoutConfig{...}  // For user management
    RolesLayout     = LayoutConfig{...}  // For role management  
    SettingsLayout  = LayoutConfig{...}  // For settings pages
)
```

## Migration dari Handler Lama

### Sebelum (Manual Layout):

```go
func CreateOldHandler() lokstra.HandlerFunc {
    return func(c *lokstra.Context) error {
        isHTMXRequest := c.GetHeader("HX-Request") == "true"
        
        content := `<div>My content</div>`
        
        if isHTMXRequest {
            return c.HTML(200, content)
        }
        
        // Manually build full HTML with DOCTYPE, head, body, sidebar, etc.
        fullHTML := fmt.Sprintf(`<!DOCTYPE html>...`, content)
        return c.HTML(200, fullHTML)
    }
}
```

### Sesudah (Layout Utility):

```go
func CreateNewHandler() lokstra.HandlerFunc {
    config := DashboardLayout
    config.Title = "My Page"
    
    return SmartPageHandler(func(c *lokstra.Context) (string, error) {
        content := `<div>My content</div>`
        return content, nil
    }, config)
}
```

## Benefits

1. **DRY Principle**: Tidak ada duplikasi layout code
2. **Focus on Content**: Developer fokus pada business logic dan content
3. **Consistent Layout**: Semua page punya struktur layout yang sama
4. **HTMX Ready**: Auto-handle full page vs partial requests
5. **Configurable**: Easy customize per page (title, scripts, styles)
6. **Maintainable**: Update layout di 1 tempat affect semua pages
7. **Type Safe**: Strong typing dengan Go structs

## Best Practices

1. **Use Predefined Layouts** untuk consistency
2. **Keep Content Functions Simple** - pure content generation
3. **Handle Errors** dalam content functions dengan proper error returns  
4. **Minimize Custom Scripts** - prefer HTMX and Alpine.js
5. **Use Semantic HTML** dalam content untuk accessibility
6. **Test Both Modes** - direct access dan HTMX requests

## Future Enhancements

Potential improvements untuk layout system:
- Template inheritance system
- Layout caching mechanism
- Component system untuk reusable UI elements
- Theme switching support
- SEO optimization helpers
- Error page layouts
