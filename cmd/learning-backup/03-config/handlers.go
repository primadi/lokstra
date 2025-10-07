package main

import (
	"fmt"

	"github.com/primadi/lokstra"
)

// Mock data for demonstration
var usersDB = []map[string]any{
	{"id": 1, "name": "Alice Johnson", "email": "alice@example.com", "age": 28},
	{"id": 2, "name": "Bob Smith", "email": "bob@example.com", "age": 35},
	{"id": 3, "name": "Charlie Brown", "email": "charlie@example.com", "age": 42},
}

// healthCheckHandler returns server health status
func healthCheckHandler(c *lokstra.RequestContext) error {
	return c.Api.Ok(map[string]any{
		"status":  "healthy",
		"service": "config-demo",
		"message": "Server is running with YAML configuration",
	})
}

// listUsersHandler returns list of all users
func listUsersHandler(c *lokstra.RequestContext) error {
	return c.Api.Ok(map[string]any{
		"data":  usersDB,
		"count": len(usersDB),
	})
}

// createUserHandler creates a new user
func createUserHandler(c *lokstra.RequestContext, req *CreateUserRequest) error {
	// Validation happens automatically via Smart Binding!

	// Create new user
	newUser := map[string]any{
		"id":    len(usersDB) + 1,
		"name":  req.Name,
		"email": req.Email,
		"age":   req.Age,
	}

	usersDB = append(usersDB, newUser)

	return c.Api.Created(newUser, "User created successfully")
}

// getUserHandler returns a single user by ID
func getUserHandler(c *lokstra.RequestContext) error {
	id := c.Req.PathParam("id", "0")

	// Find user
	for _, user := range usersDB {
		if userID, ok := user["id"].(int); ok && fmt.Sprintf("%d", userID) == id {
			return c.Api.Ok(user)
		}
	}

	return c.Api.NotFound("User not found")
}

// CreateUserRequest defines the structure for creating a user
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=1,max=120"`
}

// adminStatsHandler returns admin statistics
func adminStatsHandler(c *lokstra.RequestContext) error {
	return c.Api.Ok(map[string]any{
		"total_users": len(usersDB),
		"endpoint":    "/admin/stats",
		"note":        "This endpoint demonstrates route options",
	})
}
