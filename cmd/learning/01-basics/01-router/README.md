# 01-router: Basic Router

## What You'll Learn
- Create a Lokstra router
- Define HTTP routes (GET, POST, PUT, DELETE)
- Handle path parameters
- Use the router with standard library's `http.ListenAndServe`

## Key Concepts

### Router
A **Router** is the core building block of Lokstra. It:
- Maps HTTP requests to handler functions
- Supports all standard HTTP methods
- Can handle path parameters (`:name`)
- Implements `http.Handler` interface

### Handler Function
```go
func(c *lokstra.RequestContext) error
```
- Takes a `RequestContext` which contains request and response helpers
- Returns an error (nil if successful)
- Use `c.Api` methods for standard responses

## Running the Example

```bash
cd cmd/learning/01-basics/01-router
go run main.go
```

## Testing

```bash
# Simple GET
curl http://localhost:8080/hello

# GET with path parameter
curl http://localhost:8080/hello/John

# POST
curl -X POST http://localhost:8080/greet

# PUT
curl -X PUT http://localhost:8080/update/123

# DELETE
curl -X DELETE http://localhost:8080/delete/123
```

## What's Next?
- **02-app**: Learn how to combine multiple routers into an App
- **04-handlers**: Learn advanced handler patterns (smart bind, manual bind)
