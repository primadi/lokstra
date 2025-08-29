# Static Files with Fallback Example

This example demonstrates how to serve static files with fallback priority in Lokstra.

## How it works

The `MountStaticWithFallback` method allows you to serve static files from multiple sources with fallback priority:

```go
app.MountStaticWithFallback("/static", 
    http.Dir("./project/assets"),     // Highest priority
    http.Dir("./project/static"),     // Medium priority  
    frameworkAssets,                  // Lowest priority (embed.FS)
)
```

## Fallback Priority

1. **Project Assets** (`./project/assets/`) - Highest priority
   - Project-specific overrides
   - Custom styling, branding, etc.

2. **Project Static** (`./project/static/`) - Medium priority
   - Main project assets
   - Default project files

3. **Framework Assets** (`frameworkAssets` embed.FS) - Lowest priority
   - Framework default assets
   - Embedded in binary as fallback

## Testing the Fallback

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Test different scenarios:**
   
   - **Access `/static/app.js`** → Serves `./project/assets/app.js` (project override)
   - **Access `/static/default.js`** → Serves `./project/static/default.js` (project default)
   - **Access `/static/style.css`** → Serves from `frameworkAssets` embed.FS (framework fallback)
   - **Access `/static/nonexistent.js`** → Returns 404 (not found in any source)

3. **Experiment with overrides:**
   - Delete `./project/assets/app.js` → `/static/app.js` will fallback to `./project/static/app.js` (if exists)
   - Delete both → `/static/app.js` will fallback to `frameworkAssets/app.js`
   - Delete all → `/static/app.js` will return 404

## Use Cases

- **Multi-tenant applications:** Each tenant can override specific assets
- **Plugin systems:** Plugins can override framework default assets
- **Theme systems:** Themes can override default styling
- **Development vs Production:** Different asset sources based on environment

## API Endpoints

- `GET /` - Server information
- `GET /api/info` - Detailed mount information
- `GET /health` - Health check
- `GET /static/*` - Static files with fallback
- `GET /assets/*` - Alternative static files with fallback

## Directory Structure

```
06_serve_static_with_fallback/
├── main.go
├── README.md
├── framework_assets/          # Embedded framework assets
│   ├── app.js
│   ├── style.css
│   └── index.html
└── project/
    ├── assets/               # Project overrides (highest priority)
    │   └── app.js
    └── static/               # Project defaults (medium priority)
        └── default.js
```
