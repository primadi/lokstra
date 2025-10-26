# Response Priority Rules

## Overview

Ketika handler menggunakan **kombinasi** dari:
- `c.Resp` / `c.Api` (context methods)
- Return `*response.Response` / `*response.ApiHelper`
- Return error

Framework perlu menentukan mana yang digunakan. Berikut adalah **priority rules** yang diterapkan.

---

## 🎯 **Priority Order (High to Low)**

```
1. Error (non-nil)                    ← HIGHEST PRIORITY
2. Return Response/ApiHelper (non-nil) ← Overrides c.Resp/c.Api
3. c.Resp/c.Api                       ← Used if no return value
4. Default success (Api.Ok(nil))      ← Fallback
```

---

## 📋 **Decision Matrix**

| c.Resp/c.Api Set? | Return Value? | Error? | **Result** |
|-------------------|---------------|--------|------------|
| ❌ No | ❌ No | ❌ No | Default: `Api.Ok(nil)` |
| ✅ Yes | ❌ No | ❌ No | Use `c.Resp`/`c.Api` |
| ❌ No | ✅ Yes (non-nil) | ❌ No | Use return value |
| ✅ Yes | ✅ Yes (non-nil) | ❌ No | **Use return value** (override) |
| ✅ Yes | ✅ Yes (nil) | ❌ No | Default: `Api.Ok(nil)` |
| Any | Any | ✅ Yes | **Return error** (ignore everything) |

---

## 🔍 **Detailed Examples**

### **Rule 1: Error Always Wins** ⚠️

```go
func Handler(c *request.Context) (*response.Response, error) {
    // Set via c.Resp
    c.Resp.WithStatus(200).Json(map[string]string{"status": "ok"})
    
    // Return Response with success
    resp := response.NewResponse()
    resp.WithStatus(201).Json(map[string]string{"created": "yes"})
    
    // ERROR TAKES PRECEDENCE - both above IGNORED!
    return resp, errors.New("something failed")
}
```

**Result:** Error is returned, status depends on error handling middleware.

---

### **Rule 2: Return Value Overrides c.Resp/c.Api** ✅

```go
func Handler(c *request.Context) (*response.Response, error) {
    // Set via c.Resp (WILL BE IGNORED!)
    c.Resp.WithStatus(200).Json(map[string]string{
        "source": "c.Resp",
    })
    
    // Return Response (THIS IS USED!)
    resp := response.NewResponse()
    resp.WithStatus(201).Json(map[string]string{
        "source": "return",
    })
    return resp, nil
}
```

**Result:** 
- Status: `201` (from return value)
- Body: `{"source": "return"}`
- c.Resp is **completely replaced** by return value

**Why?** Return value is more explicit and intentional.

---

### **Rule 3: Regular Data Return Overrides c.Api** ✅

```go
func Handler(c *request.Context) (map[string]string, error) {
    // Set via c.Api (WILL BE IGNORED!)
    c.Api.Created(map[string]string{"source": "c.Api"}, "Message")
    
    // Return data (THIS IS USED and wrapped with Api.Ok!)
    return map[string]string{"source": "return"}, nil
}
```

**Result:**
- Status: `200` (from Api.Ok wrapping, NOT 201 from c.Api.Created)
- Body: Wrapped with success response format
- Data: `{"source": "return"}`

---

### **Rule 4: Nil Return Triggers Default** 🔄

```go
func Handler(c *request.Context) (*response.Response, error) {
    // Set via c.Resp
    c.Resp.WithStatus(202).Json(map[string]string{"source": "c.Resp"})
    
    // Return nil (triggers default success!)
    return nil, nil
}
```

**Result:**
- Status: `200` (default success)
- Body: Empty or `{}`
- c.Resp is **ignored** because nil return is explicit

**Note:** Nil return value is treated as "I want default success, ignore c.Resp"

---

### **Rule 5: c.Resp/c.Api Used When No Return Value** ✅

```go
func Handler(c *request.Context) error {
    // Set via c.Resp (THIS IS USED - no return value)
    c.Resp.WithStatus(202).Json(map[string]string{
        "source": "c.Resp",
    })
    return nil
}
```

**Result:**
- Status: `202`
- Body: `{"source": "c.Resp"}`

**Why?** No return value means "use whatever is in c.Resp"

---

## 🤔 **Why These Rules?**

### **1. Error Priority = Safety** 🛡️
Error menunjukkan failure, harus selalu ditangani. Mengabaikan error berbahaya.

### **2. Return Value Priority = Explicit Intent** 📝
```go
// IMPLICIT (developer might forget it's set)
c.Resp.WithStatus(200).Json(data)

// EXPLICIT (clear intent to override)
return response.NewResponse().WithStatus(201).Json(data), nil
```

Explicit always wins over implicit.

### **3. Type Safety** 🔒
```go
// Return value is type-checked at compile time
func Handler() (*response.Response, error) {
    return response.NewResponse(), nil  // ✅ Compile-time safe
}

// c.Resp can be any state at runtime
func Handler(c *request.Context) error {
    c.Resp.WithStatus(200)  // State can be anything
    return nil
}
```

### **4. Functional Programming Style** 🎯
Return value mendukung functional style yang lebih predictable:
```go
// Functional - pure function style
func buildResponse(data any) *response.Response {
    resp := response.NewResponse()
    resp.WithStatus(200).Json(data)
    return resp
}

func Handler(c *request.Context) (*response.Response, error) {
    return buildResponse(data), nil  // Clear data flow
}
```

