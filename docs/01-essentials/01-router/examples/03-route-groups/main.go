package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/primadi/lokstra"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Version string `json:"version"` // Track which API version
}

var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}
var nextID = 3

func main() {
	r := lokstra.NewRouter("api")

	// Root level
	r.GET("/", func() map[string]string {
		return map[string]string{
			"message":  "Welcome to API",
			"versions": "v1, v2",
		}
	})

	// ========================================================================
	// API Version 1 - Basic responses
	// ========================================================================
	v1 := r.AddGroup("/v1")

	v1.GET("/users", func() ([]User, error) {
		// V1: Simple list
		result := make([]User, len(users))
		copy(result, users)
		for i := range result {
			result[i].Version = "v1"
		}
		return result, nil
	})

	v1.GET("/users/{id}", func(req *GetUserRequest) (*User, error) {
		user, err := getUser(req.ID)
		if err != nil {
			return nil, err
		}
		user.Version = "v1"
		return user, nil
	})

	// ========================================================================
	// API Version 2 - Enhanced responses with metadata
	// ========================================================================
	v2 := r.AddGroup("/v2")

	v2.GET("/users", func() (map[string]any, error) {
		// V2: With metadata
		result := make([]User, len(users))
		copy(result, users)
		for i := range result {
			result[i].Version = "v2"
		}

		return map[string]any{
			"data": result,
			"meta": map[string]any{
				"count":   len(result),
				"version": "v2",
			},
		}, nil
	})

	v2.GET("/users/{id}", func(req *GetUserRequest) (map[string]any, error) {
		user, err := getUser(req.ID)
		if err != nil {
			return nil, err
		}
		user.Version = "v2"

		return map[string]any{
			"data": user,
			"meta": map[string]any{
				"version":   "v2",
				"retrieved": time.Now().Format(time.RFC3339),
			},
		}, nil
	})

	// ========================================================================
	// Admin routes - Nested group
	// ========================================================================
	admin := r.AddGroup("/admin")

	admin.GET("/", func() map[string]string {
		return map[string]string{
			"message": "Admin Panel",
			"routes":  "/stats, /config",
		}
	})

	admin.GET("/stats", func() map[string]any {
		return map[string]any{
			"total_users": len(users),
			"next_id":     nextID,
		}
	})

	admin.GET("/config", func() map[string]any {
		return map[string]any{
			"api_version": "2.0",
			"environment": "development",
		}
	})

	// Nested group: /admin/users
	adminUsers := admin.AddGroup("/users")

	adminUsers.GET("", func() ([]User, error) {
		return users, nil
	})

	adminUsers.POST("", createUser)

	adminUsers.DELETE("/{id}", deleteUser)

	// ========================================================================
	// Print routes for visualization
	// ========================================================================
	fmt.Println("\n📋 Registered Routes:")
	fmt.Println(strings.Repeat("=", 50))
	r.PrintRoutes()
	fmt.Println(strings.Repeat("=", 50))

	app := lokstra.NewApp("route-groups", ":3000", r)

	fmt.Println("\n🚀 Server running on http://localhost:3000")
	fmt.Println("\n📖 Try different versions:")
	fmt.Println("   # API v1 (simple)")
	fmt.Println("   curl http://localhost:3000/v1/users")
	fmt.Println("   curl http://localhost:3000/v1/users/1")
	fmt.Println("\n   # API v2 (with metadata)")
	fmt.Println("   curl http://localhost:3000/v2/users")
	fmt.Println("   curl http://localhost:3000/v2/users/1")
	fmt.Println("\n   # Admin routes")
	fmt.Println("   curl http://localhost:3000/admin/stats")
	fmt.Println("   curl http://localhost:3000/admin/users")

	app.Run(30 * time.Second)
}

// ============================================================================
// Handlers
// ============================================================================

type GetUserRequest struct {
	ID int `path:"id"`
}

func getUser(id int) (*User, error) {
	for _, u := range users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

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

type DeleteUserRequest struct {
	ID int `path:"id"`
}

func deleteUser(req *DeleteUserRequest) (map[string]string, error) {
	for i, u := range users {
		if u.ID == req.ID {
			users = append(users[:i], users[i+1:]...)
			return map[string]string{
				"message": fmt.Sprintf("User %d deleted", req.ID),
			}, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}
