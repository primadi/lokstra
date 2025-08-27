# Demo: Unified PageContent System
## Konsistensi Behavior antara Full Page Load dan HTMX Partial Load

### Problem yang Diselesaikan

**Sebelum (SmartPageHandler + PageConfig):**
- Full page load: PageConfig digunakan (title, scripts, CSS bekerja)
- HTMX partial load: PageConfig diabaikan (hanya content, assets hilang)
- **Hasil**: Behavior tidak konsisten!

**Sesudah (UnifiedPageHandler + PageContent):**
- Full page load: PageContent lengkap (HTML + assets)
- HTMX partial load: PageContent yang sama (HTML + assets tetap ada)
- **Hasil**: Behavior 100% konsisten!

### Struktur PageContent

```go
type PageContent struct {
    HTML        string            // Konten HTML
    Title       string            // Page title
    CurrentPage string            // Untuk sidebar active state
    Scripts     []string          // External scripts (konsisten)
    Styles      []string          // External stylesheets (konsisten)
    CustomCSS   string            // Inline CSS (konsisten)
    MetaTags    map[string]string // Meta tags (konsisten)
}
```

### Contoh Implementasi Dashboard

```go
func CreateDashboardHandler() lokstra.HandlerFunc {
    return UnifiedPageHandler(func(c *lokstra.Context) (*PageContent, error) {
        content := `<div>Dashboard content...</div>`
        
        return &PageContent{
            HTML:        content,
            Title:       "Dashboard",
            CurrentPage: "dashboard",
        }, nil
    })
}
```

### Contoh dengan Page-Specific Assets

```go
func CreateUsersHandler() lokstra.HandlerFunc {
    return UnifiedPageHandler(func(c *lokstra.Context) (*PageContent, error) {
        content := generateUserTable()
        
        return &PageContent{
            HTML:        content,
            Title:       "User Management",
            CurrentPage: "users",
            
            // CSS ini akan KONSISTEN muncul di full page DAN HTMX
            CustomCSS: `
                .user-row:hover {
                    background-color: rgba(75, 85, 99, 0.5);
                    transform: translateY(-1px);
                    transition: all 0.2s ease;
                }
                
                .action-button:hover {
                    transform: scale(1.1);
                }
            `,
        }, nil
    })
}
```

### Advanced Example dengan External Scripts

```go
func CreateAdvancedUserFormHandler() lokstra.HandlerFunc {
    return UnifiedPageHandler(func(c *lokstra.Context) (*PageContent, error) {
        return &PageContent{
            HTML:        formHTML,
            Title:       "Advanced User Form",
            CurrentPage: "user_form",
            
            // External scripts - KONSISTEN di full page DAN HTMX
            Scripts: []string{
                "https://cdn.jsdelivr.net/npm/validator@13.7.0/validator.min.js",
            },
            
            // Custom CSS - KONSISTEN di full page DAN HTMX
            CustomCSS: `
                .user-form-container {
                    background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
                    border-radius: 12px;
                    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.3);
                }
                
                .validation-error {
                    color: #ef4444;
                    font-size: 0.875rem;
                }
            `,
            
            // Meta tags - KONSISTEN di full page DAN HTMX
            MetaTags: map[string]string{
                "description": "User management form with advanced validation",
                "keywords":    "user, form, validation, management",
            },
        }, nil
    })
}
```

### Rendering System Comparison

#### OLD SYSTEM (Inconsistent):
```go
// Full page load
RenderFullPage(content, config) // config digunakan

// HTMX partial load
return content // config DIABAIKAN!
```

#### NEW SYSTEM (Consistent):
```go
// Full page load
RenderFullPage(pageContent) // semua data dari pageContent

// HTMX partial load  
RenderPartialContent(pageContent) // data SAMA dari pageContent
```

### Auto-Navigation System Integration

Dengan `.nav-page` class, HTMX attributes otomatis ditambahkan:

```html
<!-- Template -->
<button class="nav-page" data-target="/users">Users</button>

<!-- Hasil automatic injection -->
<button class="nav-page" data-target="/users" 
        hx-get="/users" 
        hx-target="#main-content"
        hx-indicator="#loading-indicator">Users</button>
```

### Testing Scenarios

1. **Full Page Load Test:**
   - Navigate langsung ke `/users`
   - CSS hover effects harus bekerja
   - Scripts harus loaded
   - Title harus berubah

2. **HTMX Partial Load Test:**
   - Click sidebar navigation ke users  
   - CSS hover effects harus SAMA bekerja
   - Scripts harus SAMA loaded
   - Title harus SAMA berubah

3. **Page-Specific Assets Test:**
   - Load form page dengan external validator.js
   - Validasi harus bekerja di full page load
   - Validasi harus SAMA bekerja di HTMX load

### Benefits

1. **Developer Experience:**
   - Satu struktur untuk semua rendering
   - Tidak perlu khawatir tentang inkonsistensi
   - Page-specific assets guaranteed bekerja

2. **User Experience:**
   - Behavior identik di semua scenario
   - No missing assets di HTMX navigation
   - Consistent visual feedback

3. **Maintainability:**
   - Unified system, less confusion
   - Easier debugging
   - Clear separation of concerns

### Migration Guide

**Convert dari SmartPageHandler:**

```go
// OLD
return SmartPageHandler(func(c *lokstra.Context) (string, error) {
    content := generateContent()
    return content, nil
}, config)

// NEW
return UnifiedPageHandler(func(c *lokstra.Context) (*PageContent, error) {
    content := generateContent()
    return &PageContent{
        HTML:        content,
        Title:       "Page Title",
        CurrentPage: "page_name",
        // Add page-specific assets as needed
    }, nil
})
```

### Result

🎯 **Unified Behavior**: Page-specific assets bekerja identik di full page load dan HTMX partial load
🎯 **Simplified Development**: Satu system untuk semua rendering scenarios  
🎯 **Enhanced UX**: Consistent behavior eliminates user confusion
🎯 **Future-Proof**: Extensible structure untuk enhancement berikutnya
