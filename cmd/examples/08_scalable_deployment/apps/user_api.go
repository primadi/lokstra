package apps

import (
	"github.com/primadi/lokstra/core/request"
)

// User API handlers
func GetUsers(ctx *request.Context) error {
	users := []map[string]any{
		{"id": 1, "name": "John Doe", "email": "john@example.com"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
	}

	return ctx.Ok(users)
}

func GetUser(ctx *request.Context) error {
	id := ctx.GetPathParam("id")

	user := map[string]any{
		"id":    id,
		"name":  "John Doe",
		"email": "john@example.com",
	}

	return ctx.Ok(user)
}

type createUserRequest struct {
	Name  string `body:"name"`
	Email string `body:"email"`
}

func CreateUser(ctx *request.Context, req *createUserRequest) error {
	// Simulate user creation
	user := map[string]any{
		"id":    123,
		"name":  req.Name,
		"email": req.Email,
	}

	return ctx.OkCreated(user)
}
