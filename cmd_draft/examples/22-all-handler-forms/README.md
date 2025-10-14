# Example 22: All Handler Forms Test

Contoh yang menguji **semua 9 bentuk handler** yang didukung oleh Lokstra router.

## 📋 Handler Forms yang Didukung

### 1. `func() error`
Handler paling sederhana tanpa parameter.
```go
func Handler1NoParams() error {
    return nil
}
```

### 2. `func() (data, error)`
Handler tanpa parameter yang mengembalikan data.
```go
func Handler2NoParamsWithData() (any, error) {
    return map[string]string{"message": "success"}, nil
}
```

### 3. `func(*T) error`
Handler dengan struct parameter (path, query, body via tags).
```go
type UserRequest struct {
    ID   string `path:"id"`
    Name string `query:"name"`
}

func Handler3StructOnly(req *UserRequest) error {
    return nil
}
```

### 4. `func(*T) (data, error)`
Handler dengan struct parameter yang mengembalikan data.
```go
func Handler4StructWithData(req *UserRequest) (any, error) {
    return req, nil
}
```

### 5. `func(*request.Context) error`
Handler dengan akses penuh ke request context.
```go
func Handler5ContextOnly(ctx *request.Context) error {
    id := ctx.Req.PathParam("id", "")
    return nil
}
```

### 6. `func(*request.Context) (data, error)`
Handler dengan context yang mengembalikan data.
```go
func Handler6ContextWithData(ctx *request.Context) (any, error) {
    return map[string]string{"method": ctx.R.Method}, nil
}
```

### 7. `func(*request.Context, *T) error`
Handler dengan context dan struct parameter.
```go
func Handler7ContextAndStruct(ctx *request.Context, req *UserRequest) error {
    // Akses context dan struct
    return nil
}
```

### 8. `func(*request.Context, *T) (data, error)`
Handler paling lengkap: context + struct + return data.
```go
func Handler8ContextAndStructWithData(ctx *request.Context, req *UserRequest) (any, error) {
    return map[string]any{
        "method": ctx.R.Method,
        "request": req,
    }, nil
}
```

### 9. `http.HandlerFunc`
Standard Go handler untuk kompatibilitas.
```go
func Handler9StandardHTTP(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"message":"success"}`))
}
```

## 🚀 Cara Menjalankan

### Build dan Run
```bash
# Build
go build

# Run
go run main.go

# Atau jalankan executable
./22-all-handler-forms
```

### Testing dengan curl
```bash
# Form 1: func() error
curl http://localhost:3000/form1

# Form 2: func() (data, error)
curl http://localhost:3000/form2

# Form 3: func(*T) error
curl "http://localhost:3000/form3/user123?name=Alice"

# Form 4: func(*T) (data, error)
curl "http://localhost:3000/form4/user456?name=Bob"

# Form 5: func(*request.Context) error
curl http://localhost:3000/form5/ctx789

# Form 6: func(*request.Context) (data, error)
curl http://localhost:3000/form6/ctx999

# Form 7: func(*request.Context, *T) error
curl -X POST "http://localhost:3000/form7/post123?name=Charlie"

# Form 8: func(*request.Context, *T) (data, error)
curl -X POST -H "Content-Type: application/json" -d '{"name":"Diana"}' http://localhost:3000/form8/post456

# Form 9: http.HandlerFunc
curl http://localhost:3000/form9
```

### Testing dengan REST Client (VS Code)
1. Install extension: **REST Client** by Huachao Mao
2. Open file: `test.http`
3. Click "Send Request" di atas setiap HTTP request
4. Lihat response di panel sebelah kanan

## 📝 Catatan Penting

### Struct-Based Parameters
- Path parameters: `path:"id"`
- Query parameters: `query:"name"`
- Body parameters: `json:"email"`
- Semua bisa dikombinasikan dalam satu struct

### Direct Path Parameters (TIDAK DIDUKUNG)
❌ **Tidak didukung lagi sejak optimasi:**
```go
// TIDAK DIDUKUNG
func GetUser(ctx *request.Context, depID string, userID string) error
```

✅ **Gunakan struct dengan tags:**
```go
type GetUserRequest struct {
    DepartmentID string `path:"dep"`
    UserID       string `path:"id"`
}

func GetUser(ctx *request.Context, req *GetUserRequest) error
```

### Keuntungan Struct-Based
1. **Type-safe**: Compiler checking
2. **Self-documenting**: Tags menjelaskan sumber parameter
3. **Flexible**: Bisa combine path + query + body
4. **Validation**: Bisa tambahkan `validate` tags
5. **No reflection limitation**: Tag names explicit

## 🎯 Output yang Diharapkan

Setiap test akan print ke console dan return JSON response:

```
✅ Form 1: func() error
✅ Form 2: func() (data, error)
✅ Form 3: func(*T) error - ID: user123, Name: Alice
✅ Form 4: func(*T) (data, error) - ID: user456, Name: Bob
✅ Form 5: func(*request.Context) error
   Path param ID: ctx789
✅ Form 6: func(*request.Context) (data, error)
✅ Form 7: func(*request.Context, *T) error - ID: post123, Name: Charlie
   Request method: POST
✅ Form 8: func(*request.Context, *T) (data, error) - ID: post456, Name: Diana
✅ Form 9: http.HandlerFunc (standard Go handler)
```

## 🔍 Implementasi Detail

### Router Setup
```go
r := router.New("")
r.GET("/form1", Handler1NoParams)
r.GET("/form2", Handler2NoParamsWithData)
// ... dst
```

### Smart Handler Adaptation
Router secara otomatis mendeteksi signature handler dan:
1. **Registration Time** (once): 
   - Build metadata via reflection
   - Create parameter extractors with closures
   - Pre-calculate args capacity

2. **Per-Request Time** (hot path):
   - Pre-allocated args slice
   - Extract parameters via compiled extractors
   - Call function with reflect.Call
   - Handle return values

### Performance Optimizations
- ✅ Metadata extracted once (registration time)
- ✅ Extractors created once with closures
- ✅ Per-request handler minimal overhead
- ✅ Pre-allocated args slice with exact capacity
- ✅ Branch prediction optimization (common cases first)
- ✅ Only struct-based parameters (simplified logic)

## 📚 Related Examples
- Example 18: Service Router basics
- Example 20: Service Router with struct-based parameters
- Example 21: Type alias context detection
