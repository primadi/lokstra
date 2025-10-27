# Putting It Together - Complete REST API

> **Build a production-ready REST API from scratch**  
> **Time**: 45-60 minutes • **Level**: Intermediate • **Project**: Complete Todo API

---

## 🎯 What You'll Build

A **complete Todo API** with:

✅ CRUD operations (Create, Read, Update, Delete)  
✅ Input validation  
✅ Error handling  
✅ Authentication middleware  
✅ YAML configuration  
✅ Service pattern  
✅ Graceful shutdown  
✅ Production-ready structure

**API Endpoints:**

```
POST   /api/todos          - Create todo
GET    /api/todos          - List all todos
GET    /api/todos/:id      - Get single todo
PUT    /api/todos/:id      - Update todo
DELETE /api/todos/:id      - Delete todo
GET    /api/health         - Health check
```

---

## 📁 Project Structure

```
todo-api/
├── main.go                 # Entry point
├── config.yaml             # Configuration
├── handlers/
│   └── todo_handler.go     # HTTP handlers
├── services/
│   └── todo_service.go     # Business logic
├── models/
│   └── todo.go             # Data models
├── middleware/
│   └── auth.go             # Auth middleware
└── go.mod
```

---

## 💻 Step 1: Project Setup

**Initialize project:**

```bash
mkdir todo-api
cd todo-api
go mod init github.com/yourusername/todo-api
go get github.com/primadi/lokstra
```

---

## 💻 Step 2: Define Models

**File: `models/todo.go`**

```go
package models

import (
    "errors"
    "time"
)

// Todo represents a todo item
type Todo struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTodoRequest for creating new todos
type CreateTodoRequest struct {
    Title       string `json:"title"`
    Description string `json:"description"`
}

func (r *CreateTodoRequest) Validate() error {
    if r.Title == "" {
        return errors.New("title is required")
    }
    if len(r.Title) < 3 {
        return errors.New("title must be at least 3 characters")
    }
    if len(r.Title) > 100 {
        return errors.New("title must be less than 100 characters")
    }
    return nil
}

// UpdateTodoRequest for updating todos
type UpdateTodoRequest struct {
    Title       *string `json:"title,omitempty"`
    Description *string `json:"description,omitempty"`
    Completed   *bool   `json:"completed,omitempty"`
}

func (r *UpdateTodoRequest) Validate() error {
    if r.Title != nil {
        if *r.Title == "" {
            return errors.New("title cannot be empty")
        }
        if len(*r.Title) < 3 {
            return errors.New("title must be at least 3 characters")
        }
        if len(*r.Title) > 100 {
            return errors.New("title must be less than 100 characters")
        }
    }
    return nil
}

// ErrorResponse for API errors
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message,omitempty"`
}

// SuccessResponse for API success messages
type SuccessResponse struct {
    Message string `json:"message"`
    Data    any    `json:"data,omitempty"`
}
```

---

## 💻 Step 3: Create Todo Service

**File: `services/todo_service.go`**

```go
package services

import (
    "errors"
    "sync"
    "time"
    
    "github.com/yourusername/todo-api/models"
)

// TodoService manages todo operations
type TodoService struct {
    todos    map[int]*models.Todo
    mu       sync.RWMutex
    nextID   int
}

// NewTodoService creates a new todo service
func NewTodoService() *TodoService {
    return &TodoService{
        todos:  make(map[int]*models.Todo),
        nextID: 1,
    }
}

// Create a new todo
func (s *TodoService) Create(title, description string) (*models.Todo, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    now := time.Now()
    todo := &models.Todo{
        ID:          s.nextID,
        Title:       title,
        Description: description,
        Completed:   false,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
    
    s.todos[s.nextID] = todo
    s.nextID++
    
    return todo, nil
}

// GetAll returns all todos
func (s *TodoService) GetAll() []*models.Todo {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    todos := make([]*models.Todo, 0, len(s.todos))
    for _, todo := range s.todos {
        todos = append(todos, todo)
    }
    
    return todos
}

// GetByID returns a todo by ID
func (s *TodoService) GetByID(id int) (*models.Todo, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    todo, exists := s.todos[id]
    if !exists {
        return nil, errors.New("todo not found")
    }
    
    return todo, nil
}

// Update a todo
func (s *TodoService) Update(id int, req *models.UpdateTodoRequest) (*models.Todo, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    todo, exists := s.todos[id]
    if !exists {
        return nil, errors.New("todo not found")
    }
    
    // Update fields if provided
    if req.Title != nil {
        todo.Title = *req.Title
    }
    if req.Description != nil {
        todo.Description = *req.Description
    }
    if req.Completed != nil {
        todo.Completed = *req.Completed
    }
    
    todo.UpdatedAt = time.Now()
    
    return todo, nil
}

// Delete a todo
func (s *TodoService) Delete(id int) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if _, exists := s.todos[id]; !exists {
        return errors.New("todo not found")
    }
    
    delete(s.todos, id)
    return nil
}

