# Advanced Features

This guide covers advanced Lokstra features including testing strategies, deployment patterns, performance optimization, monitoring, and production best practices. These topics help you build enterprise-ready applications with Lokstra.

## Testing

### Unit Testing

#### Handler Testing

Test your handlers in isolation:

```go
package handlers

import (
    "encoding/json"
    "net/http/httptest"
    "strings"
    "testing"
    
    "github.com/primadi/lokstra"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateUserHandler(t *testing.T) {
    // Setup
    regCtx := lokstra.NewRegistrationContext()
    setupTestServices(regCtx)
    
    // Create test request
    requestBody := `{"name": "John Doe", "email": "john@example.com"}`
    req := httptest.NewRequest("POST", "/users", strings.NewReader(requestBody))
    req.Header.Set("Content-Type", "application/json")
    
    // Create test context
    w := httptest.NewRecorder()
    ctx := lokstra.NewContext(req, w, regCtx)
    
    // Execute handler
    err := createUserHandler(ctx)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, 201, w.Code)
    
    var response map[string]any
    err = json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    
    assert.True(t, response["success"].(bool))
    assert.Equal(t, "User created successfully", response["message"])
}

func setupTestServices(regCtx lokstra.RegistrationContext) {
    // Register mock services for testing
    regCtx.RegisterService("db.main", &mockDbPool{})
    regCtx.RegisterService("logger", &mockLogger{})
}
```

#### Service Testing

Test services with mock dependencies:

```go
func TestUserService(t *testing.T) {
    // Setup mock database
    mockDB := &mockDbPool{
        users: map[string]*User{},
    }
    
    service := &UserService{
        db: mockDB,
    }
    
    // Test user creation
    user, err := service.CreateUser(context.Background(), "John Doe", "john@example.com")
    
    require.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
    assert.Equal(t, "john@example.com", user.Email)
    assert.NotEmpty(t, user.ID)
    
    // Test user retrieval
    retrievedUser, err := service.GetUserByID(context.Background(), user.ID)
    
    require.NoError(t, err)
    assert.Equal(t, user.ID, retrievedUser.ID)
    assert.Equal(t, user.Name, retrievedUser.Name)
}

type mockDbPool struct {
    users map[string]*User
}

func (m *mockDbPool) CreateUser(ctx context.Context, name, email string) (*User, error) {
    user := &User{
        ID:    generateID(),
        Name:  name,
        Email: email,
    }
    m.users[user.ID] = user
    return user, nil
}
```

### Integration Testing

#### Full Application Testing

Test complete request flows:

```go
func TestUserAPI(t *testing.T) {
    // Setup test server
    regCtx := lokstra.NewRegistrationContext()
    setupProductionServices(regCtx)
    
    app := lokstra.NewApp(regCtx, "test-app", ":0")
    setupRoutes(app)
    
    // Create test server
    server := httptest.NewServer(app)
    defer server.Close()
    
    // Test user creation
    userPayload := `{"name": "Jane Doe", "email": "jane@example.com"}`
    resp, err := http.Post(
        server.URL+"/api/users",
        "application/json",
        strings.NewReader(userPayload),
    )
    
    require.NoError(t, err)
    assert.Equal(t, 201, resp.StatusCode)
    
    // Parse response
    var createResponse map[string]any
    err = json.NewDecoder(resp.Body).Decode(&createResponse)
    require.NoError(t, err)
    
    userID := createResponse["data"].(map[string]any)["id"].(string)
    
    // Test user retrieval
    resp, err = http.Get(server.URL + "/api/users/" + userID)
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode)
    
    var getResponse map[string]any
    err = json.NewDecoder(resp.Body).Decode(&getResponse)
    require.NoError(t, err)
    
    userData := getResponse["data"].(map[string]any)
    assert.Equal(t, "Jane Doe", userData["name"])
    assert.Equal(t, "jane@example.com", userData["email"])
}
```

#### Database Integration Testing

Test with real database:

