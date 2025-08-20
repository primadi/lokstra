# Recovery Middleware

Recovery middleware untuk menangani panic dan mencegah aplikasi crash. Middleware ini akan menangkap panic yang terjadi dalam handler dan mengembalikan response error yang sesuai.

## Fitur

- Menangkap panic dari handler
- Mencatat error log dengan detail lengkap
- Mengembalikan HTTP 500 Internal Server Error
- Konfigurasi untuk mengaktifkan/nonaktifkan stack trace
- Tetap menjaga request ID untuk tracing

## Konfigurasi

```go
type Config struct {
    EnableStackTrace bool `json:"enable_stack_trace" yaml:"enable_stack_trace"`
}
```

### Parameter Konfigurasi

- `enable_stack_trace` (bool): Mengontrol apakah stack trace disertakan dalam log error
  - Default: `true`
  - `true`: Stack trace akan dicatat dalam log untuk debugging
  - `false`: Hanya pesan panic yang dicatat, tanpa stack trace

## Cara Penggunaan

### 1. Penggunaan Dasar (dengan stack trace)

```go
// Menggunakan konfigurasi default dengan stack trace aktif
router.Use(recovery.GetModule().GetFactory()(nil))
```

### 2. Dengan Konfigurasi Map

```go
// Mengaktifkan stack trace
config := map[string]any{
    "enable_stack_trace": true,
}
router.Use(recovery.GetModule().GetFactory()(config))

// Menonaktifkan stack trace untuk production
config := map[string]any{
    "enable_stack_trace": false,
}
router.Use(recovery.GetModule().GetFactory()(config))
```

### 3. Dengan Struct Config

```go
config := &recovery.Config{
    EnableStackTrace: false, // untuk production environment
}
router.Use(recovery.GetModule().GetFactory()(config))
```

## Contoh Response

Ketika terjadi panic, middleware akan mengembalikan response berikut:

```json
{
    "success": false,
    "message": "Internal Server Error",
    "code": "INTERNAL"
}
```

## Log Output

### Dengan Stack Trace Aktif (Development)

```json
{
    "level": "error",
    "msg": "Recovered from panic in middleware",
    "error": "division by zero",
    "request_id": "req-12345",
    "url": "/api/users",
    "method": "GET",
    "stack": "goroutine 1 [running]:\nruntime/debug.Stack()..."
}
```

### Tanpa Stack Trace (Production)

```json
{
    "level": "error", 
    "msg": "Recovered from panic in middleware",
    "error": "division by zero",
    "request_id": "req-12345", 
    "url": "/api/users",
    "method": "GET"
}
```

## Contoh Implementasi

```go
package main

import (
    "github.com/primadi/lokstra"
    "github.com/primadi/lokstra/middleware/recovery"
)

func main() {
    regCtx := lokstra.NewRegistrationContext()
    
    // Register recovery middleware
    recovery.GetModule().Register(regCtx)
    
    // Create server
    server := lokstra.NewServer(regCtx, "my-app")
    
    // Untuk development dengan stack trace
    devConfig := map[string]any{
        "enable_stack_trace": true,
    }
    server.Use(recovery.GetModule().GetFactory()(devConfig))
    
    // Untuk production tanpa stack trace
    prodConfig := map[string]any{
        "enable_stack_trace": false,
    }
    // server.Use(recovery.GetModule().GetFactory()(prodConfig))
    
    // Handler yang mungkin panic
    server.GET("/test", func(ctx *lokstra.Context) error {
        panic("something went wrong!")
        return nil
    })
    
    server.Listen(":8080")
}
```

## Best Practices

### 1. Environment-based Configuration

```go
config := map[string]any{
    "enable_stack_trace": os.Getenv("ENV") != "production",
}
```

### 2. Selective Stack Trace

Untuk production, pertimbangkan untuk menonaktifkan stack trace untuk mengurangi ukuran log dan menghemat performa:

```go
// Production
config := map[string]any{
    "enable_stack_trace": false,
}

// Development/Staging
config := map[string]any{
    "enable_stack_trace": true,
}
```

### 3. Monitoring Integration

Recovery middleware bekerja dengan baik dengan sistem monitoring. Log error dapat diteruskan ke sistem seperti:
- Sentry
- Rollbar  
- New Relic
- CloudWatch

## Testing

Recovery middleware dilengkapi dengan test suite yang komprehensif:

```bash
go test ./middleware/recovery -v
```

Test mencakup:
- Parsing konfigurasi berbagai format
- Recovery dari panic dengan stack trace
- Recovery dari panic tanpa stack trace
- Eksekusi normal tanpa panic
- Edge cases konfigurasi
