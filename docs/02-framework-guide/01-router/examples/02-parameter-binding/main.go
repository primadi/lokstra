package main

import (
	"fmt"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/request"
)

// ============================================
// Path Parameters
// ============================================

type PathParams struct {
	ID       int    `path:"id"`
	Category string `path:"category"`
	Slug     string `path:"slug"`
}

func GetByPath(ctx *request.Context, params PathParams) (map[string]any, error) {
	return map[string]any{
		"message":  "Path parameters extracted",
		"id":       params.ID,
		"category": params.Category,
		"slug":     params.Slug,
	}, nil
}

// ============================================
// Query Parameters
// ============================================

type QueryParams struct {
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
	Sort   string `query:"sort"`
	Filter string `query:"filter"`
	Search string `query:"search"`
}

func GetWithQuery(ctx *request.Context, params QueryParams) (map[string]any, error) {
	// Set defaults
	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 10
	}
	if params.Sort == "" {
		params.Sort = "created_at"
	}

	return map[string]any{
		"message": "Query parameters extracted",
		"page":    params.Page,
		"limit":   params.Limit,
		"sort":    params.Sort,
		"filter":  params.Filter,
		"search":  params.Search,
	}, nil
}

// ============================================
// Header Parameters
// ============================================

func GetWithHeaders(ctx *request.Context) (map[string]any, error) {
	return map[string]any{
		"message":       "Headers extracted",
		"api_key":       ctx.Req.HeaderParam("X-API-Key", ""),
		"user_agent":    ctx.Req.HeaderParam("User-Agent", "unknown"),
		"content_type":  ctx.Req.HeaderParam("Content-Type", ""),
		"authorization": ctx.Req.HeaderParam("Authorization", ""),
	}, nil
}

// ============================================
// Combined Parameters
// ============================================

type CombinedParams struct {
	// Path params
	UserID int `path:"user_id"`

	// Query params
	Include string `query:"include"`
	Fields  string `query:"fields"`

	// No header params in struct - use ctx.Req.HeaderParam()
}

func GetCombined(ctx *request.Context, params CombinedParams) (map[string]any, error) {
	return map[string]any{
		"message": "Combined path, query, and header params",
		"user_id": params.UserID,
		"include": params.Include,
		"fields":  params.Fields,
		"api_key": ctx.Req.HeaderParam("X-API-Key", "none"),
	}, nil
}

// ============================================
// Body Binding
// ============================================

type CreateUserRequest struct {
	Name   string   `json:"name" validate:"required"`
	Email  string   `json:"email" validate:"required,email"`
	Age    int      `json:"age" validate:"min=0,max=150"`
	Tags   []string `json:"tags"`
	Active bool     `json:"active"`
}

func CreateUser(ctx *request.Context, body CreateUserRequest) (map[string]any, error) {
	return map[string]any{
		"message": "User created",
		"user":    body,
		"id":      123,
	}, nil
}

// ============================================
// Update with Pointer Fields (Partial Updates)
// ============================================

type UpdateUserRequest struct {
	Name   *string `json:"name,omitempty"`
	Email  *string `json:"email,omitempty"`
	Age    *int    `json:"age,omitempty"`
	Active *bool   `json:"active,omitempty"`
}

func UpdateUser(ctx *request.Context, params PathParams, body UpdateUserRequest) (map[string]any, error) {
	updates := make(map[string]any)

	if body.Name != nil {
		updates["name"] = *body.Name
	}
	if body.Email != nil {
		updates["email"] = *body.Email
	}
	if body.Age != nil {
		updates["age"] = *body.Age
	}
	if body.Active != nil {
		updates["active"] = *body.Active
	}

	return map[string]any{
		"message": "User updated",
		"id":      params.ID,
		"updates": updates,
	}, nil
}

// ============================================
// Custom Parameter Types
// ============================================

type DateRangeParams struct {
	StartDate string `query:"start_date"` // Format: 2006-01-02
	EndDate   string `query:"end_date"`
	TimeZone  string `query:"timezone"`
}

func GetByDateRange(ctx *request.Context, params DateRangeParams) (map[string]any, error) {
	// Parse dates (in production, add validation)
	startDate := params.StartDate
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}

	endDate := params.EndDate
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	timezone := params.TimeZone
	if timezone == "" {
		timezone = "UTC"
	}

	return map[string]any{
		"message":    "Date range query",
		"start_date": startDate,
		"end_date":   endDate,
		"timezone":   timezone,
	}, nil
}

// ============================================
// Array Parameters
// ============================================

type FilterParams struct {
	IDs      []int    `query:"ids"`      // ?ids=1,2,3
	Tags     []string `query:"tags"`     // ?tags=go,rust,python
	Statuses []string `query:"statuses"` // ?statuses=active,pending
}

func GetWithArrays(ctx *request.Context, params FilterParams) (map[string]any, error) {
	return map[string]any{
		"message":  "Array parameters extracted",
		"ids":      params.IDs,
		"tags":     params.Tags,
		"statuses": params.Statuses,
	}, nil
}

// ============================================
// Validation Example
// ============================================

type ValidatedParams struct {
	ID    int    `path:"id" validate:"required,min=1"`
	Email string `query:"email" validate:"required,email"`
	Age   int    `query:"age" validate:"min=0,max=150"`
}

func GetWithValidation(ctx *request.Context, params ValidatedParams) (map[string]any, error) {
	// Validation happens automatically based on tags
	// If validation fails, Lokstra returns 400 Bad Request

	return map[string]any{
		"message": "Validation passed",
		"id":      params.ID,
		"email":   params.Email,
		"age":     params.Age,
	}, nil
}

func main() {
	router := lokstra.NewRouter("parameter-binding")

	// Path parameters
	router.GET("/path/:id/:category/:slug", GetByPath)

	// Query parameters
	router.GET("/query", GetWithQuery)

	// Headers
	router.GET("/headers", GetWithHeaders)

	// Combined
	router.GET("/combined/:user_id", GetCombined)

	// Body binding
	router.POST("/users", CreateUser)
	router.PATCH("/users/:id", UpdateUser)

	// Date range
	router.GET("/reports", GetByDateRange)

	// Arrays
	router.GET("/filter", GetWithArrays)

	// Validation
	router.GET("/validate/:id", GetWithValidation)

	app := lokstra.NewApp("param-demo", ":3000", router)

	fmt.Println("🚀 Parameter Binding Demo")
	fmt.Println("📖 Server: http://localhost:3000")
	fmt.Println("\nEndpoints:")
	fmt.Println("  GET  /path/:id/:category/:slug  - Path parameters")
	fmt.Println("  GET  /query                     - Query parameters")
	fmt.Println("  GET  /headers                   - Header parameters")
	fmt.Println("  GET  /combined/:user_id         - Combined params")
	fmt.Println("  POST /users                     - Body binding")
	fmt.Println("  PATCH /users/:id                - Partial updates")
	fmt.Println("  GET  /reports                   - Date ranges")
	fmt.Println("  GET  /filter                    - Array params")
	fmt.Println("  GET  /validate/:id              - With validation")
	fmt.Println("\nUse test.http to test all examples")

	if err := app.Run(30 * time.Second); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