```go
func TestUserServiceWithDB(t *testing.T) {
    // Skip if no test database
    if testing.Short() {
        t.Skip("Skipping database integration test")
    }
    
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    service := &UserService{db: db}
    
    // Test with real database operations
    user, err := service.CreateUser(context.Background(), "Test User", "test@example.com")
    require.NoError(t, err)
    
    // Verify in database
    retrievedUser, err := service.GetUserByID(context.Background(), user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Name, retrievedUser.Name)
}

func setupTestDB(t *testing.T) serviceapi.DbPool {
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        dsn = "postgres://test:test@localhost/test_db?sslmode=disable"
    }
    
    db, err := dbpool_pg.NewPgxPostgresPool(context.Background(), dsn)
    require.NoError(t, err)
    
    // Run migrations
    runTestMigrations(t, db)
    
    return db
}

func cleanupTestDB(t *testing.T, db serviceapi.DbPool) {
    conn, err := db.Acquire(context.Background(), "public")
    require.NoError(t, err)
    defer conn.Release()
    
    _, err = conn.Exec(context.Background(), "TRUNCATE users CASCADE")
    require.NoError(t, err)
}
```

### Testing Middleware

```go
func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name           string
        token          string
        expectedStatus int
        shouldCallNext bool
    }{
        {
            name:           "valid token",
            token:          "Bearer valid-token",
            expectedStatus: 200,
            shouldCallNext: true,
        },
        {
            name:           "invalid token",
            token:          "Bearer invalid-token",
            expectedStatus: 401,
            shouldCallNext: false,
        },
        {
            name:           "missing token",
            token:          "",
            expectedStatus: 401,
            shouldCallNext: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            nextCalled := false
            nextHandler := func(ctx *lokstra.Context) error {
                nextCalled = true
                return ctx.Ok("success")
            }
            
            middleware := authMiddleware(nextHandler)
            
            // Create test request
            req := httptest.NewRequest("GET", "/protected", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", tt.token)
            }
            
            w := httptest.NewRecorder()
            ctx := lokstra.NewContext(req, w, lokstra.NewRegistrationContext())
            
            // Execute
            err := middleware(ctx)
            
            // Assertions
            if tt.expectedStatus >= 400 {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
            
            assert.Equal(t, tt.shouldCallNext, nextCalled)
            assert.Equal(t, tt.expectedStatus, w.Code)
        })
    }
}
```

### Test Utilities

Create reusable test utilities:

```go
package testutil

import (
    "context"
    "testing"
    
    "github.com/primadi/lokstra"
    "github.com/stretchr/testify/require"
)

// TestApp creates a configured test application
func TestApp(t *testing.T) *lokstra.App {
    regCtx := lokstra.NewRegistrationContext()
    
    // Register test services
    RegisterTestServices(regCtx)
    
    app := lokstra.NewApp(regCtx, "test-app", ":0")
    return app
}

// RegisterTestServices sets up services for testing
func RegisterTestServices(regCtx lokstra.RegistrationContext) {
    regCtx.RegisterService("db.main", &MockDbPool{})
    regCtx.RegisterService("logger", &MockLogger{})
    regCtx.RegisterService("cache", &MockCache{})
}

// MockDbPool provides a mock database for testing
type MockDbPool struct {
    data map[string]map[string]any
}

func (m *MockDbPool) Acquire(ctx context.Context, schema string) (serviceapi.DbConn, error) {
    return &MockDbConn{pool: m}, nil
}

// AssertJSONResponse checks JSON response structure
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedData map[string]any) {
    require.Equal(t, expectedStatus, w.Code)
    require.Equal(t, "application/json", w.Header().Get("Content-Type"))
    
    var response map[string]any
    err := json.Unmarshal(w.Body.Bytes(), &response)
    require.NoError(t, err)
    
    for key, value := range expectedData {
        assert.Equal(t, value, response[key])
    }
}
```

## Deployment

### Docker Deployment

#### Basic Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary and static files
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/static ./static

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -S -D -H -u 1001 -h /root -s /sbin/nologin -G appuser -g appuser appuser

USER appuser

EXPOSE 8080
CMD ["./main"]
```

#### Multi-stage with Assets

```dockerfile
# Node.js build stage for frontend assets
FROM node:18-alpine AS frontend

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY frontend/ ./
RUN npm run build

