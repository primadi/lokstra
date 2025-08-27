# HTMX Navigation Implementation

## Overview
This document explains the implementation of HTMX-powered navigation that provides SPA-like experience without the complexity of a traditional Single Page Application.

## Architecture

### 1. Template Structure
- **Sidebar Template**: Located in `handlers/templates.go`
- **HTMX Attributes**: Added to navigation links for partial updates
- **Main Content Area**: Target element `#main-content` for content swapping

### 2. Navigation Links Configuration
Each navigation link includes the following HTMX attributes:
```html
<a href="/dashboard" 
   hx-get="/api/content/dashboard" 
   hx-target="#main-content" 
   hx-swap="innerHTML" 
   hx-indicator="#loading-indicator" 
   hx-push-url="/dashboard"
   class="nav-item ...">
```

**Attribute Explanation:**
- `hx-get`: Makes a GET request to the content API endpoint
- `hx-target`: Specifies which element to update (`#main-content`)
- `hx-swap`: Determines how content is updated (`innerHTML`)
- `hx-indicator`: Shows loading indicator during requests
- `hx-push-url`: Updates browser URL for proper navigation history

### 3. Content API Endpoints

#### Route Configuration (`config/user_management.yaml`)
```yaml
# HTMX Content API Routes for partial updates
- method: "GET"
  path: "/api/content/dashboard"
  handler: "ui.content.dashboard"
  
- method: "GET"
  path: "/api/content/users"
  handler: "ui.content.users"
  
- method: "GET"
  path: "/api/content/roles"
  handler: "ui.content.roles"
  
- method: "GET"
  path: "/api/content/settings"
  handler: "ui.content.settings"
```

#### Handler Implementation (`handlers/ui_handlers.go`)
Content handlers return HTML fragments instead of full pages:

```go
func CreateDashboardContentHandler() lokstra.HandlerFunc {
    return func(c *lokstra.Context) error {
        content := `
            <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
                <!-- Dashboard content here -->
            </div>
        `
        return c.HTML(200, content)
    }
}
```

### 4. Loading Indicators & Animations

#### CSS Styling
```css
/* HTMX Loading Indicator */
.htmx-indicator {
    display: none;
}
.htmx-request .htmx-indicator {
    display: block;
}

/* Smooth transitions for content updates */
#main-content {
    transition: opacity 0.2s ease-in-out;
}
.htmx-request #main-content {
    opacity: 0.7;
}

/* Navigation active state animation */
.nav-item {
    transition: all 0.2s ease-in-out;
}
.nav-item:hover {
    transform: translateX(4px);
}
```

#### Loading Indicator Component
```html
<div id="loading-indicator" class="htmx-indicator fixed top-4 right-4 z-50">
    <div class="bg-blue-600 text-white px-4 py-2 rounded-lg shadow-lg flex items-center">
        <svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white">
            <!-- Spinner SVG -->
        </svg>
        Loading...
    </div>
</div>
```

## Benefits

### 1. Performance
- **Partial Updates**: Only main content is replaced, sidebar stays static
- **Faster Loading**: No need to reload entire page structure
- **Reduced Bandwidth**: Smaller payloads compared to full page reloads

### 2. User Experience
- **Smooth Transitions**: Visual feedback during navigation
- **Loading Indicators**: Clear feedback during content updates
- **Browser History**: Proper URL updates maintain bookmarkability
- **Fallback Support**: Links work normally if JavaScript is disabled

### 3. Developer Experience
- **Simple Implementation**: No complex state management
- **Server-Side Rendering**: Familiar template-based approach
- **Progressive Enhancement**: Builds on standard HTML forms
- **Easy Debugging**: Standard HTTP requests for content

## Usage Examples

### Testing Navigation
1. Start the server: `go run main.go`
2. Open browser to: `http://localhost:8081`
3. Click navigation links to see smooth content updates
4. Check Network tab to see HTMX requests to `/api/content/*` endpoints

### Adding New Pages
1. **Create Content Handler**:
   ```go
   func CreateNewPageContentHandler() lokstra.HandlerFunc {
       return func(c *lokstra.Context) error {
           content := `<div>New page content</div>`
           return c.HTML(200, content)
       }
   }
   ```

2. **Register Handler** (`main.go`):
   ```go
   regCtx.RegisterHandler("ui.content.newpage", handlers.CreateNewPageContentHandler())
   ```

3. **Add Route** (`config/user_management.yaml`):
   ```yaml
   - method: "GET"
     path: "/api/content/newpage"
     handler: "ui.content.newpage"
   ```

4. **Update Menu** (`handlers/menu_data.go`):
   ```go
   {
       Icon: "icon-svg-path",
       Text: "New Page",
       URL:  "/newpage",
       Active: currentPage == "newpage",
   }
   ```

## Browser Compatibility
- **Modern Browsers**: Full HTMX support with all features
- **Legacy Browsers**: Graceful degradation to standard navigation
- **JavaScript Disabled**: Standard HTML links provide fallback

## Performance Considerations
- **Caching**: Content API responses can be cached
- **Bundle Size**: HTMX library is lightweight (~10kb gzipped)
- **SEO**: Content is server-rendered and crawlable
- **Progressive Loading**: Initial page load includes sidebar, subsequent loads are content-only

## Next Steps
1. **Error Handling**: Add error states for failed content loads
2. **Advanced Animations**: Implement page transition effects
3. **Form Integration**: Add HTMX form handling for data operations
4. **WebSocket Support**: Real-time updates for dynamic content
5. **Testing**: Add automated tests for HTMX interactions
