package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/core/response/api_formatter"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role,omitempty"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=3"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,min=18"`
}

type ListUsersRequest struct {
	request.PagingRequest
	Role   string `query:"role"`
	Status string `query:"status"`
}

var users = []User{
	{ID: 1, Name: "Alice", Email: "alice@example.com", Role: "admin"},
	{ID: 2, Name: "Bob", Email: "bob@example.com", Role: "user"},
	{ID: 3, Name: "Charlie", Email: "charlie@example.com", Role: "user"},
	{ID: 4, Name: "Diana", Email: "diana@example.com", Role: "admin"},
	{ID: 5, Name: "Eve", Email: "eve@example.com", Role: "user"},
}

func main() {
	r := lokstra.NewRouter("api")

	// ========================================================================
	// RESPONSE PATH 1: Using request.Context (ctx)
	// ========================================================================

	// --- Method 1A: Manual Response (http.ResponseWriter) ---
	r.GET("/manual/json", func(ctx *request.Context) error {
		// Full manual control using http.ResponseWriter
		ctx.W.Header().Set("Content-Type", "application/json")
		ctx.W.Header().Set("X-Custom-Header", "manual-response")
		ctx.W.WriteHeader(200)
		ctx.W.Write([]byte(`{"message":"Manual JSON response","method":"http.ResponseWriter"}`))
		return nil
	})

	r.GET("/manual/text", func(ctx *request.Context) error {
		// Manual text response
		ctx.W.Header().Set("Content-Type", "text/plain")
		ctx.W.WriteHeader(200)
		ctx.W.Write([]byte("Plain text response using manual method"))
		return nil
	})

	// ========================================================================
	// GROUP 1: Via Context - ctx.Resp (Generic Response)
	// ========================================================================

	r.GET("/ctx-resp/json", func(ctx *request.Context) error {
		// Using ctx.Resp for JSON - can return directly
		ctx.Resp.RespHeaders = map[string][]string{
			"X-Via": {"ctx.Resp"},
		}
		return ctx.Resp.WithStatus(200).Json(map[string]any{
			"message": "Response via ctx.Resp",
			"method":  "context",
			"users":   users[:2],
		})
	})

	r.GET("/ctx-resp/html", func(ctx *request.Context) error {
		// Using ctx.Resp for HTML - can return directly
		html := `<html><body><h1>Via ctx.Resp</h1><p>HTML response using context</p></body></html>`
		return ctx.Resp.Html(html)
	})

	r.GET("/ctx-resp/text", func(ctx *request.Context) error {
		// Using ctx.Resp for plain text - can return directly
		return ctx.Resp.Text("Plain text via ctx.Resp")
	})

	// ========================================================================
	// GROUP 2: Via Context - ctx.Api (Opinionated API)
	// ========================================================================

	r.GET("/ctx-api/success", func(ctx *request.Context) error {
		// Using ctx.Api for success response - can return directly
		return ctx.Api.Ok(users)
	})

	r.GET("/ctx-api/success-message", func(ctx *request.Context) error {
		// Using ctx.Api with message - can return directly
		return ctx.Api.OkWithMessage(users[:2], "First 2 users retrieved")
	})

	r.POST("/ctx-api/created", func(ctx *request.Context) error {
		// Using ctx.Api for created response - can return directly
		newUser := User{ID: 6, Name: "Frank", Email: "frank@example.com"}
		return ctx.Api.Created(newUser, "User created via ctx.Api")
	})

	r.GET("/ctx-api/error-notfound", func(ctx *request.Context) error {
		// Using ctx.Api for error - can return directly
		return ctx.Api.NotFound("User not found via ctx.Api")
	})

	r.GET("/ctx-api/error-validation", func(ctx *request.Context) error {
		// Using ctx.Api for validation error - can return directly
		fieldErrors := []api_formatter.FieldError{
			{Field: "username", Code: "REQUIRED", Message: "Username is required"},
			{Field: "password", Code: "TOO_SHORT", Message: "Password must be at least 8 characters"},
		}
		return ctx.Api.ValidationError("Validation failed via ctx.Api", fieldErrors)
	})

	// ========================================================================
	// GROUP 3: Via Return - response.Response (Generic Response)
	// ========================================================================

	// --- Method 1B: Generic response.Response (unopinionated) ---
	r.GET("/response/json", func() *response.Response {
		// Generic Response - can be JSON
		resp := response.NewResponse()
		resp.RespHeaders = map[string][]string{
			"X-Response-Type": {"generic-json"},
		}
		resp.WithStatus(200).Json(map[string]any{
			"message": "Generic JSON using response.Response",
			"data":    users,
		})
		return resp
	})

	r.GET("/response/html", func() *response.Response {
		// Generic Response - can be HTML
		resp := response.NewResponse()
		html := `<html><body><h1>HTML Response</h1><p>Using response.Response</p></body></html>`
		resp.Html(html)
		return resp
	})

	r.GET("/response/text", func() *response.Response {
		// Generic Response - can be plain text
		resp := response.NewResponse()
		resp.Text("Plain text using response.Response")
		return resp
	})

	r.GET("/response/custom-status", func() *response.Response {
		// Generic Response with custom status code
		resp := response.NewResponse()
		resp.WithStatus(201).Json(map[string]string{
			"message": "Resource created",
			"status":  "201",
		})
		return resp
	})

	// ========================================================================
	// GROUP 4: Via Return - response.ApiHelper (Opinionated API)
	// ========================================================================

	// --- Method 1C: Opinionated response.ApiHelper (JSON only) ---
	r.GET("/api/success", func() *response.ApiHelper {
		// Opinionated API response - standard success format
		api := response.NewApiHelper()
		api.Ok(users)
		return api
	})

	r.GET("/api/success-message", func() *response.ApiHelper {
		// Opinionated API response with custom message
		api := response.NewApiHelper()
		api.OkWithMessage(users, "Users retrieved successfully")
		return api
	})

	r.POST("/api/created", func() *response.ApiHelper {
		// Opinionated API response - 201 Created
		api := response.NewApiHelper()
		newUser := User{ID: 3, Name: "Charlie", Email: "charlie@example.com"}
		api.Created(newUser, "User created successfully")
		return api
	})

	r.GET("/api/error-notfound", func() *response.ApiHelper {
		// Opinionated API error response
		api := response.NewApiHelper()
		api.NotFound("User not found")
		return api
	})

	r.GET("/api/error-validation", func() *response.ApiHelper {
		// Opinionated API validation error with field details
		api := response.NewApiHelper()

		// ValidationError accepts []FieldError for detailed field-level errors
		fieldErrors := []api_formatter.FieldError{
			{Field: "email", Code: "INVALID_FORMAT", Message: "Invalid email format"},
			{Field: "age", Code: "OUT_OF_RANGE", Message: "Age must be at least 18"},
			{Field: "name", Code: "TOO_SHORT", Message: "Name must be at least 3 characters"},
		}

		api.ValidationError("Validation failed", fieldErrors)
		return api
	})

	r.GET("/api/list", func(req *ListUsersRequest) (*response.ApiHelper, error) {
		// Opinionated API with pagination using PagingRequest
		req.SetDefaults()

		api := response.NewApiHelper()

		// Simulate filtered data based on role
		filteredUsers := users
		if req.Role != "" {
			filtered := []User{}
			for _, u := range users {
				if u.Role == req.Role {
					filtered = append(filtered, u)
				}
			}
			filteredUsers = filtered
		}

		// Calculate pagination
		totalItems := len(filteredUsers)

		offset := req.GetOffset()
		limit := req.GetLimit()

		// Slice data for current page
		end := offset + limit
		if end > totalItems {
			end = totalItems
		}

		pageData := filteredUsers[offset:end]

		// Return paginated list with metadata
		meta := api_formatter.CalculateListMeta(req.Page, req.PageSize, totalItems)

		api.OkList(pageData, meta)
		return api, nil
	})

	// ========================================================================
	// GROUP 5: Return Values - Plain Data
	// ========================================================================

	// --- Method 2A: Return any (plain data) ---
	r.GET("/return/data", func() any {
		// Return plain data - auto JSON
		return map[string]any{
			"message": "Direct data return",
			"users":   users,
			"count":   len(users),
		}
	})

	r.GET("/return/struct", func() []User {
		// Return struct - auto JSON
		return users
	})

	// --- Method 2B: Return (any, error) ---
	r.GET("/return/data-error", func() (any, error) {
		// Return data with potential error
		return map[string]any{
			"message": "Data with error handling",
			"users":   users,
		}, nil
	})

	r.GET("/return/data-error-fail", func() (any, error) {
		// Simulate error
		return nil, fmt.Errorf("simulated database error")
	})

	// --- Method 2C: Return *response.Response ---
	r.GET("/return/response", func() (*response.Response, error) {
		// Return Response object
		resp := response.NewResponse()
		resp.RespHeaders = map[string][]string{
			"X-Response-Method": {"return-response"},
		}
		resp.WithStatus(200).Json(map[string]any{
			"message": "Response via return",
			"data":    users,
		})
		return resp, nil
	})

	r.GET("/return/response-html", func() (*response.Response, error) {
		// Return HTML via Response
		resp := response.NewResponse()
		html := `<html><body><h1>HTML via Return</h1></body></html>`
		resp.Html(html)
		return resp, nil
	})

	// --- Method 2D: Return *response.ApiHelper ---
	r.GET("/return/api", func() (*response.ApiHelper, error) {
		// Return ApiHelper object
		api := response.NewApiHelper()
		api.Ok(users)
		return api, nil
	})

	r.POST("/return/api-created", func() (*response.ApiHelper, error) {
		// Return ApiHelper with Created
		api := response.NewApiHelper()
		newUser := User{ID: 4, Name: "Dave", Email: "dave@example.com"}
		api.Created(newUser, "User created via return")
		return api, nil
	})

	r.GET("/return/api-error", func() (*response.ApiHelper, error) {
		// Return ApiHelper with error
		api := response.NewApiHelper()
		api.NotFound("User not found via return")
		return api, nil
	})

	// ========================================================================
	// COMPARISON: Same endpoint, different methods
	// ========================================================================

	r.GET("/compare/manual", func(ctx *request.Context) error {
		ctx.W.Header().Set("Content-Type", "application/json")
		ctx.W.WriteHeader(200)
		ctx.W.Write([]byte(`{"method":"manual","message":"Full control"}`))
		return nil
	})

	r.GET("/compare/response", func() *response.Response {
		resp := response.NewResponse()
		resp.Json(map[string]string{
			"method":  "response.Response",
			"message": "Generic, unopinionated",
		})
		return resp
	})

	r.GET("/compare/api", func() *response.ApiHelper {
		api := response.NewApiHelper()
		api.Ok(map[string]string{
			"method":  "response.ApiHelper",
			"message": "Opinionated, structured",
		})
		return api
	})

	r.GET("/compare/return", func() any {
		return map[string]string{
			"method":  "return data",
			"message": "Simplest, auto JSON",
		}
	})

	// Print routes
	fmt.Println("\n📋 Response Pattern Examples")
	fmt.Println("=" + string(make([]byte, 50)))
	r.PrintRoutes()
	fmt.Println("=" + string(make([]byte, 50)))

	app := lokstra.NewApp("response-patterns", ":3000", r)

	fmt.Println("\n🚀 Server running on http://localhost:3000")
	fmt.Println("\n📖 Test different response patterns:")
	fmt.Println("\n   # Manual Response (http.ResponseWriter)")
	fmt.Println("   curl http://localhost:3000/manual/json")
	fmt.Println("   curl http://localhost:3000/manual/text")
	fmt.Println("\n   # Generic Response (response.Response)")
	fmt.Println("   curl http://localhost:3000/response/json")
	fmt.Println("   curl http://localhost:3000/response/html")
	fmt.Println("   curl http://localhost:3000/response/text")
	fmt.Println("\n   # Opinionated API (response.ApiHelper)")
	fmt.Println("   curl http://localhost:3000/api/success")
	fmt.Println("   curl http://localhost:3000/api/error-notfound")
	fmt.Println("\n   # Return Values")
	fmt.Println("   curl http://localhost:3000/return/data")
	fmt.Println("   curl http://localhost:3000/return/response")
	fmt.Println("   curl http://localhost:3000/return/api")
	fmt.Println("\n   # Comparison")
	fmt.Println("   curl http://localhost:3000/compare/manual")
	fmt.Println("   curl http://localhost:3000/compare/response")
	fmt.Println("   curl http://localhost:3000/compare/api")
	fmt.Println("   curl http://localhost:3000/compare/return")

	app.Run(30 * time.Second)
}