# Go build stage
FROM golang:1.21-alpine AS backend

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy backend binary
COPY --from=backend /app/main .
COPY --from=backend /app/configs ./configs

# Copy frontend assets
COPY --from=frontend /app/dist ./static

EXPOSE 8080
CMD ["./main"]
```

#### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=myapp
      - POSTGRES_USER=appuser
      - POSTGRES_PASSWORD=secret
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Deployment

#### Basic Deployment

```yaml
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lokstra-app
  labels:
    app: lokstra-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: lokstra-app
  template:
    metadata:
      labels:
        app: lokstra-app
    spec:
      containers:
      - name: app
        image: myregistry/lokstra-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: db-host
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: db-password
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: lokstra-app-service
spec:
  selector:
    app: lokstra-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

#### Ingress Configuration

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: lokstra-app-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - api.myapp.com
    secretName: api-tls
  rules:
  - host: api.myapp.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: lokstra-app-service
            port:
              number: 80
```

### Cloud Deployment

#### AWS ECS

```json
{
  "family": "lokstra-app",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "ACCOUNT.dkr.ecr.REGION.amazonaws.com/lokstra-app:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENV",
          "value": "production"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT:secret:db-password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/lokstra-app",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3
      }
    }
  ]
}
```

#### Google Cloud Run

```yaml
# cloudrun.yaml
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: lokstra-app
  annotations:
    run.googleapis.com/ingress: all
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/maxScale: "100"
        run.googleapis.com/cpu-throttling: "false"
    spec:
      containerConcurrency: 1000
      timeoutSeconds: 300
      containers:
      - image: gcr.io/PROJECT-ID/lokstra-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: production
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: db-host
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## Performance Optimization

### Application Optimization

#### Connection Pooling

```go
func optimizeDatabase(regCtx lokstra.RegistrationContext) {
    // Optimize database connections
    dbConfig := map[string]any{
        "host":            os.Getenv("DB_HOST"),
        "max_connections": 50,          // Increase for high load
        "min_connections": 10,          // Maintain minimum connections
        "max_idle_time":   "30m",       // Close idle connections
        "max_lifetime":    "1h",        // Rotate connections
        "connect_timeout": "10s",       // Connection timeout
    }
    
    regCtx.CreateService("lokstra.dbpool_pg", "db.main", dbConfig)
}
```

#### Caching Strategy

```go
func setupCaching(app *lokstra.App) {
    // Redis cache for session data
    app.RegisterService("cache.session", "lokstra.redis", map[string]any{
        "host":        os.Getenv("REDIS_HOST"),
        "max_idle":    100,
        "max_active":  1000,
        "idle_timeout": "5m",
    })
    
    // In-memory cache for frequently accessed data
    app.RegisterService("cache.hot", "lokstra.kvstore_mem")
    
    // Cache middleware
    app.Use(cacheMiddleware())
}

func cacheMiddleware() lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            // Check cache for GET requests
            if ctx.GetMethod() == "GET" {
                cacheKey := generateCacheKey(ctx.GetPath(), ctx.GetQuery())
                
                cache, _ := lokstra.GetService[serviceapi.KvStore](ctx.RegistrationContext, "cache.hot")
                if data, err := cache.Get(ctx.Context, cacheKey); err == nil {
                    ctx.SetHeader("X-Cache", "HIT")
                    return ctx.Ok(data)
                }
            }
            
            // Continue to handler
            err := next(ctx)
            
            // Cache successful GET responses
            if err == nil && ctx.GetMethod() == "GET" && ctx.Response.StatusCode == 200 {
                cacheKey := generateCacheKey(ctx.GetPath(), ctx.GetQuery())
                cache, _ := lokstra.GetService[serviceapi.KvStore](ctx.RegistrationContext, "cache.hot")
                cache.Set(ctx.Context, cacheKey, ctx.Response.Data, time.Hour)
                ctx.SetHeader("X-Cache", "MISS")
            }
            
            return err
        }
    }
}
```

#### Memory Management

