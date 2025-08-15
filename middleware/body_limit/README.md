# Body Limit Middleware

Middleware untuk membatasi ukuran request body guna mencegah serangan DoS.

## Features

- **Early Validation**: Cek Content-Length header untuk reject request besar sebelum baca body
- **Runtime Protection**: Limit pembacaan body dengan LimitReader
- **Configurable**: Bisa set custom limit size dan skip patterns
- **Security**: Return 413 Request Entity Too Large untuk body yang terlalu besar
- **Convenience Functions**: Predefined functions untuk ukuran umum

## Usage

### Basic Usage

```go
import "github.com/primadi/lokstra/middleware/body_limit"

// Menggunakan preset 1MB limit
app.Use(body_limit.BodyLimit1MB())

// Menggunakan preset 5MB limit  
app.Use(body_limit.BodyLimit5MB())

// Custom limit
app.Use(body_limit.BodyLimit(2 * 1024 * 1024)) // 2MB
```

### Advanced Configuration

```go
app.Use(body_limit.BodyLimitMiddleware(body_limit.Config{
    Limit: 1024 * 1024, // 1MB
    Skipper: func(ctx *request.Context) bool {
        // Skip untuk endpoint file upload
        return ctx.Request.URL.Path == "/upload"
    },
}))
```

## Available Presets

- `BodyLimit1MB()` - 1 MB limit
- `BodyLimit5MB()` - 5 MB limit  
- `BodyLimit10MB()` - 10 MB limit
- `BodyLimit50MB()` - 50 MB limit
- `BodyLimit100MB()` - 100 MB limit

## Testing

```bash
cd middleware/body_limit
go test -v
```

## Example

Lihat `example/main.go` untuk contoh penggunaan lengkap.
