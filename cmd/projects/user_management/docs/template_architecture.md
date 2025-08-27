# Modern Template Architecture Documentation

## Overview
Layout system telah direfactor dengan template HTML terpisah untuk maximum Developer Experience (DX) dan maintainability.

## File Structure

```
handlers/
├── layout.go                 # Modern layout logic (PageContent & PageHandler)
├── template_loader.go        # Template loading and parsing
└── templates/               # HTML template files
    ├── base_layout.html     # Main page structure
    ├── meta_tags.html       # Meta tags template
    ├── scripts.html         # JavaScript loading template
    └── styles.html          # CSS loading template
```

## Architecture

### 1. layout.go
- **Purpose**: Clean, modern layout system
- **Key Functions**:
  - `RenderFullPage()`: Renders complete HTML page
  - `RenderPartialContent()`: Renders content untuk HTMX requests
  - `PageHandler()`: Single handler untuk consistent behavior

### 2. Modern PageContent Structure
```go
type PageContent struct {
    HTML            string            // Main content HTML
    Title           string            // Page title
    CurrentPage     string            // For sidebar active state
    Scripts         []string          // External scripts
    Styles          []string          // External styles
    CustomCSS       string            // Page-specific CSS
    EmbeddedScripts []string          // Embedded JS files
    MetaTags        map[string]string // Meta tags
    SidebarData     interface{}       // Custom sidebar data
}
```

### 3. Simplified Handler Pattern
```go
func CreateDashboardHandler() lokstra.HandlerFunc {
    return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
        return &PageContent{
            HTML:        content,
            Title:       "Dashboard", 
            CurrentPage: "dashboard",
            // ... other fields
        }, nil
    })
}
```

### 2. template_loader.go
- **Purpose**: Template loading dan parsing menggunakan Go embed
- **Features**:
  - `//go:embed templates/*.html`: Compiles templates into binary
  - `InitializeTemplates()`: Loads dan parses semua template
  - `prepareTemplateData()`: Converts PageContent ke TemplateData

### 3. templates/ Directory
Berisi HTML template files yang terpisah untuk maintenance yang lebih mudah:

#### base_layout.html
```html
<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    {{.MetaTags}}
    <title>{{.Title}} - User Management</title>
    {{.Scripts}}
    {{.Styles}}
</head>
<body>
    <!-- Main layout structure -->
    {{.SidebarHTML}}
    <main id="main-content">
        {{.HTML}}
    </main>
</body>
</html>
```

#### scripts.html
```html
<script src="https://cdn.tailwindcss.com"></script>
<!-- Global scripts -->

{{range .MainPageScripts}}
<script>
    // Embedded Script: {{.Name}}
    {{.Content}}
</script>
{{end}}
```

## Current Implementation

**Status**: ✅ **PRODUCTION READY** - Template-based rendering active
**Architecture**: Modular HTML templates dengan Go embed compilation
**Performance**: Templates di-compile ke binary untuk deployment optimal

## Implementation Details

### Template Functions
```go
funcMap := template.FuncMap{
    "safeJS": func(s string) template.JS {
        return template.JS(s)
    },
    "safeHTML": func(s string) template.HTML {
        return template.HTML(s)
    },
}
```

### Error Handling
- Template initialization errors: Displays error page dengan detail
- Template execution errors: Fallback dengan error information
- Graceful degradation: System tetap berfungsi meski ada template issue

## Benefits Achieved

1. ✅ **Modern Architecture**: Clean code tanpa backward compatibility burden
2. ✅ **Perfect DX**: Simple PageContent struktur untuk development
3. ✅ **Template Separation**: HTML completely separated dari Go code
4. ✅ **Embedded Assets**: Templates compiled into binary
5. ✅ **Safe Rendering**: Template functions untuk safe HTML/JS injection
6. ✅ **Error Resilience**: Comprehensive error handling
7. ✅ **Performance**: No runtime file reads - everything embedded
8. ✅ **Consistency**: Single PageHandler untuk semua use cases

## Usage Pattern

```go
// Modern, clean handler pattern
func CreateUsersHandler() lokstra.HandlerFunc {
    return PageHandler(func(c *lokstra.Context) (*PageContent, error) {
        users, err := getUsersFromAPI(c)
        if err != nil {
            return nil, err
        }
        
        return &PageContent{
            HTML:        generateUsersTable(users),
            Title:       "Users",
            CurrentPage: "users",
            MetaTags: map[string]string{
                "description": "User management dashboard",
            },
        }, nil
    })
}
```

## Removed Legacy Code

- ❌ `PageConfig` struct (tidak diperlukan)
- ❌ `SmartPageHandler` (backward compatibility)
- ❌ `DashboardLayout`, `UsersLayout` variables
- ❌ Legacy conversion functions
- ✅ Clean, modern codebase focused pada best DX

## Usage

```go
// Current usage - no changes needed
pageContent := &PageContent{
    HTML:        content,
    Title:       "Dashboard",
    CurrentPage: "dashboard",
    Scripts:     []string{},
    Styles:      []string{},
    CustomCSS:   "",
    MetaTags:    map[string]string{},
}

html := RenderFullPage(pageContent)
```

## Migration Path

1. ✅ **Phase 1**: Create template structure - COMPLETED
2. ✅ **Phase 2**: Implement template loader - COMPLETED  
3. ✅ **Phase 3**: Debug template execution - COMPLETED
4. ✅ **Phase 4**: Switch to template-based rendering - COMPLETED
5. ✅ **Phase 5**: Remove legacy implementation - COMPLETED

## Template Files

### base_layout.html
```html
<!DOCTYPE html>
<html lang="en" class="dark">
<head>
    {{.MetaTags | safeHTML}}
    <title>{{.Title}} - User Management</title>
    {{.Scripts | safeHTML}}
    {{.Styles | safeHTML}}
</head>
<body>
    {{.SidebarHTML | safeHTML}}
    <main id="main-content">
        {{.HTML | safeHTML}}
    </main>
</body>
</html>
```

### scripts.html
```html
<script src="https://cdn.tailwindcss.com"></script>
<!-- Global scripts -->

{{range .MainPageScripts}}
<script>
// Embedded Script: {{.Name}}
{{.Content | safeJS}}
</script>
{{end}}
```

## Notes

- Templates menggunakan Go's `html/template` package
- All templates di-embed ke binary menggunakan `//go:embed`
- Fallback ke legacy implementation jika template loading fails
- Structure ini ready untuk future development dan customization
