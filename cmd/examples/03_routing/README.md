# 03 - Routing

This section covers Lokstra's powerful routing capabilities and request handling patterns. Learn how to define routes, handle parameters, implement middleware, organize route groups, and apply constraints.

## Examples Overview

### [01 - Basic Routing](./01_basic_routing/)
**Concepts**: Route definition, HTTP methods, path parameters, query parameters
**Key Features**:
- HTTP method handlers (GET, POST, PUT, DELETE)
- Path parameter extraction and validation
- Query parameter handling with defaults
- Request/response patterns
- URL pattern matching

**Why Learn This**: Foundation of web API development - understanding how to map URLs to handlers and extract data from requests.

### [02 - Middleware](./02_middleware/)
**Concepts**: Middleware patterns, execution order, cross-cutting concerns
**Key Features**:
- Global middleware (applies to all routes)
- Group middleware (applies to route groups)
- Route-specific middleware
- Middleware chaining and composition
- Common patterns (auth, logging, CORS, rate limiting)

**Why Learn This**: Essential for implementing cross-cutting concerns like authentication, logging, and request validation in a clean, reusable way.

### [03 - Route Groups](./03_route_groups/)
**Concepts**: Route organization, API versioning, nested groups
**Key Features**:
- Logical route grouping by functionality
- API versioning patterns (/api/v1, /api/v2)
- Nested groups with middleware inheritance
- Access control by group (public, protected, admin)
- Scalable API organization

**Why Learn This**: Critical for building maintainable APIs - learn how to organize routes logically and implement clean API architectures.

### [04 - Route Constraints](./04_route_constraints/)
**Concepts**: Parameter validation, custom constraints, pattern matching
**Key Features**:
- Numeric parameter validation
- Pattern-based constraints (email, UUID, slug)
- Range validation
- Enum constraints
- Query parameter validation
- Multiple constraint composition

**Why Learn This**: Ensures data integrity and provides better error handling by validating inputs at the routing level.

## Learning Path

1. **Start with Basic Routing** - Learn fundamental routing concepts and parameter handling
2. **Add Middleware** - Understand how to implement cross-cutting concerns
3. **Organize with Groups** - Learn to structure APIs for maintainability
4. **Apply Constraints** - Implement robust input validation

## Key Routing Concepts

### Route Definition
```go
app.GET("/users/:id", handler)           // Path parameter
app.POST("/users", handler)              // HTTP method
app.PUT("/users/:id/profile", handler)   // Complex path
```

### Parameter Extraction
```go
userID := ctx.GetPathParam("id")         // Path parameter
page := ctx.GetQueryParam("page")        // Query parameter
token := ctx.GetHeader("Authorization")   // Header
```

### Middleware Application
```go
app.Use(globalMiddleware)                // Global
group.Use(groupMiddleware)               // Group-specific
app.GET("/path", middleware, handler)    // Route-specific
```

### Route Grouping
```go
api := app.Group("/api/v1")              // Create group
api.Use(authMiddleware)                  // Group middleware
users := api.Group("/users")             // Nested group
```

## Best Practices

1. **Consistent URL Patterns**: Use consistent naming and structure
2. **Logical Grouping**: Group related routes together
3. **Middleware Order**: Apply middleware in logical order (auth before business logic)
4. **Parameter Validation**: Validate inputs at the routing level
5. **Error Handling**: Provide clear error messages for invalid requests
6. **Documentation**: Document your API structure and constraints

## Common Patterns

### RESTful API Structure
```
GET    /api/v1/users           # List users
POST   /api/v1/users           # Create user
GET    /api/v1/users/:id       # Get user
PUT    /api/v1/users/:id       # Update user
DELETE /api/v1/users/:id       # Delete user
```

### Middleware Stacking
```
Global → Group → Route-specific → Handler
```

### Access Control Layers
```
/public/*    - No authentication
/api/*       - API key required
/admin/*     - Admin authentication
```

## Next Steps

After completing this section, you'll be ready to explore:
- **04 - HTMX Integration**: Server-side rendering and dynamic content
- **05 - Services**: Dependency injection and service management
- **06 - Built-in Features**: Leveraging Lokstra's built-in services
- **07 - Advanced Patterns**: Complex application architectures
- **08 - Real World**: Production-ready applications

## Testing Your Knowledge

Try building a complete API with:
1. Multiple API versions (v1, v2)
2. Authentication middleware
3. Rate limiting
4. Input validation constraints
5. Proper error handling
6. Comprehensive route organization