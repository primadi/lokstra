# Response

**Source of truth:** This document is generated from the `Response` type and its methods in `lokstra-0.2.1` to remain accurate for both programmers and AI agents.

Lokstra uses a structured `Response` object for every handler. Instead of writing raw JSON, return response helpers (which set status, message, headers, etc.) and **always return an `error`** from handlers. Returning a *plain* error signals an internal failure (HTTP **500**). Use the helpers below for 2xx/4xx/5xx outcomes.

---

## Struct

```go
type Response struct {
    StatusCode   int          `json:"-"`              // HTTP status
    ResponseCode ResponseCode `json:"code,omitempty"` // Logical code
    Success      bool         `json:"success"`        // true/false

    Message string      `json:"message,omitempty"` // Localizable message
    Data    any         `json:"data,omitempty"`    // Payload
    Meta    any         `json:"meta,omitempty"`    // List/pagination or extra info
    Headers http.Header `json:"-"`                 // HTTP-only
    RawData []byte      `json:"-"`                 // Non-JSON body

    FieldErrors map[string]string `json:"errors,omitempty"` // Form/field-level errors
}
```

---

## Method Reference

### Success

- `Ok(data any) error` — 200 OK
- `OkCreated(data any) error` — 201 Created
- `OkUpdated(data any) error` — 200 OK (typical for update)
- `OkList(data any, meta any) error` — 200 OK with `Meta` payload (e.g., pagination info)

### Errors

- `ErrorBadRequest(msg string) error` — 400
- `ErrorValidation(globalMsg string, fieldErrors map[string]string) error` — 400 with `errors` map
- `ErrorNotFound(msg string) error` — 404
- `ErrorDuplicate(msg string) error` — 409 (conflict/duplicate)
- `ErrorInternal(msg string) error` — 500
- `ErrorHTML(status int, html string) error` — send HTML error with custom HTTP status

> Tip: Use these helpers instead of returning the raw error value. Raw `error` → 500.

### HTMX

- `HtmxPageData(title string, description string, data map[string]any) error`  
  Produces a JSON payload for **page-data** endpoints used by `MountHtmx(...)` conventions (title, description, data).

### Raw / HTML

- `HTML(html string) error` — 200 `text/html`
- `WriteRaw(contentType string, status int, data []byte) error` — write non-JSON response body

### Headers / Meta / Message (Chaining)

These return `*Response` so you can chain after a helper:

- `WithHeader(key, value string) *Response`
- `WithMessage(msg string) *Response`
- `WithMeta(meta any) *Response`
- `WithData(data any) *Response`
- `SetStatusCode(code int) *Response`
- `GetHeaders() http.Header`
- `GetStatusCode() int`

### Output (Framework Use)

- `WriteHttp(w http.ResponseWriter) error` — low-level writer (invoked by framework)

---

## Usage Patterns

### 1) Success

```go
func getUser(ctx *lokstra.RequestContext) error {
    user := User{ID: "1", Name: "Prim"}
    return ctx.Ok(user).
        WithMessage("User fetched")
}
```

### 2) List + Meta (Pagination)

```go
func listUsers(ctx *lokstra.RequestContext) error {
    users, total, page, size := repo.FindAll(), 120, 1, 10
    meta := map[string]any{"total": total, "page": page, "pageSize": size}
    return ctx.OkList(users, meta)
}
```

### 3) Validation Errors (400)

```go
func createUser(ctx *lokstra.RequestContext) error {
    fields := map[string]string{"email": "invalid format"}
    return ctx.ErrorValidation("Please check your input", fields)
}
```

### 4) HTMX Page-Data

```go
func dashboardPageData(ctx *lokstra.RequestContext) error {
    data := map[string]any{"stats": 42}
    return ctx.HtmxPageData("Dashboard", "Today summary", data)
}
```

### 5) HTML / Raw

```go
func helloHTML(ctx *lokstra.RequestContext) error {
    return ctx.HTML("<h1>Hello</h1>")
}

func exportCSV(ctx *lokstra.RequestContext) error {
    return ctx.WriteRaw("text/csv", 200, []byte("id,name
1,Prim
"))
}
```

---

## Handler Rule of Thumb

- Handlers **must** return `error`.  
- Returning a raw `error` → HTTP **500** by default.  
- For domain/validation errors, return a **Response helper** (e.g., `ErrorBadRequest`, `ErrorValidation`) so the correct HTTP status and JSON shape are produced.  
- Success helpers (`Ok`, `OkCreated`, `OkUpdated`, `OkList`) already set `Success=true`, `StatusCode`, and `Data`.  
- Chain `WithMessage`, `WithHeader`, `WithMeta`, or `SetStatusCode` as needed.

This contract keeps responses consistent for clients and AI agents consuming your API.
