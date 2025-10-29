package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com"},
	{ID: 2, Name: "Bob", Email: "bob@example.com"},
}

func main() {
	r := lokstra.NewRouter("api")

	// ========================================================================
	// FORM 1: Simple Return Value
	// Use when: No parameters, no errors
	// ========================================================================
	r.GET("/ping", func() string {
		return "pong"
	})

	r.GET("/time", func() map[string]string {
		return map[string]string{
			"current_time": time.Now().Format(time.RFC3339),
		}
	})

	// ========================================================================
	// FORM 2: Return with Error (Most Common!)
	// Use when: Operations that can fail
	// ========================================================================
	r.GET("/users", func() ([]User, error) {
		// Simulated database query
		return users, nil
	})

	r.GET("/users-might-fail", func() ([]User, error) {
		// Simulate error
		if len(users) == 0 {
			return nil, fmt.Errorf("no users found")
		}
		return users, nil
	})

	// ========================================================================
	// FORM 3: Request Binding with Error
	// Use when: Need request data (POST/PUT)
	// ========================================================================
	type CreateUserRequest struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	r.POST("/users", func(req *CreateUserRequest) (*User, error) {
		// req is auto-bound from JSON body and validated
		user := &User{
			ID:    len(users) + 1,
			Name:  req.Name,
			Email: req.Email,
		}
		users = append(users, *user)
		return user, nil
	})

	type GetUserRequest struct {
		ID int `path:"id"`
	}

	r.GET("/users/{id}", func(req *GetUserRequest) (*User, error) {
		for _, u := range users {
			if u.ID == req.ID {
				return &u, nil
			}
		}
		return nil, fmt.Errorf("user not found")
	})

	// ========================================================================
	// FORM 4: Context + Request with Error
	// Use when: Need full control (headers, status codes, etc)
	// ========================================================================
	r.GET("/users/{id}/details", func(ctx *request.Context, req *GetUserRequest) (*User, error) {
		// Access request context
		userAgent := ctx.R.Header.Get("User-Agent")
		fmt.Printf("Request from: %s\n", userAgent)

		// Still have request binding
		for _, u := range users {
			if u.ID == req.ID {
				return &u, nil
			}
		}
		return nil, fmt.Errorf("user not found")
	})

	// ========================================================================
	// FORM 5: Full Custom Response
	// Use when: Need complete control over response
	// ========================================================================
	r.GET("/users/{id}/custom", func(ctx *request.Context, req *GetUserRequest) (*response.Response, error) {
		for _, u := range users {
			if u.ID == req.ID {
				// Custom response with headers and status
				resp := response.NewResponse()
				resp.RespHeaders = map[string][]string{
					"X-User-ID":       {fmt.Sprintf("%d", u.ID)},
					"X-Response-Time": {time.Now().Format(time.RFC3339)},
				}
				resp.WithStatus(200).Json(u)
				return resp, nil
			}
		}

		// Custom 404 response
		resp := response.NewResponse()
		resp.RespHeaders = map[string][]string{
			"X-Error-Code": {"USR404"},
		}
		resp.WithStatus(404).Json(map[string]any{
			"error": map[string]string{
				"code":    "USER_NOT_FOUND",
				"message": "User does not exist",
			},
		})
		return resp, nil
	})

	// ========================================================================
	// COMPARISON ENDPOINTS
	// ========================================================================

	// Simple comparison - same handler, different forms
	r.GET("/compare/form1", func() string {
		return "This is Form 1: Simple return value"
	})

	r.GET("/compare/form2", func() (string, error) {
		return "This is Form 2: Return with error", nil
	})

	type MessageRequest struct {
		Text string `query:"text"`
	}

	r.GET("/compare/form3", func(req *MessageRequest) (string, error) {
		if req.Text == "" {
			req.Text = "default message"
		}
		return fmt.Sprintf("This is Form 3: %s", req.Text), nil
	})

	r.GET("/compare/form4", func(ctx *request.Context) (string, error) {
		userAgent := ctx.R.Header.Get("User-Agent")
		return fmt.Sprintf("This is Form 4: Requested by %s", userAgent), nil
	})

	app := lokstra.NewApp("handler-forms", ":3000", r)

	fmt.Println("🚀 Server running on http://localhost:3000")
	fmt.Println("\n📖 Try different handler forms:")
	fmt.Println("\n   # Form 1: Simple return")
	fmt.Println("   curl http://localhost:3000/ping")
	fmt.Println("   curl http://localhost:3000/time")
	fmt.Println("\n   # Form 2: Return with error")
	fmt.Println("   curl http://localhost:3000/users")
	fmt.Println("\n   # Form 3: Request binding")
	fmt.Println("   curl -X POST http://localhost:3000/users -H 'Content-Type: application/json' -d '{\"name\":\"Charlie\",\"email\":\"charlie@example.com\"}'")
	fmt.Println("   curl http://localhost:3000/users/1")
	fmt.Println("\n   # Form 4: With context")
	fmt.Println("   curl http://localhost:3000/users/1/details")
	fmt.Println("\n   # Form 5: Custom response")
	fmt.Println("   curl -i http://localhost:3000/users/1/custom")
	fmt.Println("\n   # Compare forms")
	fmt.Println("   curl http://localhost:3000/compare/form1")
	fmt.Println("   curl http://localhost:3000/compare/form3?text=hello")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
