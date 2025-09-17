# Getting Started with Lokstra

Lokstra is a modern Go web framework that supports both imperative (code-based) and declarative (configuration-based) approaches for building web applications.

## Installation

```bash
go get github.com/primadi/lokstra
```

## Quick Start - Code Approach

### Single Application

```go
package main

import (
    "github.com/primadi/lokstra"
)

func main() {
    // Create global registration context
    regCtx := lokstra.NewGlobalRegistrationContext()

    // Create a new application
    app := lokstra.NewApp(regCtx, "demo-app", ":8080")
    
    // Add routes
    app.GET("/hello", func(ctx *lokstra.Context) error {
        return ctx.Ok(map[string]any{
            "message": "Hello, Lokstra!",
        })
    })
    
    // Start the application
    if err := app.Start(); err != nil {
        panic(err)
    }
}
```

### With Graceful Shutdown (Recommended)

```go
import "time"

func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    app := lokstra.NewApp(regCtx, "demo-app", ":8080")
    
    app.GET("/hello", func(ctx *lokstra.Context) error {
        return ctx.Ok("Hello, World!")
    })
    
    // Start with graceful shutdown
    if err := app.StartAndWaitForShutdown(30 * time.Second); err != nil {
        panic(err)
    }
}
```

## Multiple Applications with Server

When you need to run multiple applications, use the Server:

```go
func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    server := lokstra.NewServer(regCtx, "demo-server")

    // Create multiple apps
    app1 := lokstra.NewApp(regCtx, "api-app", ":8080")
    app1.GET("/api/users", func(ctx *lokstra.Context) error {
        return ctx.Ok("Users API")
    })

    app2 := lokstra.NewApp(regCtx, "admin-app", ":8081")
    app2.GET("/admin", func(ctx *lokstra.Context) error {
        return ctx.Ok("Admin Panel")
    })

    // Add apps to server
    server.AddApp(app1)
    server.AddApp(app2)

    // Start all apps
    if err := server.StartAndWaitForShutdown(30 * time.Second); err != nil {
        panic(err)
    }
}
```

### Port Merging

If multiple apps use the same port, Lokstra automatically merges them:

```go
app1 := lokstra.NewApp(regCtx, "app1", ":8080")
app1.GET("/api", apiHandler)

app2 := lokstra.NewApp(regCtx, "app2", ":8080") // Same port
app2.GET("/admin", adminHandler)

server.AddApp(app1).AddApp(app2)
// Both apps will run on the same :8080 listener
```

## Configuration Approach

Lokstra supports YAML-based configuration for declarative application setup.

### Basic Configuration File

Create `lokstra.yaml`:

```yaml
server:
  name: demo-server

apps:
  - name: demo-app
    addr: ":8080"
    routes:
      - method: GET
        path: /hello
        handler: hello
      - method: POST
        path: /api/users
        handler: createUser
```

### Loading Configuration

```go
func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // Register handlers referenced in config
    regCtx.RegisterHandler("hello", func(ctx *lokstra.Context) error {
        return ctx.Ok("Hello from config!")
    })
    
    regCtx.RegisterHandler("createUser", func(ctx *lokstra.Context) error {
        return ctx.Ok("User created")
    })
    
    // Load configuration
    cfg, err := lokstra.LoadConfigDir("./configs")
    if err != nil {
        panic(err)
    }
    
    // Create server from config
    server, err := lokstra.NewServerFromConfig(regCtx, cfg)
    if err != nil {
        panic(err)
    }
    
    if err := server.StartAndWaitForShutdown(30 * time.Second); err != nil {
        panic(err)
    }
}
```

### Loading Single File vs Directory

```go
// Load single file
cfg, err := config.LoadConfigFile("lokstra.yaml")

// Load entire directory (merges all .yaml files)
cfg, err := lokstra.LoadConfigDir("./configs")
```

## Router Embedded Interface

Lokstra apps have an embedded router, so these are equivalent:

```go
// Direct on app (shorthand)
app.GET("/users", handler)
app.POST("/users", handler)

// Via router field (explicit)
app.Router.GET("/users", handler)
app.Router.POST("/users", handler)
```

## Working with AI Assistants and Copilot

When working with AI assistants like GitHub Copilot, VS Code Copilot, or other AI agents, you can help them understand your Lokstra project by providing clear context.

### Option 1: Add Comments in main.go

Add descriptive comments at the top of your main function:

```go
// This project is built using Lokstra Framework
// (https://github.com/primadi/lokstra).
//
// Lokstra documentation available at /docs/README.md
// Core concepts: /docs/core-concepts.md
// Configuration: /docs/configuration.md
func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    
    // Create application using Lokstra framework
    app := lokstra.NewApp(regCtx, "my-app", ":8080")
    
    // Add routes using Lokstra's handler pattern
    app.GET("/health", func(ctx *lokstra.Context) error {
        return ctx.Ok("Service is running")
    })
    
    app.Start()
}
```

### Option 2: Create Project README.md

Add a README.md file in your project root:

```markdown
# My Lokstra Project

This project uses the [Lokstra](https://github.com/primadi/lokstra) Go web framework.

## Framework Documentation

- [Framework Overview](./docs/README.md)
- [Getting Started](./docs/getting-started.md) 
- [Core Concepts](./docs/core-concepts.md)
- [Built-in Services](./docs/built-in-services.md)
- [Built-in Middleware](./docs/built-in-middleware.md)

## Lokstra Key Features

- Registration Context for dependency injection
- Request binding with struct tags: `path:"id"`, `query:"page"`, `body:"name"`
- Response helpers: `ctx.Ok()`, `ctx.ErrorBadRequest()`, etc.
- Auto-bind smart handlers: `func handler(ctx *lokstra.Context, req *MyRequest) error`
- YAML configuration support with schema validation
- Built-in middleware: cors, recovery, request_logger, body_limit
- Built-in services: database pools, caching, metrics, health checks

## Running the Application

```bash
go run .
```
```

### Option 3: Include Both Approaches

For maximum AI assistant effectiveness, use both comments and README.md:

**In your main.go:**
```go
// This application uses Lokstra framework for web development.
// See README.md for project documentation and framework features.
func main() {
    // ... your Lokstra application code
}
```

**In your README.md:**
Include framework information and link to the docs folder.

### AI Assistant Benefits

With proper documentation, AI assistants can:

✅ **Understand Lokstra patterns** - Request binding, response helpers, middleware  
✅ **Suggest correct syntax** - Handler signatures, context methods, configuration  
✅ **Provide better completions** - Service registration, route definitions  
✅ **Help with debugging** - Common patterns and error handling  
✅ **Generate boilerplate** - Handlers, middleware, configuration files  

### Example AI-Friendly Project Structure

```
my-lokstra-project/
├── README.md                 # Project overview with Lokstra info
├── main.go                   # Main application with descriptive comments
├── docs/                     # Copy of Lokstra documentation
│   ├── README.md
│   ├── getting-started.md
│   └── ...
├── configs/
│   └── server.yaml           # YAML configuration
├── handlers/
│   └── user_handlers.go      # Handler functions
└── go.mod
```

The comment approach is sufficient for basic AI assistance, but combining it with a README.md provides the most comprehensive context for AI agents to understand and help with your Lokstra project effectively.

## Next Steps

- [Core Concepts](./core-concepts.md) - Understanding Lokstra's architecture
- [Routing](./routing.md) - Advanced routing features
- [Configuration](./configuration.md) - Detailed YAML configuration
- [Examples](./examples.md) - Complete example applications

---

*For more examples, check the `/cmd/examples/` directory in the repository.*