// Stats returns statistics
func (s *TodoService) Stats() map[string]int {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    total := len(s.todos)
    completed := 0
    
    for _, todo := range s.todos {
        if todo.Completed {
            completed++
        }
    }
    
    return map[string]int{
        "total":     total,
        "completed": completed,
        "pending":   total - completed,
    }
}
```

---

## 💻 Step 4: Create Handlers

**File: `handlers/todo_handler.go`**

```go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/yourusername/todo-api/models"
    "github.com/yourusername/todo-api/services"
)

type TodoHandler struct {
    service *services.TodoService
}

func NewTodoHandler(service *services.TodoService) *TodoHandler {
    return &TodoHandler{service: service}
}

// CreateTodo handles POST /api/todos
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
    var req models.CreateTodoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
        return
    }
    
    // Validate request
    if err := req.Validate(); err != nil {
        sendError(w, http.StatusBadRequest, "Validation failed", err.Error())
        return
    }
    
    // Create todo
    todo, err := h.service.Create(req.Title, req.Description)
    if err != nil {
        sendError(w, http.StatusInternalServerError, "Failed to create todo", err.Error())
        return
    }
    
    sendSuccess(w, http.StatusCreated, "Todo created successfully", todo)
}

// GetTodos handles GET /api/todos
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
    todos := h.service.GetAll()
    
    sendSuccess(w, http.StatusOK, "", map[string]any{
        "todos": todos,
        "count": len(todos),
    })
}

// GetTodo handles GET /api/todos/:id
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
    id, err := extractID(r)
    if err != nil {
        sendError(w, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }
    
    todo, err := h.service.GetByID(id)
    if err != nil {
        sendError(w, http.StatusNotFound, "Todo not found", err.Error())
        return
    }
    
    sendSuccess(w, http.StatusOK, "", todo)
}

// UpdateTodo handles PUT /api/todos/:id
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
    id, err := extractID(r)
    if err != nil {
        sendError(w, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }
    
    var req models.UpdateTodoRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        sendError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
        return
    }
    
    // Validate request
    if err := req.Validate(); err != nil {
        sendError(w, http.StatusBadRequest, "Validation failed", err.Error())
        return
    }
    
    // Update todo
    todo, err := h.service.Update(id, &req)
    if err != nil {
        sendError(w, http.StatusNotFound, "Todo not found", err.Error())
        return
    }
    
    sendSuccess(w, http.StatusOK, "Todo updated successfully", todo)
}

// DeleteTodo handles DELETE /api/todos/:id
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
    id, err := extractID(r)
    if err != nil {
        sendError(w, http.StatusBadRequest, "Invalid ID", err.Error())
        return
    }
    
    if err := h.service.Delete(id); err != nil {
        sendError(w, http.StatusNotFound, "Todo not found", err.Error())
        return
    }
    
    sendSuccess(w, http.StatusOK, "Todo deleted successfully", nil)
}

// HealthCheck handles GET /api/health
func (h *TodoHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    stats := h.service.Stats()
    
    sendSuccess(w, http.StatusOK, "", map[string]any{
        "status": "healthy",
        "stats":  stats,
    })
}

// Helper functions

func extractID(r *http.Request) (int, error) {
    // Extract ID from URL path (e.g., /api/todos/123)
    path := r.URL.Path
    // Simple extraction - in production, use proper router params
    var id int
    _, err := fmt.Sscanf(path, "/api/todos/%d", &id)
    return id, err
}

func sendError(w http.ResponseWriter, status int, error, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(models.ErrorResponse{
        Error:   error,
        Message: message,
    })
}

func sendSuccess(w http.ResponseWriter, status int, message string, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(models.SuccessResponse{
        Message: message,
        Data:    data,
    })
}
```

---

## 💻 Step 5: Create Middleware

**File: `middleware/auth.go`**

```go
package middleware

