package main

import (
	"regexp"
	"strconv"

	"github.com/primadi/lokstra"
)

// This example demonstrates route constraints and advanced routing patterns.
// It shows how to validate path parameters, implement custom route matching,
// and handle complex URL patterns in Lokstra.
//
// Learning Objectives:
// - Understand route parameter validation
// - Learn custom route constraints
// - Explore advanced routing patterns
// - See URL pattern matching techniques

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "route-constraints-app", ":8080")

	// ===== Basic Path Parameters =====

	app.GET("/users/:id", func(ctx *lokstra.Context) error {
		userID := ctx.GetPathParam("id")

		// Basic validation
		if userID == "" {
			return ctx.ErrorBadRequest("User ID is required")
		}

		return ctx.Ok(map[string]interface{}{
			"message": "Basic path parameter",
			"user_id": userID,
			"type":    "string",
		})
	})

	// ===== Numeric Constraints =====

	app.GET("/users/:id/posts/:postId", func(ctx *lokstra.Context) error {
		userIDStr := ctx.GetPathParam("id")
		postIDStr := ctx.GetPathParam("postId")

		// Validate numeric IDs
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			return ctx.ErrorBadRequest("User ID must be numeric")
		}

		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			return ctx.ErrorBadRequest("Post ID must be numeric")
		}

		return ctx.Ok(map[string]interface{}{
			"message":    "Numeric constraints",
			"user_id":    userID,
			"post_id":    postID,
			"validation": "passed",
		})
	})

	// ===== Custom Validation Middleware =====

	// Middleware to validate numeric ID
	validateNumericID := func(paramName string) func(*lokstra.Context, func(*lokstra.Context) error) error {
		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			value := ctx.GetPathParam(paramName)

			if _, err := strconv.Atoi(value); err != nil {
				return ctx.ErrorBadRequest(paramName + " must be a valid number")
			}

			return next(ctx)
		}
	}

	// Routes with numeric validation
	app.GET("/validated/users/:id",
		validateNumericID("id"),
		func(ctx *lokstra.Context) error {
			userID, _ := strconv.Atoi(ctx.GetPathParam("id"))
			return ctx.Ok(map[string]interface{}{
				"message":    "Validated numeric ID",
				"user_id":    userID,
				"validation": "middleware",
			})
		})

	// ===== Pattern-Based Constraints =====

	// Email validation middleware
	validateEmail := func(paramName string) func(*lokstra.Context, func(*lokstra.Context) error) error {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			email := ctx.GetPathParam(paramName)

			if !emailRegex.MatchString(email) {
				return ctx.ErrorBadRequest("Invalid email format")
			}

			return next(ctx)
		}
	}

	app.GET("/users/email/:email",
		validateEmail("email"),
		func(ctx *lokstra.Context) error {
			email := ctx.GetPathParam("email")
			return ctx.Ok(map[string]interface{}{
				"message":    "Valid email parameter",
				"email":      email,
				"validation": "regex",
			})
		})

	// UUID validation
	validateUUID := func(paramName string) func(*lokstra.Context, func(*lokstra.Context) error) error {
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			uuid := ctx.GetPathParam(paramName)

			if !uuidRegex.MatchString(uuid) {
				return ctx.ErrorBadRequest("Invalid UUID format")
			}

			return next(ctx)
		}
	}

	app.GET("/resources/:uuid",
		validateUUID("uuid"),
		func(ctx *lokstra.Context) error {
			uuid := ctx.GetPathParam("uuid")
			return ctx.Ok(map[string]interface{}{
				"message":    "Valid UUID parameter",
				"uuid":       uuid,
				"validation": "uuid_regex",
			})
		})

	// ===== Range Constraints =====

	validateRange := func(paramName string, min, max int) func(*lokstra.Context, func(*lokstra.Context) error) error {
		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			valueStr := ctx.GetPathParam(paramName)

			value, err := strconv.Atoi(valueStr)
			if err != nil {
				return ctx.ErrorBadRequest(paramName + " must be a number")
			}

			if value < min || value > max {
				return ctx.ErrorBadRequest(paramName + " must be between " + strconv.Itoa(min) + " and " + strconv.Itoa(max))
			}

			return next(ctx)
		}
	}

	app.GET("/pages/:pageNum",
		validateRange("pageNum", 1, 100),
		func(ctx *lokstra.Context) error {
			pageNum, _ := strconv.Atoi(ctx.GetPathParam("pageNum"))
			return ctx.Ok(map[string]interface{}{
				"message":    "Valid page number",
				"page":       pageNum,
				"validation": "range_1_100",
			})
		})

	// ===== Enum Constraints =====

	validateEnum := func(paramName string, validValues []string) func(*lokstra.Context, func(*lokstra.Context) error) error {
		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			value := ctx.GetPathParam(paramName)

			for _, valid := range validValues {
				if value == valid {
					return next(ctx)
				}
			}

			return ctx.ErrorBadRequest(paramName + " must be one of: " + joinStrings(validValues, ", "))
		}
	}

	app.GET("/users/:id/profile/:section",
		validateNumericID("id"),
		validateEnum("section", []string{"basic", "contact", "preferences", "security"}),
		func(ctx *lokstra.Context) error {
			userID, _ := strconv.Atoi(ctx.GetPathParam("id"))
			section := ctx.GetPathParam("section")

			return ctx.Ok(map[string]interface{}{
				"message":    "Profile section",
				"user_id":    userID,
				"section":    section,
				"validation": "numeric_id_and_enum",
			})
		})

	// ===== Complex Pattern Matching =====

	// Slug validation (URL-friendly strings)
	validateSlug := func(paramName string) func(*lokstra.Context, func(*lokstra.Context) error) error {
		slugRegex := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			slug := ctx.GetPathParam(paramName)

			if !slugRegex.MatchString(slug) {
				return ctx.ErrorBadRequest("Invalid slug format (use lowercase letters, numbers, and hyphens)")
			}

			return next(ctx)
		}
	}

	app.GET("/blog/:slug",
		validateSlug("slug"),
		func(ctx *lokstra.Context) error {
			slug := ctx.GetPathParam("slug")
			return ctx.Ok(map[string]interface{}{
				"message":      "Blog post",
				"slug":         slug,
				"validation":   "slug_format",
				"url_friendly": true,
			})
		})

	// Date validation (YYYY-MM-DD)
	validateDate := func(paramName string) func(*lokstra.Context, func(*lokstra.Context) error) error {
		dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

		return func(ctx *lokstra.Context, next func(*lokstra.Context) error) error {
			date := ctx.GetPathParam(paramName)

			if !dateRegex.MatchString(date) {
				return ctx.ErrorBadRequest("Invalid date format (use YYYY-MM-DD)")
			}

			return next(ctx)
		}
	}

	app.GET("/reports/:date",
		validateDate("date"),
		func(ctx *lokstra.Context) error {
			date := ctx.GetPathParam("date")
			return ctx.Ok(map[string]interface{}{
				"message":    "Daily report",
				"date":       date,
				"validation": "date_format",
			})
		})

	// ===== Multiple Constraints =====

	app.GET("/api/:version/users/:userId/posts/:postId",
		validateEnum("version", []string{"v1", "v2", "v3"}),
		validateNumericID("userId"),
		validateNumericID("postId"),
		func(ctx *lokstra.Context) error {
			version := ctx.GetPathParam("version")
			userID, _ := strconv.Atoi(ctx.GetPathParam("userId"))
			postID, _ := strconv.Atoi(ctx.GetPathParam("postId"))

			return ctx.Ok(map[string]interface{}{
				"message":     "API endpoint with multiple constraints",
				"api_version": version,
				"user_id":     userID,
				"post_id":     postID,
				"validation":  "version_enum_and_numeric_ids",
			})
		})

	// ===== Query Parameter Validation =====

	app.GET("/search", func(ctx *lokstra.Context) error {
		// Validate required query parameters
		query := ctx.GetQueryParam("q")
		if query == "" {
			return ctx.ErrorBadRequest("Query parameter 'q' is required")
		}

		// Validate optional numeric parameters
		pageStr := ctx.GetQueryParam("page")
		page := 1
		if pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil || page < 1 {
				return ctx.ErrorBadRequest("Page must be a number >= 1")
			}
		}

		limitStr := ctx.GetQueryParam("limit")
		limit := 10
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit < 1 || limit > 100 {
				return ctx.ErrorBadRequest("Limit must be a number between 1 and 100")
			}
		}

		// Validate enum parameter
		sortBy := ctx.GetQueryParam("sort")
		if sortBy != "" {
			validSorts := []string{"date", "relevance", "popularity"}
			if !contains(validSorts, sortBy) {
				return ctx.ErrorBadRequest("Sort must be one of: " + joinStrings(validSorts, ", "))
			}
		}

		return ctx.Ok(map[string]interface{}{
			"message":    "Search results",
			"query":      query,
			"page":       page,
			"limit":      limit,
			"sort_by":    sortBy,
			"validation": "query_parameters",
		})
	})

	// ===== Helper Routes for Testing =====

	app.GET("/", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]interface{}{
			"message": "Route Constraints Example",
			"examples": map[string]string{
				"basic_param":          "/users/123",
				"numeric_ids":          "/users/123/posts/456",
				"validated_id":         "/validated/users/789",
				"email_param":          "/users/email/john@example.com",
				"uuid_param":           "/resources/550e8400-e29b-41d4-a716-446655440000",
				"page_range":           "/pages/5",
				"profile_section":      "/users/123/profile/basic",
				"blog_slug":            "/blog/my-first-post",
				"date_param":           "/reports/2024-01-15",
				"multiple_constraints": "/api/v2/users/123/posts/456",
				"search_query":         "/search?q=lokstra&page=1&limit=10&sort=relevance",
			},
		})
	})

	lokstra.Logger.Infof("🚀 Route Constraints Example started on :8080")
	lokstra.Logger.Infof("")
	lokstra.Logger.Infof("Route Constraint Examples:")
	lokstra.Logger.Infof("  GET  /users/:id                     - Basic path parameter")
	lokstra.Logger.Infof("  GET  /users/:id/posts/:postId       - Numeric validation")
	lokstra.Logger.Infof("  GET  /validated/users/:id           - Middleware validation")
	lokstra.Logger.Infof("  GET  /users/email/:email            - Email format validation")
	lokstra.Logger.Infof("  GET  /resources/:uuid               - UUID format validation")
	lokstra.Logger.Infof("  GET  /pages/:pageNum                - Range constraints (1-100)")
	lokstra.Logger.Infof("  GET  /users/:id/profile/:section    - Enum constraints")
	lokstra.Logger.Infof("  GET  /blog/:slug                    - Slug format validation")
	lokstra.Logger.Infof("  GET  /reports/:date                 - Date format (YYYY-MM-DD)")
	lokstra.Logger.Infof("  GET  /api/:version/users/:userId/posts/:postId - Multiple constraints")
	lokstra.Logger.Infof("  GET  /search                        - Query parameter validation")

	app.Start()
}

// Utility functions
func joinStrings(strings []string, separator string) string {
	if len(strings) == 0 {
		return ""
	}

	result := strings[0]
	for i := 1; i < len(strings); i++ {
		result += separator + strings[i]
	}
	return result
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
