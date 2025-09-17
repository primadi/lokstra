# Request Context

The **RequestContext** (`ctx`) is passed into every handler in Lokstra.  
It provides unified access to the HTTP request, response, and smart binding utilities.

---

## 🔑 Structure

```go
type RequestContext struct {
    context.Context
    *response.Response
    Writer  http.ResponseWriter
    Request *http.Request
}
```

- Embeds Go’s `context.Context` for cancellation & deadlines  
- Embeds a `Response` for standardized output  
- Provides direct access to `http.ResponseWriter` and `*http.Request`

---

## 🧩 Handler Signature

All handlers in Lokstra **return `error`**:

- Standard form:
  ```go
  func(ctx *lokstra.RequestContext) error
  ```

- Generic auto-binding form:
  ```go
  func[T any](ctx *lokstra.RequestContext, params *T) error
  ```

👉 Returning a raw error → **500 Internal Server Error**.  
Use response helpers (`ctx.ErrorBadRequest`, `ctx.ErrorNotFound`, etc.) for application errors.

---

## 📥 Binding

Lokstra can automatically bind request data to struct fields, using **tags**.  
Tags are **required** — without a tag, the field won’t be auto-filled.

### 1. Standard Binding
- `BindQuery(&dto)` → from query string  
- `BindPath(&dto)` → from path params  
- `BindHeader(&dto)` → from headers  
- `BindBody(&dto)` → from body (**JSON only**)  

### 2. Combined Binding
- `BindAll(&dto)` → Path + Query + Header + Body (JSON only)  

### 3. Smart Binding
- `BindBodySmart(&dto)` → body auto-detect (JSON, form-urlencoded, multipart, text)  
- `BindAllSmart(&dto)` → Path + Query + Header + Body (smart detect)  

---

## 🏷️ Tag Rules

| Tag     | Source            | Example                              |
|---------|------------------|--------------------------------------|
| `path`  | Path parameters   | `/users/:id` → `ID string \`path:"id"\`` |
| `query` | Query string      | `/users?active=true` → `Active bool \`query:"active"\`` |
| `header`| HTTP headers      | `Authorization: Bearer ...` → `Token string \`header:"Authorization"\`` |
| `body`  | Request body      | `{ "name": "Prim" }` → `Name string \`body:"name"\`` |

---

## 🔍 Type Conversion

Binding automatically converts:
- `"123"` → `int`
- `"true"` → `bool`
- `"2025-09-16"` → `time.Time`
- `"1.23"` → `decimal.Decimal`
- `"a,b,c"` → `[]string{"a","b","c"}`

Supports nested struct binding as well.

---

## 🧰 Examples

### Manual Binding
```go
type UserRequest struct {
    ID    string `path:"id"`
    Token string `header:"Authorization"`
    Name  string `body:"name"`
}

func createUser(ctx *lokstra.RequestContext) error {
    var req UserRequest
    if err := ctx.BindAllSmart(&req); err != nil {
        return ctx.ErrorBadRequest(err.Error())
    }
    return ctx.Ok("Created user " + req.Name)
}
```

### Generic Handler
```go
func createUser(ctx *lokstra.RequestContext, req *UserRequest) error {
    return ctx.Ok("Created user " + req.Name)
}
```

👉 In the generic form, Lokstra auto-binds `*UserRequest`.  
If binding/validation fails, it prepares a 400 response automatically.

---

## ✅ Summary

- `RequestContext` unifies request, response, and context.  
- Handlers always return `error`. Use response helpers for non-500 outcomes.  
- Binding requires struct tags (`path`, `query`, `header`, `body`).  
- Use `BindAllSmart` or generic handler for most cases.  