import (
    "encoding/json"
    "net/http"
    "strings"
)

// Simple API key authentication
func APIKeyAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Skip auth for health check
        if r.URL.Path == "/api/health" {
            next.ServeHTTP(w, r)
            return
        }
        
        // Check API key in header
        apiKey := r.Header.Get("X-API-Key")
        if apiKey == "" {
            sendAuthError(w, "API key is required")
            return
        }
        
        // Validate API key (in production, check against database)
        if !isValidAPIKey(apiKey) {
            sendAuthError(w, "Invalid API key")
            return
        }
        
        // Continue to next handler
        next.ServeHTTP(w, r)
    })
}

func isValidAPIKey(key string) bool {
    // Simple validation - in production, check database
    validKeys := []string{
        "dev-key-123",
        "prod-key-456",
    }
    
    for _, validKey := range validKeys {
        if key == validKey {
            return true
        }
    }
    
    return false
}

func sendAuthError(w http.ResponseWriter, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnauthorized)
    json.NewEncoder(w).Encode(map[string]string{
        "error":   "Unauthorized",
        "message": message,
    })
}

// CORS middleware
func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}

// Request logging middleware
func Logger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("[%s] %s %s", r.RemoteAddr, r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

---

## 💻 Step 6: Configuration File

**File: `config.yaml`**

```yaml
# Server configuration
server:
  name: todo-api
  port: ${PORT:8080}
  timeout: 30

# API configuration
api:
  version: v1
  prefix: /api

# Feature flags
features:
  auth_enabled: true
  cors_enabled: true
  logging_enabled: true
```

---

## 💻 Step 7: Main Application

**File: `main.go`**

```go
package main

import (
    "fmt"
    "log"
    "os"
    "time"
    
    "github.com/primadi/lokstra"
    "github.com/yourusername/todo-api/handlers"
    "github.com/yourusername/todo-api/middleware"
    "github.com/yourusername/todo-api/services"
)

func main() {
    // Get port from environment
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    // Create service
    todoService := services.NewTodoService()
    
    // Create handlers
    todoHandler := handlers.NewTodoHandler(todoService)
    
    // Create router
    router := lokstra.NewRouter("api")
    
    // Global middleware
    router.Use(middleware.Logger)
    router.Use(middleware.CORS)
    router.Use(middleware.APIKeyAuth)
    
    // Register routes
    router.Post("/api/todos", todoHandler.CreateTodo)
    router.Get("/api/todos", todoHandler.GetTodos)
    router.Get("/api/todos/:id", todoHandler.GetTodo)
    router.Put("/api/todos/:id", todoHandler.UpdateTodo)
    router.Delete("/api/todos/:id", todoHandler.DeleteTodo)
    router.Get("/api/health", todoHandler.HealthCheck)
    
    // Create and start app
    app := lokstra.NewApp("todo-api", ":"+port, router)
    
    log.Printf("🚀 Todo API starting on port %s", port)
    log.Printf("📋 API endpoints:")
    log.Printf("  POST   /api/todos")
    log.Printf("  GET    /api/todos")
    log.Printf("  GET    /api/todos/:id")
    log.Printf("  PUT    /api/todos/:id")
    log.Printf("  DELETE /api/todos/:id")
    log.Printf("  GET    /api/health")
    log.Println()
    log.Println("🔑 Use header: X-API-Key: dev-key-123")
    log.Println("🛑 Press Ctrl+C to stop")
    
    // Run with graceful shutdown
    if err := app.Run(30 * time.Second); err != nil {
        log.Fatal(err)
    }
}
```

---

## 🧪 Testing the API

### Start the server:

```bash
go run main.go
```

### Test with curl:

**Health check (no auth required):**

```bash
curl http://localhost:8080/api/health
```

**Create todo:**

```bash
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-123" \
  -d '{
    "title": "Learn Lokstra",
    "description": "Complete all essentials tutorials"
  }'
```

**Get all todos:**

```bash
curl http://localhost:8080/api/todos \
  -H "X-API-Key: dev-key-123"
```

**Get single todo:**

```bash
curl http://localhost:8080/api/todos/1 \
  -H "X-API-Key: dev-key-123"
```

**Update todo:**

```bash
curl -X PUT http://localhost:8080/api/todos/1 \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-key-123" \
  -d '{
    "completed": true
  }'
```

**Delete todo:**

```bash
curl -X DELETE http://localhost:8080/api/todos/1 \
  -H "X-API-Key: dev-key-123"
```

**Test without API key (should fail):**

```bash
curl http://localhost:8080/api/todos
# Response: {"error":"Unauthorized","message":"API key is required"}
```

---

## 🎯 What We Built

### ✅ Production Features

1. **CRUD Operations** - Complete Create, Read, Update, Delete
2. **Input Validation** - Request validation with clear error messages
3. **Error Handling** - Consistent error responses
4. **Authentication** - API key middleware
5. **CORS Support** - Cross-origin resource sharing
6. **Logging** - Request logging middleware
7. **Service Layer** - Business logic separated from handlers
8. **Clean Structure** - Organized by concern
9. **Graceful Shutdown** - Handles SIGINT/SIGTERM
10. **Health Check** - Monitoring endpoint

### ✅ Best Practices

- ✅ **Separation of Concerns** - Models, Services, Handlers separate
- ✅ **Dependency Injection** - Service injected into handlers
- ✅ **Middleware Chain** - Logging → CORS → Auth
- ✅ **Error Handling** - Consistent error responses
- ✅ **Validation** - Input validation at request level
- ✅ **Thread Safety** - Service uses sync.RWMutex
- ✅ **Clean API** - RESTful endpoints
- ✅ **Environment Config** - Port from environment variable

---

## 🚀 Next Steps

### Enhancements to Try:

1. **Database Integration**
   ```go
   // Replace in-memory storage with PostgreSQL
   todoService := services.NewTodoServiceWithDB(db)
   ```

2. **JWT Authentication**
   ```go
   // Replace API key with JWT tokens
   router.Use(middleware.JWTAuth)
   ```

3. **Rate Limiting**
   ```go
   // Add rate limiting middleware
   router.Use(middleware.RateLimit(100, time.Minute))
   ```

4. **Pagination**
   ```go
   // Add pagination to GetAll
   func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
       page := r.URL.Query().Get("page")
       limit := r.URL.Query().Get("limit")
       todos := h.service.GetPaginated(page, limit)
       // ...
   }
   ```

5. **Search & Filter**
   ```go
   // Add search capability
   GET /api/todos?search=learn&completed=false
   ```

6. **Testing**
   ```go
   // Add unit tests
   func TestCreateTodo(t *testing.T) {
       service := services.NewTodoService()
       todo, err := service.Create("Test", "Description")
       assert.NoError(t, err)
       assert.Equal(t, "Test", todo.Title)
   }
   ```

7. **Docker Deployment**
   ```dockerfile
   FROM golang:1.21-alpine
   WORKDIR /app
   COPY . .
   RUN go build -o todo-api
   CMD ["./todo-api"]
   ```

---

## 📚 What You Learned

### Core Concepts:
- ✅ Router setup and route registration
- ✅ Service pattern for business logic
- ✅ Middleware for cross-cutting concerns
- ✅ Input validation and error handling
- ✅ Clean project structure
- ✅ Graceful shutdown

### Lokstra Features:
- ✅ `lokstra.NewRouter()` - Create routers
- ✅ `router.Use()` - Apply middleware
- ✅ `router.Get/Post/Put/Delete()` - Register routes
- ✅ `lokstra.NewApp()` - Create application
- ✅ `app.Run()` - Start with graceful shutdown

### Production Patterns:
- ✅ Separation of concerns (MVC-like)
- ✅ Dependency injection
- ✅ Middleware composition
- ✅ Error handling strategy
- ✅ API authentication
- ✅ CORS configuration

---

## 🎓 Congratulations!

You've completed the **Essentials** section and built a **production-ready REST API**!

### You now know how to:
- ✅ Build complete REST APIs with Lokstra
- ✅ Structure applications properly
- ✅ Implement authentication and authorization
- ✅ Handle errors gracefully
- ✅ Validate input data
- ✅ Use middleware effectively
- ✅ Deploy production-ready applications

---

## 🚀 Continue Learning

**Explore more:**

- [Deep Dive](../../02-deep-dive/README.md) - Advanced patterns and features
- [Complete Examples](../../05-examples/README.md) - Real-world applications
- [API Reference](../../03-api-reference/README.md) - Detailed documentation
- [Guides](../../04-guides/README.md) - Specific implementation patterns

**Build something amazing!** 🚀
