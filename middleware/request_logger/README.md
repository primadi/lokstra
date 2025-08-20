# Request Logger Middleware

Middleware untuk logging request HTTP yang masuk beserta metadata dan optional request/response body logging.

## Features

- ✅ **Request Logging**: Log semua request HTTP yang masuk
- ✅ **Response Time Tracking**: Ukur dan log duration dari request
- ✅ **Request Body Logging**: Optional logging body request (configurable)
- ✅ **Response Body Logging**: Optional logging body response (configurable)
- ✅ **Metadata Logging**: Method, path, query, IP, User-Agent
- ✅ **Status Code Based Logging**: Different log levels untuk error vs success
- ✅ **Body Truncation**: Automatic truncation untuk body yang terlalu panjang
- ✅ **JSON Detection**: Parse dan format JSON body dengan baik

## Configuration

Middleware ini mendukung 2 parameter konfigurasi utama:

```yaml
middleware:
  - name: "request_logger"
    enabled: true
    config:
      include_request_body: false   # Log request body
      include_response_body: false  # Log response body (TODO: belum implementasi)
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `include_request_body` | boolean | `false` | Apakah akan log request body |
| `include_response_body` | boolean | `false` | Apakah akan log response body (belum diimplementasi) |

## Usage Examples

### 1. Basic Request Logging (Default)

```yaml
middleware:
  - name: "request_logger"
    enabled: true
```

Output log:
```json
{
  "level": "info",
  "time": "2025-08-20T10:30:00Z",
  "method": "GET",
  "path": "/api/users",
  "query": "page=1&limit=10",
  "remote_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "message": "Incoming request"
}

{
  "level": "info",
  "time": "2025-08-20T10:30:00.250Z",
  "method": "GET",
  "path": "/api/users",
  "query": "page=1&limit=10",
  "remote_ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "duration": "250ms",
  "duration_ms": 250,
  "status": 200,
  "message": "Request completed successfully"
}
```

### 2. Request Logging dengan Request Body

```yaml
middleware:
  - name: "request_logger"
    enabled: true
    config:
      include_request_body: true
```

Untuk JSON request:
```json
{
  "level": "info",
  "time": "2025-08-20T10:30:00Z",
  "method": "POST",
  "path": "/api/users",
  "query": "",
  "remote_ip": "192.168.1.100",
  "user_agent": "curl/7.68.0",
  "request_body": {
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user"
  },
  "message": "Incoming request"
}
```

Untuk text request:
```json
{
  "level": "info",
  "time": "2025-08-20T10:30:00Z",
  "method": "POST",
  "path": "/api/webhook",
  "request_body": "plain text payload data",
  "message": "Incoming request"
}
```

### 3. Error Response Logging

Untuk response dengan error (4xx, 5xx), level log akan disesuaikan:

```json
{
  "level": "warn",
  "time": "2025-08-20T10:30:00.150Z",
  "method": "GET",
  "path": "/api/users/999",
  "duration": "150ms",
  "duration_ms": 150,
  "status": 404,
  "message": "Request completed with client error"
}

{
  "level": "error",
  "time": "2025-08-20T10:30:00.500Z",
  "method": "POST",
  "path": "/api/users",
  "duration": "500ms",
  "duration_ms": 500,
  "status": 500,
  "message": "Request completed with server error"
}
```

## Log Fields

### Request Log Fields

| Field | Type | Description |
|-------|------|-------------|
| `method` | string | HTTP method (GET, POST, etc.) |
| `path` | string | Request path |
| `query` | string | Query string |
| `remote_ip` | string | Client IP address |
| `user_agent` | string | User-Agent header |
| `request_body` | object/string | Request body (jika enabled) |

### Response Log Fields

| Field | Type | Description |
|-------|------|-------------|
| `duration` | string | Request duration (human readable) |
| `duration_ms` | number | Request duration dalam milliseconds |
| `status` | number | HTTP status code |
| `response_body` | object/string | Response body (jika enabled) |

## Log Levels

Middleware menggunakan log level yang berbeda berdasarkan status code:

- **Info**: Status 200-399 (success responses)
- **Warn**: Status 400-499 (client errors)
- **Error**: Status 500+ (server errors)

## Body Handling

### Request Body

- ✅ Body di-buffer dan di-restore untuk handler selanjutnya
- ✅ JSON body di-parse dan di-log sebagai object
- ✅ Non-JSON body di-log sebagai string
- ✅ Body > 1000 karakter akan di-truncate dengan "... (truncated)"
- ✅ Mendukung semua content types

### Response Body

- ⚠️ **TODO**: Belum diimplementasikan
- 🔄 Memerlukan response writer wrapper untuk capture response
- 📝 Saat ini hanya log debug message bahwa fitur diminta

## Integration dengan Lokstra

### 1. Dalam Konfigurasi YAML

```yaml
apps:
  - name: "my_api"
    middleware:
      - name: "request_logger"
        enabled: true
        config:
          include_request_body: true
          include_response_body: false
```

### 2. Pada Level Group

```yaml
groups:
  - prefix: "/api/v1"
    middleware:
      - name: "request_logger"
        enabled: true
        config:
          include_request_body: true
```

### 3. Global Level

```yaml
server:
  middleware:
    - name: "request_logger"
      enabled: true
```

## Performance Considerations

- **Minimal Overhead**: Tanpa body logging, overhead sangat minimal
- **Request Body Buffering**: Menambah memory usage saat `include_request_body: true`
- **Body Truncation**: Otomatis truncate body > 1000 karakter untuk performa
- **JSON Parsing**: Attempt JSON parse, fallback ke string jika gagal

## Best Practices

### 1. Selective Body Logging

```yaml
# Hanya log body untuk endpoints tertentu
groups:
  - prefix: "/api/v1/auth"
    middleware:
      - name: "request_logger"
        config:
          include_request_body: true  # Log untuk debugging auth
          
  - prefix: "/api/v1/users"
    middleware:
      - name: "request_logger"
        config:
          include_request_body: false # Tidak log untuk performa
```

### 2. Production vs Development

```yaml
# Development
middleware:
  - name: "request_logger"
    config:
      include_request_body: true
      include_response_body: true
      
# Production  
middleware:
  - name: "request_logger"
    config:
      include_request_body: false
      include_response_body: false
```

### 3. Sensitive Data

⚠️ **Warning**: Jangan enable body logging untuk endpoints yang mengandung:
- Passwords
- API keys
- Personal data
- Payment information

## Troubleshooting

### Logger Tidak Tersedia

Jika global logger tidak tersedia (seperti dalam test), middleware akan skip logging dan langsung jalankan handler berikutnya.

### Body Tidak Terbaca

Jika request body sudah terbaca sebelumnya, middleware tetap akan coba baca dan restore untuk handler berikutnya.

### Memory Usage

Untuk aplikasi dengan traffic tinggi dan body besar, pertimbangkan:
- Disable body logging di production
- Implement body size limits
- Monitor memory usage

## Testing

Middleware sudah dilengkapi dengan comprehensive test:

```bash
go test -v ./middleware/request_logger/
```

Test coverage:
- ✅ Config parsing (5 test cases)
- ✅ Module registration
- ✅ Basic logging functionality  
- ✅ Request body logging (JSON & text)
- ✅ Error status logging (200, 400, 404, 500)
- ✅ Long body truncation

## Future Enhancements

- [ ] Response body logging implementation
- [ ] Configurable truncation limits
- [ ] Body size limits per endpoint
- [ ] Selective field logging
- [ ] Request correlation IDs
- [ ] Performance metrics collection