```go
func optimizeMemory(server *lokstra.Server) {
    // Set garbage collection target
    debug.SetGCPercent(100)
    
    // Limit memory usage
    debug.SetMemoryLimit(1 << 30) // 1GB
    
    // Monitor memory usage
    go func() {
        ticker := time.NewTicker(time.Minute)
        defer ticker.Stop()
        
        for range ticker.C {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            
            if m.Alloc > 800*1024*1024 { // 800MB
                runtime.GC()
                log.Printf("Forced GC: Alloc=%d KB", m.Alloc/1024)
            }
        }
    }()
}
```

### HTTP Optimization

#### Compression Middleware

```go
func setupCompression(app *lokstra.App) {
    app.Use(compressionMiddleware())
}

func compressionMiddleware() lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            // Check if client accepts compression
            acceptEncoding := ctx.GetHeader("Accept-Encoding")
            if !strings.Contains(acceptEncoding, "gzip") {
                return next(ctx)
            }
            
            // Set compression headers
            ctx.SetHeader("Content-Encoding", "gzip")
            ctx.SetHeader("Vary", "Accept-Encoding")
            
            // Wrap response writer with gzip
            gzipWriter := gzip.NewWriter(ctx.Response.Writer)
            defer gzipWriter.Close()
            
            originalWriter := ctx.Response.Writer
            ctx.Response.Writer = gzipWriter
            
            err := next(ctx)
            
            ctx.Response.Writer = originalWriter
            return err
        }
    }
}
```

#### Request Size Limits

```go
func setupRequestLimits(app *lokstra.App) {
    app.Use(requestSizeLimitMiddleware(10 * 1024 * 1024)) // 10MB limit
}

func requestSizeLimitMiddleware(maxSize int64) lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            if ctx.Request.ContentLength > maxSize {
                return ctx.ErrorRequestEntityTooLarge("Request too large")
            }
            
            // Limit reader
            ctx.Request.Body = http.MaxBytesReader(ctx.Response.Writer, ctx.Request.Body, maxSize)
            
            return next(ctx)
        }
    }
}
```

## Monitoring and Observability

### Health Checks

```go
func setupHealthChecks(app *lokstra.App) {
    // Basic health check
    app.GET("/health", func(ctx *lokstra.Context) error {
        return ctx.Ok(map[string]string{
            "status": "healthy",
            "time":   time.Now().UTC().Format(time.RFC3339),
        })
    })
    
    // Detailed readiness check
    app.GET("/ready", func(ctx *lokstra.Context) error {
        checks := map[string]string{}
        
        // Check database
        if db, err := lokstra.GetService[serviceapi.DbPool](ctx.RegistrationContext, "db.main"); err == nil {
            if conn, err := db.Acquire(ctx.Context, "public"); err == nil {
                conn.Release()
                checks["database"] = "healthy"
            } else {
                checks["database"] = "unhealthy: " + err.Error()
            }
        }
        
        // Check Redis
        if redis, err := lokstra.GetService[serviceapi.Redis](ctx.RegistrationContext, "cache"); err == nil {
            if err := redis.Ping(ctx.Context); err == nil {
                checks["redis"] = "healthy"
            } else {
                checks["redis"] = "unhealthy: " + err.Error()
            }
        }
        
        // Overall status
        allHealthy := true
        for _, status := range checks {
            if !strings.Contains(status, "healthy") {
                allHealthy = false
                break
            }
        }
        
        if allHealthy {
            return ctx.Ok(map[string]any{
                "status": "ready",
                "checks": checks,
            })
        } else {
            return ctx.ErrorServiceUnavailable(map[string]any{
                "status": "not ready",
                "checks": checks,
            })
        }
    })
}
```

### Metrics Collection

