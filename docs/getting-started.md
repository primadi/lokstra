# Getting Started with Lokstra

Lokstra apps can be started **by code** (imperative) or **by configuration** (declarative). This page shows both, plus graceful shutdown.

---

## 1) Installation

```bash
go get github.com/primadi/lokstra
```

---

## 2) Start by Code

### 2.1 Single App

```go
package main

import (
	"time"
	"github.com/primadi/lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	// Register handler (named or inline). Router is embedded on App.
	// Preferred shorthand:
	app := lokstra.NewApp(ctx, "demo-app", ":8080")
	app.GET("/hello", func(c *lokstra.RequestContext) {
		c.Ok(map[string]any{"message": "Hello, Lokstra!"})
	})

	// Equivalent long form (both are valid):
	// app.Router.GET("/hello", ctx.GetHandler("hello"))

	// Start the app
	if err := app.Start(); err != nil {
		panic(err)
	}
}
```

**Graceful shutdown (recommended in production):**

```go
if err := app.StartAndWaitForShutdown(30 * time.Second); err != nil {
	panic(err)
}
```

> **Note**: `App.Router` is an **embedded field**. You may call route methods via `app.GET(...)` / `app.POST(...)` or via `app.Router.GET(...)`. They are equivalent.

---

### 2.2 Multiple Apps via Server

```go
package main

import (
	"time"
	"github.com/primadi/lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	server := lokstra.NewServer(ctx, "demo-server")

	app1 := lokstra.NewApp(ctx, "app1", ":8080")
	app1.GET("/a1", func(c *lokstra.RequestContext) { c.Ok("app1") })

	app2 := lokstra.NewApp(ctx, "app2", ":8080") // same port → will be merged
	app2.GET("/a2", func(c *lokstra.RequestContext) { c.Ok("app2") })

	server.AddApp(app1).AddApp(app2)

	// Start all apps
	if err := server.Start(); err != nil {
		panic(err)
	}

	// Or with graceful shutdown
	// if err := server.StartAndWaitForShutdown(30 * time.Second); err != nil { panic(err) }
}
```

> **Behavior on same port**: If two or more Apps specify the **same port**, Lokstra **merges** them into a single listener when running.

---

## 3) Start by Configuration

Lokstra supports YAML-based configuration. You can load **a single file** or **an entire directory**.

### 3.1 Load a single YAML file

```go
package main

import (
	"fmt"
	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/config"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	// Register named handlers used by config-defined routes (if any)
	ctx.RegisterHandler("hello", func(c *lokstra.RequestContext) {
		c.Ok(map[string]any{"message": "Hello from YAML route!"})
	})

	cfg, err := config.LoadConfigFile("lokstra.yaml")
	if err != nil { panic(fmt.Sprintf("Failed to load config: %v", err)) }

	server, err := lokstra.NewServerFromConfig(ctx, cfg)
	if err != nil { panic(fmt.Sprintf("Failed to create server: %v", err)) }

	if err := server.Start(); err != nil { panic(err) }
}
```

### 3.2 Load a directory of YAML files

```go
package main

import (
	"fmt"
	"github.com/primadi/lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	cfg, err := lokstra.LoadConfigDir("./configs")
	if err != nil { panic(fmt.Sprintf("Failed to load config dir: %v", err)) }

	server, err := lokstra.NewServerFromConfig(ctx, cfg)
	if err != nil { panic(fmt.Sprintf("Failed to create server from config: %v", err)) }

	if err := server.Start(); err != nil { panic(err) }
}
```

> **Notes**
>
> * `lokstra.LoadConfigDir(dir)` loads **all `.yaml` files** in a directory (and merges them).
> * To load **one** YAML file, prefer `config.LoadConfigFile(path)`.
> * If multiple Apps share the **same port**, they will be **merged** at runtime.
> * You can also run `server.StartAndWaitForShutdown(timeout)` for graceful shutdown.

---

## 4) Minimal `lokstra.yaml`

```yaml
server:
  name: demo-server

apps:
  - name: demo-app
    port: 8080
    routes:
      - method: GET
        path: /hello
        handler: hello   # must exist as a registered handler
```

**Wiring the named handler in code (when starting from config):**

```go
ctx.RegisterHandler("hello", func(c *lokstra.RequestContext) {
    c.Ok(map[string]any{"message": "Hello from YAML!"})
})
```

---

## 5) Graceful Shutdown

Both App and Server support graceful shutdown:

```go
// App level
a := lokstra.NewApp(ctx, "demo-app", ":8080")
_ = a.StartAndWaitForShutdown(30 * time.Second)

// Server level
s := lokstra.NewServer(ctx, "demo-server")
_ = s.StartAndWaitForShutdown(30 * time.Second)
```

The call waits up to the given timeout for the internal shutdown process to complete (e.g., finishing in-flight requests and closing listeners). Refer to the source for exact steps.

---

## 6) Quick Recap

* **By code**: `app.Start()` for one app; `server.Start()` for many apps.
* **By config**: `config.LoadConfigFile` (single file) or `lokstra.LoadConfigDir` (folder) → `lokstra.NewServerFromConfig(ctx, cfg)` → `Start()`.
* **Embedded Router**: `app.GET(...)` or `app.Router.GET(...)` are equivalent.
* **Same-port merge**: Apps with the same port are merged.
* **Graceful shutdown**: Use `StartAndWaitForShutdown(timeout)`.
