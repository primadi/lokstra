package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
)

// Simple data model
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// In-memory storage
var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}
var nextID = 3

func main() {
	// Create router
	r := lokstra.NewRouter("api")

	// GET - Simple string response
	r.GET("/ping", func() string {
		return "pong"
	})

	// GET - Return slice (auto JSON)
	r.GET("/users", func() ([]User, error) {
		return users, nil
	})

	// POST - Create user
	r.POST("/users", createUser)

	// Start server
	app := lokstra.NewApp("basic-routes", ":3000", r)

	fmt.Println("🚀 Server running on http://localhost:3000")
	fmt.Println("📖 Try:")
	fmt.Println("   curl http://localhost:3000/ping")
	fmt.Println("   curl http://localhost:3000/users")
	fmt.Println("   curl -X POST http://localhost:3000/users -H 'Content-Type: application/json' -d '{\"name\":\"Charlie\",\"email\":\"charlie@example.com\"}'")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

// Request type with validation
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func createUser(req *CreateUserRequest) (*User, error) {
	user := User{
		ID:    nextID,
		Name:  req.Name,
		Email: req.Email,
	}
	nextID++
	users = append(users, user)

	return &user, nil
}