```go
func setupMetrics(app *lokstra.App) {
    // Prometheus metrics
    app.GET("/metrics", promhttp.Handler())
    
    // Custom metrics middleware
    app.Use(metricsMiddleware())
}

func metricsMiddleware() lokstra.MiddlewareFunc {
    requestDuration := prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
        },
        []string{"method", "path", "status"},
    )
    
    requestCount := prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    prometheus.MustRegister(requestDuration, requestCount)
    
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            start := time.Now()
            
            err := next(ctx)
            
            duration := time.Since(start).Seconds()
            status := strconv.Itoa(ctx.Response.StatusCode)
            
            requestDuration.WithLabelValues(
                ctx.GetMethod(),
                ctx.GetPath(),
                status,
            ).Observe(duration)
            
            requestCount.WithLabelValues(
                ctx.GetMethod(),
                ctx.GetPath(),
                status,
            ).Inc()
            
            return err
        }
    }
}
```

### Structured Logging

```go
func setupLogging(regCtx lokstra.RegistrationContext) {
    // Configure structured logger
    loggerConfig := map[string]any{
        "level":  os.Getenv("LOG_LEVEL"),
        "format": "json",
        "output": "stdout",
    }
    
    regCtx.CreateService("lokstra.logger", "logger", loggerConfig)
    
    // Request logging middleware
    regCtx.RegisterMiddlewareFunc("request-logger", requestLoggingMiddleware())
}

func requestLoggingMiddleware() lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            start := time.Now()
            requestID := generateRequestID()
            
            // Add request ID to context
            ctx.Set("request_id", requestID)
            ctx.SetHeader("X-Request-ID", requestID)
            
            logger, _ := lokstra.GetService[serviceapi.Logger](ctx.RegistrationContext, "logger")
            
            // Log request start
            logger.Info("Request started",
                "request_id", requestID,
                "method", ctx.GetMethod(),
                "path", ctx.GetPath(),
                "user_agent", ctx.GetHeader("User-Agent"),
                "remote_addr", ctx.Request.RemoteAddr,
            )
            
            err := next(ctx)
            
            duration := time.Since(start)
            
            // Log request completion
            logLevel := "info"
            if ctx.Response.StatusCode >= 400 {
                logLevel = "warn"
            }
            if ctx.Response.StatusCode >= 500 {
                logLevel = "error"
            }
            
            logger.Log(logLevel, "Request completed",
                "request_id", requestID,
                "method", ctx.GetMethod(),
                "path", ctx.GetPath(),
                "status", ctx.Response.StatusCode,
                "duration_ms", duration.Milliseconds(),
                "response_size", len(ctx.Response.Body),
            )
            
            return err
        }
    }
}
```

### Distributed Tracing

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func setupTracing(app *lokstra.App) {
    app.Use(tracingMiddleware())
}

func tracingMiddleware() lokstra.MiddlewareFunc {
    tracer := otel.Tracer("lokstra-app")
    
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            spanCtx, span := tracer.Start(
                ctx.Context,
                fmt.Sprintf("%s %s", ctx.GetMethod(), ctx.GetPath()),
                trace.WithAttributes(
                    attribute.String("http.method", ctx.GetMethod()),
                    attribute.String("http.url", ctx.Request.URL.String()),
                    attribute.String("user_agent", ctx.GetHeader("User-Agent")),
                ),
            )
            defer span.End()
            
            // Update context with span
            ctx.Context = spanCtx
            
            err := next(ctx)
            
            // Record span status
            if err != nil {
                span.RecordError(err)
                span.SetStatus(codes.Error, err.Error())
            } else {
                span.SetStatus(codes.Ok, "")
            }
            
            span.SetAttributes(
                attribute.Int("http.status_code", ctx.Response.StatusCode),
            )
            
            return err
        }
    }
}
```

## Security

### HTTPS and TLS

```go
func setupTLS(server *lokstra.Server) {
    // TLS configuration
    tlsConfig := &tls.Config{
        MinVersion: tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
        PreferServerCipherSuites: true,
    }
    
    // Load certificates
    cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
    if err != nil {
        log.Fatal("Failed to load TLS certificates:", err)
    }
    
    tlsConfig.Certificates = []tls.Certificate{cert}
    
    // Configure secure listener
    listener, err := tls.Listen("tcp", ":443", tlsConfig)
    if err != nil {
        log.Fatal("Failed to create TLS listener:", err)
    }
    
    server.StartWithListener(listener)
}
```

### Security Headers

```go
func securityHeadersMiddleware() lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            // Security headers
            ctx.SetHeader("X-Content-Type-Options", "nosniff")
            ctx.SetHeader("X-Frame-Options", "DENY")
            ctx.SetHeader("X-XSS-Protection", "1; mode=block")
            ctx.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
            ctx.SetHeader("Content-Security-Policy", "default-src 'self'")
            ctx.SetHeader("Referrer-Policy", "strict-origin-when-cross-origin")
            
            return next(ctx)
        }
    }
}
```

### Rate Limiting

```go
func rateLimitMiddleware() lokstra.MiddlewareFunc {
    limiter := rate.NewLimiter(rate.Limit(100), 200) // 100 req/sec, burst 200
    
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            if !limiter.Allow() {
                return ctx.ErrorTooManyRequests("Rate limit exceeded")
            }
            
            return next(ctx)
        }
    }
}
```

## Production Best Practices

### Graceful Shutdown

```go
func main() {
    regCtx := lokstra.NewGlobalRegistrationContext()
    setupServices(regCtx)
    
    server := lokstra.NewServer(regCtx, "production-server")
    setupApps(server)
    
    // Graceful shutdown with timeout
    server.StartAndWaitForShutdown(30 * time.Second)
}
```

### Configuration Management

```go
func loadConfiguration() *Config {
    env := os.Getenv("ENV")
    if env == "" {
        env = "development"
    }
    
    cfg, err := lokstra.LoadConfigDir(fmt.Sprintf("configs/%s", env))
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }
    
    // Validate configuration
    if err := validateConfig(cfg); err != nil {
        log.Fatal("Invalid configuration:", err)
    }
    
    return cfg
}

