# Recovery Middleware Configuration Examples

Berikut adalah contoh-contoh konfigurasi untuk middleware recovery dalam file konfigurasi Lokstra YAML.

## Example 1: Development Configuration (dengan stack trace)

```yaml
server:
  name: "development-server"
  global_setting:
    log_level: "debug"
    log_format: "json"

apps:
  - name: "api-app"
    address: ":8080"
    middleware:
      - name: "recovery"
        config:
          enable_stack_trace: true  # Aktifkan stack trace untuk debugging
      - name: "request_logger"
        config:
          include_request_body: true
          include_response_body: true
    
    routes:
      - method: "GET"
        path: "/health"
        handler: "health.check"
      - method: "GET"
        path: "/test-panic"
        handler: "test.panic"  # Handler yang mungkin panic untuk testing
```

## Example 2: Production Configuration (tanpa stack trace)

```yaml
server:
  name: "production-server"
  global_setting:
    log_level: "error"
    log_format: "json"
    log_output: "stdout"

apps:
  - name: "api-app"
    address: ":8080"
    middleware:
      - name: "recovery"
        config:
          enable_stack_trace: false  # Nonaktifkan stack trace untuk production
      - name: "request_logger"
        config:
          include_request_body: false
          include_response_body: false
      - name: "cors"
        config:
          allowed_origins: ["https://app.example.com"]
          allowed_methods: ["GET", "POST", "PUT", "DELETE"]
          allowed_headers: ["Content-Type", "Authorization"]
    
    routes:
      - method: "GET"
        path: "/api/health"
        handler: "health.check"
```

## Example 3: Environment-based Configuration

```yaml
server:
  name: "{{.ENV}}-server"
  global_setting:
    log_level: "{{.LOG_LEVEL | default \"info\"}}"
    log_format: "json"

apps:
  - name: "main-app"
    address: ":{{.PORT | default \"8080\"}}"
    
    middleware:
      # Recovery middleware dengan conditional stack trace
      - name: "recovery"
        config:
          enable_stack_trace: "{{.ENABLE_STACK_TRACE | default \"true\"}}"
      
      # Request logger dengan conditional body logging
      - name: "request_logger"
        config:
          include_request_body: "{{.LOG_REQUEST_BODY | default \"false\"}}"
          include_response_body: "{{.LOG_RESPONSE_BODY | default \"false\"}}"
    
    groups:
      - prefix: "/api/v1"
        middleware:
          # Override dengan konfigurasi spesifik untuk API routes
          - name: "recovery"
            config:
              enable_stack_trace: false  # Selalu nonaktifkan untuk API endpoints
        routes:
          - method: "GET"
            path: "/users"
            handler: "user.list"
          - method: "POST"
            path: "/users"
            handler: "user.create"
```

## Example 4: Multi-App Configuration

```yaml
server:
  name: "multi-app-server"

apps:
  # Admin App - dengan full debugging
  - name: "admin-app"
    address: ":8080"
    middleware:
      - name: "recovery"
        config:
          enable_stack_trace: true  # Full stack trace untuk admin
      - name: "request_logger"
        config:
          include_request_body: true
          include_response_body: true
    
    routes:
      - method: "GET"
        path: "/admin/dashboard"
        handler: "admin.dashboard"

  # Public API - production ready
  - name: "public-api"
    address: ":8081"
    middleware:
      - name: "recovery"
        config:
          enable_stack_trace: false  # Tanpa stack trace untuk public API
      - name: "request_logger"
        config:
          include_request_body: false
          include_response_body: false
      - name: "cors"
        config:
          allowed_origins: ["*"]
          allowed_methods: ["GET", "POST"]
    
    groups:
      - prefix: "/api"
        routes:
          - method: "GET"
            path: "/status"
            handler: "api.status"
```

## Example 5: Fine-grained Middleware Control

```yaml
server:
  name: "fine-grained-server"

apps:
  - name: "main-app"
    address: ":8080"
    
    # Global middleware dengan default settings
    middleware:
      - name: "recovery"
        config:
          enable_stack_trace: true
    
    groups:
      # Development endpoints - dengan full debugging
      - prefix: "/dev"
        middleware:
          - name: "recovery"
            config:
              enable_stack_trace: true
        routes:
          - method: "GET"
            path: "/test"
            handler: "dev.test"
      
      # Production API endpoints - tanpa stack trace
      - prefix: "/api"
        override_middleware: true  # Override parent middleware
        middleware:
          - name: "recovery"
            config:
              enable_stack_trace: false
          - name: "request_logger"
        routes:
          - method: "GET"
            path: "/users"
            handler: "user.list"
            # Route-specific middleware override
            override_middleware: true
            middleware:
              - name: "recovery"
                config:
                  enable_stack_trace: true  # Enable untuk route ini saja
```

## Environment Variables

Untuk konfigurasi yang fleksibel, gunakan environment variables:

```bash
# Development
export ENABLE_STACK_TRACE=true
export LOG_REQUEST_BODY=true
export LOG_RESPONSE_BODY=true

# Production
export ENABLE_STACK_TRACE=false
export LOG_REQUEST_BODY=false
export LOG_RESPONSE_BODY=false
```

## JSON Schema Validation

Schema JSON Lokstra sekarang sudah mendukung validasi untuk:

- `recovery` middleware dengan properti `enable_stack_trace`
- `request_logger` middleware dengan `include_request_body` dan `include_response_body`
- `cors` middleware dengan konfigurasi CORS lengkap
- `body_limit` middleware dengan konfigurasi size limit

Ini memberikan autocomplete dan validasi yang lebih baik dalam IDE yang mendukung JSON Schema.
