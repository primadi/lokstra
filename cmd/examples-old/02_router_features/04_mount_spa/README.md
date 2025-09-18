# Mount SPA (Single Page Application) Example

This example demonstrates how to serve Single Page Applications using Lokstra's `MountSPA` functionality, which enables client-side routing with proper fallback handling.

## Features Demonstrated

1. **SPA Mount with Fallback** - All unmatched routes serve the SPA's index.html
2. **Client-Side Routing** - JavaScript-based navigation without page reloads
3. **API Coexistence** - Backend API endpoints alongside frontend SPA routes
4. **Proper Route Precedence** - API routes take priority over SPA fallback
5. **Browser History Support** - Back/forward button handling

## How SPA Mounting Works

```go
// API routes defined first (higher precedence)
apiGroup := app.Group("/api")
apiGroup.GET("/users", handler)

// Health/info endpoints
app.GET("/health", handler) 

// SPA mount last (catches all unmatched routes)
app.MountSPA("/", "./spa/index.html")
```

## Directory Structure

```
04_mount_spa/
├── main.go              # Main application with API and SPA setup
├── README.md            # This file
└── spa/                 # SPA files directory
    └── index.html       # Main SPA file (fallback for all routes)
```

## Route Handling

### Server-Side Routes (Higher Priority)
- `/api/users` - Get all users
- `/api/users/:id` - Get specific user
- `/health` - Health check
- `/server-info` - Server information

### Client-Side Routes (SPA Fallback)
- `/` - Home page
- `/about` - About page
- `/dashboard` - Dashboard page
- `/users` - Users management page
- `/users/123` - User profile page
- `/any/other/path` - Any unmatched route

## How to Run

```bash
go run main.go
```

The server will start on port 8080.

## Test the SPA

### Visit in Browser
```
http://localhost:8080/           # SPA home page
http://localhost:8080/about      # Client-side route
http://localhost:8080/dashboard  # Client-side route
http://localhost:8080/users/123  # Client-side route with parameter
```

### Test API Endpoints
```bash
# Server-side API calls
curl http://localhost:8080/api/users
curl http://localhost:8080/api/users/1
curl -X POST http://localhost:8080/api/users
curl http://localhost:8080/health
```

## Key Features in the Example

### 1. Interactive Navigation
- Click navigation buttons to see client-side routing
- URL changes without page reload
- Browser back/forward buttons work correctly

### 2. API Integration
- Click "Fetch Users" to call `/api/users`
- Click "Fetch User 1" to call `/api/users/1`
- Click "Create User" to POST to `/api/users`

### 3. Route Information
- Current route is displayed in the UI
- Shows which routes are client-side vs server-side

## Implementation Details

### 1. Route Precedence
```go
// ✅ Correct order - API routes first
app.GET("/api/users", handler)
app.MountSPA("/", "./spa/index.html")

// ❌ Wrong order - SPA would catch API routes
app.MountSPA("/", "./spa/index.html")
app.GET("/api/users", handler) // Never reached!
```

### 2. SPA Fallback Behavior
- Any route not matched by server handlers serves `index.html`
- Client-side JavaScript handles routing logic
- Perfect for React, Vue, Angular applications

### 3. API Communication
```javascript
// Frontend can still call backend APIs
fetch('/api/users')
  .then(response => response.json())
  .then(data => console.log(data));
```

## Real-World Usage

### Production Considerations

1. **Build Process**: In production, serve built/minified SPA files
2. **Static Assets**: Use `MountStatic` for CSS/JS/images alongside `MountSPA`
3. **Cache Headers**: Configure appropriate caching for SPA and API responses
4. **Error Handling**: Implement proper error pages for 404s and API failures
5. **Security**: Ensure API endpoints have proper authentication/authorization

### Typical Structure
```go
// Static assets (with long cache headers)
app.MountStatic("/static", http.Dir("./build/static"))

// API routes
app.Group("/api", "auth", "cors").GET("/users", getUsersHandler)

// SPA fallback (last!)
app.MountSPA("/", "./build/index.html")
```

## Expected Behavior

1. **Direct URL Access**: Typing `/dashboard` in browser serves SPA index.html
2. **API Calls**: `/api/*` routes return JSON from server
3. **Client Navigation**: Clicking nav buttons updates URL without page reload
4. **Refresh Works**: Refreshing on `/about` still serves the SPA
5. **History Support**: Back/forward buttons work correctly

This setup is perfect for modern web applications that need both a rich frontend experience and a robust backend API.