func validateConfig(cfg *lokstra.LokstraConfig) error {
    required := []string{"DB_HOST", "DB_PASSWORD", "JWT_SECRET"}
    
    for _, key := range required {
        if os.Getenv(key) == "" {
            return fmt.Errorf("required environment variable %s is not set", key)
        }
    }
    
    return nil
}
```

### Error Handling

```go
func setupErrorHandling(app *lokstra.App) {
    // Global error handler
    app.Use(errorHandlingMiddleware())
    
    // Panic recovery
    app.Use(recoveryMiddleware())
}

func errorHandlingMiddleware() lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            err := next(ctx)
            
            if err != nil {
                logger, _ := lokstra.GetService[serviceapi.Logger](ctx.RegistrationContext, "logger")
                
                // Log error with context
                logger.Error("Request error",
                    "error", err.Error(),
                    "path", ctx.GetPath(),
                    "method", ctx.GetMethod(),
                    "request_id", ctx.Get("request_id"),
                )
                
                // Don't expose internal errors to clients
                if ctx.Response.StatusCode == 0 {
                    return ctx.ErrorInternal("Internal server error")
                }
            }
            
            return err
        }
    }
}

func recoveryMiddleware() lokstra.MiddlewareFunc {
    return func(next lokstra.HandlerFunc) lokstra.HandlerFunc {
        return func(ctx *lokstra.Context) error {
            defer func() {
                if r := recover(); r != nil {
                    logger, _ := lokstra.GetService[serviceapi.Logger](ctx.RegistrationContext, "logger")
                    
                    logger.Error("Panic recovered",
                        "panic", r,
                        "stack", string(debug.Stack()),
                        "path", ctx.GetPath(),
                        "method", ctx.GetMethod(),
                    )
                    
                    ctx.ErrorInternal("Internal server error")
                }
            }()
            
            return next(ctx)
        }
    }
}
```

This comprehensive documentation covers the advanced features of Lokstra, providing you with the tools and knowledge needed to build production-ready applications. These patterns and practices will help you create robust, scalable, and maintainable systems with the Lokstra framework.

## Next Steps

- [Core Concepts](./core-concepts.md) - Review fundamental concepts
- [Configuration](./configuration.md) - Master configuration patterns
- [Services](./services.md) - Advanced service management
- [HTMX Integration](./htmx-integration.md) - Build modern web interfaces

---

*Master these advanced features to build enterprise-grade applications with Lokstra. Focus on testing, monitoring, and security from the beginning of your development process.*