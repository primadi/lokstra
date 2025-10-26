package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
)

// Request/Response types
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type GetUserParams struct {
	ID int `path:"id"`
}

type HeaderParams struct {
	Authorization string `header:"Authorization"`
	UserAgent     string `header:"User-Agent"`
	CustomHeader  string `header:"X-Custom-Header"`
}

type ApiCreatedParams struct {
	Data map[string]any `json:"*"` // Bind entire JSON body
}

func main() {
	r := lokstra.NewRouter("api")

	// ========================================
	// GROUP 1: Simplest Forms (Auto JSON via ApiHelper)
	// ========================================

	// Form 1: Return value - Auto wrapped in ApiHelper response
	r.GET("/ping", func() string {
		return "pong"
	})

	// Form 2: Return value + error
	r.GET("/time", func() (map[string]any, error) {
		return map[string]any{
			"timestamp": time.Now().Unix(),
			"datetime":  time.Now().Format(time.RFC3339),
		}, nil
	})

	// Form 3: Return slice - Auto wrapped
	r.GET("/users", func() ([]User, error) {
		return []User{
			{ID: 1, Name: "Alice", Email: "alice@example.com"},
			{ID: 2, Name: "Bob", Email: "bob@example.com"},
			{ID: 3, Name: "Charlie", Email: "charlie@example.com"},
		}, nil
	})

	// ========================================
	// GROUP 2: Request Binding
	// ========================================

	// Form 4: Request binding from JSON body
	r.POST("/users", func(req *CreateUserRequest) (*User, error) {
		// req auto-validated and bound from JSON body
		return &User{
			ID:    123,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	// Form 5: Path params binding
	r.GET("/users/{id}", func(req *GetUserParams) (*User, error) {
		// req.ID extracted from path
		return &User{
			ID:    req.ID,
			Name:  fmt.Sprintf("User %d", req.ID),
			Email: fmt.Sprintf("user%d@example.com", req.ID),
		}, nil
	})

	// Form 6: Context + Request binding
	r.PUT("/users/{id}", func(ctx *request.Context, req *CreateUserRequest) error {
		id := ctx.Req.PathParam("id", "")

		// Using ApiHelper (opiniated)
		return ctx.Api.Ok(map[string]any{
			"id":      id,
			"updated": req,
		})
	})

	// Form 7: Header binding
	r.GET("/headers", func(req *HeaderParams) (map[string]any, error) {
		// req auto-bound from HTTP headers
		return map[string]any{
			"authorization": req.Authorization,
			"user_agent":    req.UserAgent,
			"custom_header": req.CustomHeader,
		}, nil
	})

	// ========================================
	// GROUP 3: Response Methods - ApiHelper (Opiniated)
	// ========================================

	// Method 1: ctx.Api.Ok - Standard success response
	r.GET("/api-ok", func(ctx *request.Context) error {
		return ctx.Api.Ok(map[string]string{
			"message": "Success with ApiHelper",
		})
	})

	// Method 2: ctx.Api.OkWithMessage - Success with custom message
	r.GET("/api-ok-message", func(ctx *request.Context) error {
		return ctx.Api.OkWithMessage(
			map[string]string{"status": "completed"},
			"Operation completed successfully",
		)
	})

	// Method 3: ctx.Api.Created - 201 Created
	r.POST("/api-created", func(ctx *request.Context, req *ApiCreatedParams) error {
		return ctx.Api.Created(map[string]any{
			"id":   123,
			"data": req.Data,
		}, "Resource created successfully")
	})

	// Method 4: ctx.Api.Error - Custom error
	r.GET("/api-not-found", func(ctx *request.Context) error {
		return ctx.Api.Error(404, "NOT_FOUND", "Resource not found")
	})

	// Method 5: ctx.Api.BadRequest - 400 error
	r.GET("/api-bad-request", func(ctx *request.Context) error {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request parameters")
	})

	// ========================================
	// GROUP 4: Response Methods - response.Response (Generic, Not Opiniated)
	// ========================================

	// Method 1: response.Response with Json()
	r.GET("/resp-json", func() (*response.Response, error) {
		resp := response.NewResponse()
		resp.Json(map[string]string{
			"message": "Generic JSON response",
			"type":    "custom",
		})
		return resp, nil
	})

	// Method 2: response.Response with Html()
	r.GET("/resp-html", func() (*response.Response, error) {
		resp := response.NewResponse()
		resp.Html("<h1>Hello from Lokstra</h1><p>HTML response</p>")
		return resp, nil
	})

	// Method 3: response.Response with Text()
	r.GET("/resp-text", func() (*response.Response, error) {
		resp := response.NewResponse()
		resp.Text("Plain text response")
		return resp, nil
	})

	// Method 4: response.Response with custom status
	r.POST("/resp-custom-status", func() (*response.Response, error) {
		resp := response.NewResponse()
		resp.WithStatus(202) // 202 Accepted
		resp.Json(map[string]string{
			"status": "accepted",
		})
		return resp, nil
	})

	// Method 5: response.Response with Stream (file download)
	r.GET("/resp-download", func() (*response.Response, error) {
		resp := response.NewResponse()

		// Add custom headers
		resp.RespHeaders = map[string][]string{
			"Content-Disposition": {"attachment; filename=file.txt"},
			"Content-Type":        {"text/plain"},
		}

		resp.Stream("text/plain", func(w http.ResponseWriter) error {
			_, err := w.Write([]byte("File content here"))
			return err
		})

		return resp, nil
	})

	// ========================================
	// GROUP 5: Response Methods - Manual (http.ResponseWriter)
	// ========================================

	// Method 1: Direct write to ResponseWriter
	r.GET("/manual-json", func(ctx *request.Context) error {
		ctx.W.Header().Set("Content-Type", "application/json")
		ctx.W.WriteHeader(http.StatusOK)
		ctx.W.Write([]byte(`{"message":"Manual JSON response","method":"direct"}`))
		return nil
	})

	// Method 2: Manual with custom headers
	r.GET("/manual-custom", func(ctx *request.Context) error {
		ctx.W.Header().Set("Content-Type", "application/json")
		ctx.W.Header().Set("X-Custom-Header", "custom-value")
		ctx.W.Header().Set("X-Request-ID", "req-123")
		ctx.W.WriteHeader(http.StatusOK)
		ctx.W.Write([]byte(`{"message":"Manual response with custom headers"}`))
		return nil
	})

	// Method 3: Manual text response
	r.GET("/manual-text", func(ctx *request.Context) error {
		ctx.W.Header().Set("Content-Type", "text/plain")
		ctx.W.WriteHeader(http.StatusOK)
		ctx.W.Write([]byte("Manual plain text response"))
		return nil
	})

	// ========================================
	// GROUP 6: Error Handling
	// ========================================

	// Return error (auto 500)
	r.GET("/error-500", func() (string, error) {
		return "", fmt.Errorf("something went wrong")
	})

	// Validation error (auto 400 with field errors)
	r.POST("/validate", func(req *CreateUserRequest) (*User, error) {
		// If validation fails, auto returns 400 with field errors
		return &User{
			ID:    456,
			Name:  req.Name,
			Email: req.Email,
		}, nil
	})

	// ========================================
	// Run App
	// ========================================

	app := lokstra.NewApp("handler-forms", ":3001", r)

	app.PrintStartInfo()
	if err := app.Run(30 * time.Second); err != nil {
		panic(err) // Or use log.Fatal(err)
	}
}