---

## ⚠️ **Anti-Patterns (Don't Do This)**

### **❌ Anti-Pattern 1: Set c.Resp Then Return Response**

```go
// DON'T DO THIS - confusing and wasteful
func Handler(c *request.Context) (*response.Response, error) {
    c.Resp.WithStatus(200).Json(data1)  // WASTED EFFORT
    
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data2)
    return resp, nil  // This is used, c.Resp ignored
}
```

**Fix:** Choose one approach, don't mix!

```go
// GOOD - consistent approach
func Handler(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data)
    return resp, nil
}
```

---

### **❌ Anti-Pattern 2: Multiple c.Api Calls Then Return**

```go
// DON'T DO THIS - confusing
func Handler(c *request.Context) (any, error) {
    c.Api.Ok(data1)     // IGNORED
    c.Api.Created(data2, "msg")  // IGNORED
    
    return data3, nil  // This is used
}
```

**Fix:** Just return data!

```go
// GOOD - clear intent
func Handler(c *request.Context) (any, error) {
    return data3, nil
}
```

---

### **❌ Anti-Pattern 3: Conditional Mix**

```go
// DON'T DO THIS - hard to reason about
func Handler(c *request.Context) (*response.Response, error) {
    if condition {
        c.Resp.WithStatus(200).Json(data1)
        return nil, nil  // Which response? Confusing!
    }
    
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data2)
    return resp, nil
}
```

**Fix:** Be consistent!

```go
// GOOD - consistent return style
func Handler(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    
    if condition {
        resp.WithStatus(200).Json(data1)
    } else {
        resp.WithStatus(201).Json(data2)
    }
    
    return resp, nil
}
```

---

## ✅ **Best Practices**

### **1. Choose ONE Style Per Handler** 🎯

```go
// Style A: Context methods (simple cases)
func SimpleHandler(c *request.Context) error {
    return c.Api.Ok(data)
}

// Style B: Return value (complex cases)
func ComplexHandler(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.WithStatus(201).Json(data)
    return resp, nil
}

// DON'T MIX both styles in same handler!
```

---

### **2. Middleware Uses c.Resp/c.Api ONLY** 🔧

```go
// Middleware CANNOT return Response
func LoggingMiddleware(c *request.Context) error {
    start := time.Now()
    
    err := c.Next()
    
    // Access c.Resp for logging
    log.Printf("Status: %d, Duration: %v", 
        c.Resp.RespStatusCode, 
        time.Since(start))
    
    return err
}
```

**Why?** Middleware must call `c.Next()` and cannot hijack response.

---

### **3. Early Returns Use c.Api** ✅

```go
func Handler(c *request.Context) error {
    // Early validation
    if !isValid {
        return c.Api.BadRequest("ERR001", "Invalid input")
    }
    
    // Early auth check
    if !isAuthorized {
        return c.Api.Unauthorized("Not authorized")
    }
    
    // Main logic
    data := process()
    return c.Api.Ok(data)
}
```

Simple and readable!

---

### **4. Complex Response Uses Return** ✅

```go
func DownloadFile(c *request.Context) (*response.Response, error) {
    resp := response.NewResponse()
    resp.RespHeaders = map[string][]string{
        "Content-Disposition": {"attachment; filename=file.pdf"},
        "X-File-Size": {strconv.Itoa(fileSize)},
    }
    resp.Stream("application/pdf", func(w http.ResponseWriter) error {
        return writeFile(w, file)
    })
    return resp, nil
}
```

Explicit and type-safe!

---

## 🧪 **Testing Priority Rules**

Tests in `core/router/helper_priority_test.go`:

- ✅ Return value overrides c.Resp
- ✅ Return ApiHelper overrides c.Api
- ✅ Error overrides everything
- ✅ c.Resp used when no return value
- ✅ Nil return triggers default success
- ✅ Return overrides even WriterFunc
- ✅ Return overrides multiple Api calls
- ✅ Regular data return overrides c.Api

Run tests:
```bash
go test ./core/router -run TestPriority -v
```

---

## 📊 **Summary Table**

| Scenario | Handler Uses | Result | Reason |
|----------|-------------|--------|--------|
| Simple success | `c.Api.Ok()` | Use c.Api | Simple & clear |
| Custom response | `return *Response` | Use return | Explicit control |
| Mixed both | Both set | **Use return** | Explicit wins |
| Error case | `return err` | Use error | Safety first |
| Middleware | `c.Resp`/`c.Api` | Use context | Can't return |
| Nil return | `return nil, nil` | Default success | Explicit nil |

---

## 🎓 **When to Use What**

### Use `c.Resp` / `c.Api`:
- ✅ Simple handlers (CRUD)
- ✅ Middleware (required!)
- ✅ Early returns (validation, auth)
- ✅ Standard API responses

### Use `return *Response`:
- ✅ Custom content-type (HTML, XML)
- ✅ Streaming responses
- ✅ File downloads
- ✅ Complex header manipulation
- ✅ Need explicit type safety

### Don't Mix:
- ❌ Setting c.Resp + returning Response
- ❌ Multiple c.Api calls + return data
- ❌ Conditional mix of both styles

---

## 🔗 **Related Documentation**

- Response Return Types: `docs_draft/response-return-types.md`
- Quick Reference: `docs_draft/RESPONSE-RETURN-TYPES-QUICKREF.md`
- Tests: `core/router/helper_priority_test.go`

---

**Remember: Return value ALWAYS wins over c.Resp/c.Api (except error)!** ✅
