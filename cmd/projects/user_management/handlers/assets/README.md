# Assets Structure Documentation

## Overview
This document describes the organized asset structure for the User Management application, separating CSS and JavaScript into maintainable files.

## Directory Structure

```
handlers/assets/
├── css/
│   └── main-layout.css           # Main layout styles
├── js/
│   ├── main-page/               # Main application JavaScript modules
│   │   ├── navigation.js        # HTMX navigation utilities
│   │   ├── sidebar.js          # Sidebar management
│   │   ├── emergency-cleanup.js # Emergency UI cleanup
│   │   └── app.js              # Application initialization
│   ├── user-form-validation.js  # User form specific validation
│   ├── table-enhancements.js    # Table UI enhancements
│   └── navigation-enhancements.js # Legacy navigation enhancements
└── assets.go                    # Go embedded assets registry
```

## JavaScript Modules

### Core Modules (main-page/)

#### 1. navigation.js
- HTMX configuration and event handling
- Navigation protection and debouncing
- Loading indicator management
- Chrome-specific optimizations

#### 2. sidebar.js
- Sidebar active menu updates
- Path matching logic (exact and partial)
- DOM manipulation optimization
- Debug logging for development

#### 3. emergency-cleanup.js
- Background cleanup for loading indicators
- Failsafe mechanisms for UI states
- Interval-based monitoring

#### 4. app.js
- Main application entry point
- Initializes all modules
- DOMContentLoaded handler

### Legacy Modules
- `user-form-validation.js` - Form validation utilities
- `table-enhancements.js` - Table UI improvements
- `navigation-enhancements.js` - Legacy navigation code

## CSS Structure

### main-layout.css
- Loading indicator styles and overrides
- HTMX integration styles
- Chrome-specific fixes
- Navigation and transition styles
- Responsive design utilities

## Go Integration

### assets.go
- Embedded asset registry using `//go:embed`
- JavaScript and CSS content mapping
- Utility functions for accessing embedded content

### Functions Available:
```go
// JavaScript
assets.GetEmbeddedScript("main-page-navigation")
assets.GetEmbeddedScript("main-page-sidebar")
assets.GetInlineScript("script-name")

// CSS
assets.GetEmbeddedStyle("main-layout")
assets.GetInlineStyle("style-name")
```

## Loading Order

### JavaScript Loading Sequence:
1. External libraries (Tailwind, HTMX, Alpine.js)
2. Core modules in order:
   - navigation.js (utilities and HTMX config)
   - sidebar.js (sidebar management)
   - emergency-cleanup.js (background cleanup)
   - app.js (initialization)

### CSS Loading:
1. Embedded main-layout.css (via `GetEmbeddedStyle`)
2. Page-specific external stylesheets
3. Page-specific custom CSS

## Benefits

### Maintainability
- ✅ Separated concerns (navigation, sidebar, cleanup)
- ✅ Individual file editing without affecting layout.go
- ✅ Clear module boundaries and responsibilities
- ✅ Easy to add new functionality

### Performance
- ✅ Files compiled into binary (no external dependencies)
- ✅ Optimized loading order
- ✅ Reduced inline code in HTML
- ✅ Better browser caching for development

### Development Experience
- ✅ Syntax highlighting in JavaScript/CSS files
- ✅ Better IDE support and autocomplete
- ✅ Easier debugging and console logging
- ✅ Modular testing capabilities

## Usage Examples

### Adding New JavaScript Module:
1. Create file in `js/main-page/new-feature.js`
2. Add `//go:embed` directive in `assets.go`
3. Add to `EmbeddedScripts` map
4. Include in layout.go loading sequence

### Adding New CSS:
1. Create file in `css/new-styles.css`
2. Add `//go:embed` directive in `assets.go`
3. Add to `EmbeddedStyles` map
4. Use `GetEmbeddedStyle()` in layout template

### Page-Specific Assets:
```go
pageContent := &PageContent{
    HTML: content,
    Title: "Page Title",
    CurrentPage: "page-id",
    EmbeddedScripts: []string{"main-page-navigation", "custom-feature"},
    Scripts: []string{"/external/library.js"},
    CustomCSS: "body { background: red; }",
}
```

## Migration Notes

### Before (Inline):
- All JavaScript was embedded in layout.go template strings
- CSS mixed with JavaScript in single template block
- Hard to maintain and edit
- No syntax highlighting or IDE support

### After (Modular):
- Separated into logical modules with clear responsibilities
- Full IDE support with syntax highlighting
- Easy to maintain and extend
- Better code organization and reusability
