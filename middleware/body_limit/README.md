# Body Limit Middleware

Body limit middleware untuk membatasi ukuran request body dan mencegah serangan yang menggunakan payload besar untuk menghabiskan memori server.

## Fitur

- Membatasi ukuran request body berdasarkan konfigurasi
- Mendukung pemeriksaan Content-Length header untuk deteksi awal
- Konfigurasi pesan error dan status code yang dapat disesuaikan
- Dukungan untuk skip paths menggunakan pattern matching
- Opsi untuk skip payload besar alih-alih mengembalikan error
- Convenience functions untuk size limit umum

## Konfigurasi

```go
type Config struct {
    MaxSize           int64    `json:"max_size" yaml:"max_size"`
    SkipLargePayloads bool     `json:"skip_large_payloads" yaml:"skip_large_payloads"`
    Message           string   `json:"message" yaml:"message"`
    StatusCode        int      `json:"status_code" yaml:"status_code"`
    SkipOnPath        []string `json:"skip_on_path" yaml:"skip_on_path"`
}
```

### Parameter Konfigurasi

- `max_size` (int64): Ukuran maksimum request body dalam bytes
  - Default: `10485760` (10MB)
  - Contoh: `1048576` (1MB), `5242880` (5MB)

- `skip_large_payloads` (bool): Jika true, skip pembacaan body yang melebihi limit
  - Default: `false`
  - `true`: Lanjutkan processing tanpa membaca body
  - `false`: Return error ketika limit terlampaui

- `message` (string): Pesan error custom untuk payload yang terlalu besar
  - Default: `"Request body too large"`

- `status_code` (int): HTTP status code yang dikembalikan
  - Default: `413` (Request Entity Too Large)
  - Rentang valid: 400-599

- `skip_on_path` ([]string): Array pattern path yang akan di-skip dari pengecekan limit
  - Default: `[]` (kosong)
  - Mendukung wildcard patterns: `*`, `**`

## Cara Penggunaan

### 1. Penggunaan Dasar

```go
// Menggunakan default 10MB limit
router.Use(body_limit.GetModule().GetFactory()(nil))

// Atau menggunakan convenience function
router.Use(body_limit.BodyLimit10MB())
```

### 2. Konfigurasi dengan Map

```yaml
# config.yaml
middleware:
  - name: "body_limit"
    config:
      max_size: 5242880      # 5MB
      status_code: 413
      message: "Payload terlalu besar"
      skip_on_path:
        - "/upload/*"        # Skip untuk semua upload paths
        - "/webhook"          # Skip untuk webhook endpoint
        - "/api/files/**"     # Skip untuk semua file operations
```

```go
// Dalam kode
config := map[string]any{
    "max_size":     int64(5 * 1024 * 1024), // 5MB
    "status_code":  413,
    "message":      "Request body too large",
    "skip_on_path": []string{"/upload/*", "/webhook"},
}
router.Use(body_limit.GetModule().GetFactory()(config))
```

### 3. Konfigurasi dengan Struct

```go
config := &body_limit.Config{
    MaxSize:    1024 * 1024, // 1MB
    StatusCode: 400,
    Message:    "File terlalu besar",
    SkipOnPath: []string{
        "/api/webhooks/*",     // Skip webhook endpoints
        "/upload/large/**",    // Skip large file uploads
    },
}
router.Use(body_limit.GetModule().GetFactory()(config))
```

## Pattern Matching untuk Skip Paths

Middleware mendukung berbagai pattern untuk `skip_on_path`:

### 1. Exact Match
```yaml
skip_on_path:
  - "/webhook"              # Hanya /webhook
  - "/api/status"           # Hanya /api/status
```

### 2. Single Wildcard (*)
```yaml
skip_on_path:
  - "/upload/*"             # /upload/image, /upload/file (tidak termasuk subdirectory)
  - "/api/*/status"         # /api/v1/status, /api/v2/status
```

### 3. Double Wildcard (**)
```yaml
skip_on_path:
  - "/static/**"            # Semua files dalam /static/ dan subdirectories
  - "/api/**/upload"        # /api/v1/upload, /api/v1/files/upload, dll
```

## Convenience Functions

```go
// Predefined size limits
router.Use(body_limit.BodyLimit1MB())   // 1MB limit
router.Use(body_limit.BodyLimit5MB())   // 5MB limit
router.Use(body_limit.BodyLimit10MB())  // 10MB limit (default)
router.Use(body_limit.BodyLimit50MB())  // 50MB limit (untuk file uploads)

// Custom size
router.Use(body_limit.BodyLimit(2 * 1024 * 1024)) // 2MB

// Skip large payloads instead of error
router.Use(body_limit.BodyLimitWithSkip(1024 * 1024)) // 1MB dengan skip
```

## Contoh Response

### Ketika Limit Terlampaui
```json
{
    "success": false,
    "message": "Request body too large",
    "data": {
        "maxSize": 1048576,
        "actual": 2097152
    }
}
```

### Error Log
```
Request body too large (maxSize: 1048576, actual: 2097152)
```

## Testing

Body limit middleware dilengkapi dengan comprehensive test suite:

```bash
go test ./middleware/body_limit -v
```

Test mencakup:
- Basic body limit enforcement
- Content-Length header checking
- Skip large payloads functionality
- Custom configuration
- Skip path pattern matching
- Factory function parsing
- Error scenarios
