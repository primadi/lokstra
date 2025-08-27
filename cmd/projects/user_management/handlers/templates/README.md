# Template Structure Documentation

## Overview

Template HTML untuk User Management Application telah dipisahkan dari kode Go untuk memudahkan maintenance dan meningkatkan readability.

## File Structure

```
handlers/
├── templates/
│   ├── base_layout.html      # Main layout template
│   ├── meta_tags.html        # Meta tags template
│   ├── scripts.html          # JavaScript loading template
│   └── styles.html           # CSS/Styles loading template
├── layout.go                 # Main layout handler (refactored)
└── template_loader.go        # Template loading and rendering logic
```

## Template Files

### 1. base_layout.html
- Main HTML structure untuk full page layout
- Menggunakan Go template syntax: `{{.FieldName}}`
- Contains: HTML skeleton, header, sidebar, loading indicator, main content area

### 2. meta_tags.html
- Template untuk meta tags
- Includes: charset, viewport, dan page-specific meta tags
- Supports dynamic meta tag injection via `{{.PageMetaTags}}`

### 3. scripts.html
- Template untuk JavaScript loading
- Includes: External CDN scripts, embedded scripts, page-specific scripts
- Supports main page scripts dan embedded scripts dari assets

### 4. styles.html
- Template untuk CSS loading
- Includes: Main layout CSS, external stylesheets, custom CSS
- Supports embedded CSS dari assets dan page-specific styles

## Key Features

### Template Loading
- Templates di-compile ke binary menggunakan `//go:embed`
- Automatic initialization pada first use
- Fallback ke legacy rendering jika template loading gagal

### Data Structure
```go
type TemplateData struct {
    HTML            string                // Main content HTML
    Title           string                // Page title
    SidebarHTML     string                // Rendered sidebar HTML
    MetaTags        string                // Rendered meta tags
    Scripts         string                // Rendered scripts
    Styles          string                // Rendered styles
    PageMetaTags    map[string]string     // Page-specific meta tags
    MainPageScripts []ScriptData          // Main page navigation scripts
    ExternalScripts []string              // External script URLs
    EmbeddedScripts []ScriptData          // Embedded scripts
    MainLayoutCSS   string                // Main layout CSS content
    ExternalStyles  []string              // External style URLs
    CustomCSS       string                // Page-specific custom CSS
}
```

### Backward Compatibility
- Legacy `renderFullPageLegacy()` function sebagai fallback
- Existing `PageContent` struct tetap unchanged
- Existing handlers tidak perlu dimodifikasi

## Usage

### Template-based Rendering (New)
```go
// Template akan di-load secara otomatis
html := RenderFullPage(pageContent)
```

### Fallback Rendering (Legacy)
```go
// Fallback akan digunakan jika template loading gagal
html := renderFullPageLegacy(pageContent)
```

## Benefits

1. **Separation of Concerns**: HTML template terpisah dari Go code
2. **Maintainability**: Mudah edit HTML tanpa modify Go code
3. **Readability**: Template HTML lebih mudah dibaca dan dimengerti
4. **Binary Compilation**: Templates di-compile ke binary (no external files needed)
5. **Error Resilience**: Automatic fallback jika template system gagal
6. **Backward Compatibility**: Existing code tetap berfungsi tanpa perubahan

## Template Syntax

### Basic Variables
```html
{{.Title}}          <!-- Page title -->
{{.HTML}}           <!-- Main content -->
{{.SidebarHTML}}    <!-- Sidebar content -->
```

### Conditional Rendering
```html
{{if .CustomCSS}}
<style>
    {{.CustomCSS}}
</style>
{{end}}
```

### Loops
```html
{{range .ExternalScripts}}
<script src="{{.}}"></script>
{{end}}
```

### Nested Data
```html
{{range $name, $content := .PageMetaTags}}
<meta name="{{$name}}" content="{{$content}}">
{{end}}
```

## Future Enhancements

1. **Template Caching**: Add template caching untuk performance
2. **Template Inheritance**: Support untuk template inheritance/extension
3. **Component System**: Reusable template components
4. **Hot Reload**: Development mode dengan template hot reload
5. **Template Validation**: Compile-time validation untuk template syntax
