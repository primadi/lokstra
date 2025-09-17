# Mount Static Files Example

This example demonstrates how to serve static files using Lokstra's `MountStatic` functionality.

## Features Demonstrated

1. **Multiple Static Directories** - Serving files from different directories
2. **Custom URL Prefixes** - Mounting directories at different URL paths
3. **Mixed Content Types** - HTML, CSS, text files
4. **API Coexistence** - Static files alongside API endpoints

## Directory Structure

```
03_mount_static/
├── main.go              # Main application
├── README.md            # This file
├── static/              # Static files directory
│   └── sample.txt       # Sample text file
├── assets/              # Assets directory (CSS, JS, images)
│   └── style.css        # Sample CSS file
└── public/              # Public HTML files
    └── index.html       # Sample HTML page
```

## URL Mapping

- `/static/*` → `./static/` directory
- `/assets/*` → `./assets/` directory  
- `/public/*` → `./public/` directory

## How to Run

```bash
go run main.go
```

The server will start on port 8080.

## Test Endpoints

### Static Files
```bash
# Text file from static directory
curl http://localhost:8080/static/sample.txt

# CSS file from assets directory
curl http://localhost:8080/assets/style.css

# HTML file from public directory
curl http://localhost:8080/public/index.html
# Or visit in browser: http://localhost:8080/public/index.html
```

### API Endpoints
```bash
# Root endpoint - shows available routes
curl http://localhost:8080/

# Server information
curl http://localhost:8080/api/info

# Health check
curl http://localhost:8080/health
```

## Key Implementation Points

1. **MountStatic Usage**:
   ```go
   app.MountStatic("/static", http.Dir("./static"))
   ```

2. **http.Dir Type**: Must use `http.Dir()` to wrap directory paths

3. **Route Priority**: API routes should be defined before static mounts to avoid conflicts

4. **File Serving**: Lokstra automatically handles:
   - MIME type detection
   - HTTP caching headers
   - Directory traversal protection
   - 404 errors for missing files

## Expected Behavior

- Static files are served with appropriate MIME types
- Directory listings are disabled for security
- Files not found return 404 errors
- API endpoints remain accessible alongside static files
- CSS files are properly linked in HTML pages

## Real-World Usage

In production applications, you would typically:

1. **Separate Concerns**: Use different directories for different asset types
2. **CDN Integration**: Serve static assets from CDN while keeping API on your server
3. **Cache Headers**: Configure appropriate cache headers for static content
4. **Compression**: Enable gzip compression for text-based static files
5. **Security**: Implement proper access controls for sensitive static content
