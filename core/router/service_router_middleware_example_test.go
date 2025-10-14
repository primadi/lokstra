package router_test

import (
	"fmt"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// Example service
type UserService struct{}

func (s *UserService) GetUser(ctx *request.Context, id string) (map[string]any, error) {
	return map[string]any{"id": id, "name": "John Doe"}, nil
}

func (s *UserService) ListUsers(ctx *request.Context) ([]map[string]any, error) {
	return []map[string]any{
		{"id": "1", "name": "John"},
		{"id": "2", "name": "Jane"},
	}, nil
}

func (s *UserService) CreateUser(ctx *request.Context, req struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}) (map[string]any, error) {
	return map[string]any{"id": "new", "name": req.Name}, nil
}

func (s *UserService) DeleteUser(ctx *request.Context, id string) error {
	return nil
}

// Example 1: Using WithMethodMiddleware for simple cases
func ExampleServiceRouterOptions_WithMethodMiddleware() {
	opts := router.DefaultServiceRouterOptions().
		WithPrefix("/api").
		WithMethodMiddleware("DeleteUser", "auth", "admin_check").
		WithMethodMiddleware("CreateUser", "auth", "rate_limit")

	r := router.NewFromService(&UserService{}, opts)
	r.Build()

	fmt.Println("Router with method-specific middleware created")
	// Output: Router with method-specific middleware created
}

// Example 2: Using RouteOverride for full control (path, method, middleware)
func ExampleServiceRouterOptions_WithRouteOverride() {
	opts := router.DefaultServiceRouterOptions().
		WithPrefix("/api").
		WithRouteOverride("DeleteUser", router.RouteMeta{
			HTTPMethod:  "DELETE",
			Path:        "/users/{id}",
			Name:        "delete-user-by-id",
			Description: "Delete a user by their ID",
			Middlewares: []any{"auth", "admin_check", "audit"},
		}).
		WithRouteOverride("GetUser", router.RouteMeta{
			// Path and HTTPMethod empty = use convention
			Name:        "get-user-detail",
			Description: "Retrieve user details",
			Middlewares: []any{"auth"},
		})

	r := router.NewFromService(&UserService{}, opts)
	r.Build()

	fmt.Println("Router with route overrides created")
	// Output: Router with route overrides created
}

// Example 3: Combining global and method-specific middleware
func ExampleServiceRouterOptions_combined() {
	opts := router.DefaultServiceRouterOptions().
		WithPrefix("/api/v1").
		WithMiddlewares("cors", "logging").                // Applied to all routes
		WithMethodMiddleware("DeleteUser", "admin_check"). // Additional for DeleteUser
		WithMethodMiddleware("CreateUser", "rate_limit")   // Additional for CreateUser

	r := router.NewFromService(&UserService{}, opts)
	r.Build()

	fmt.Println("Router with combined middleware created")
	// Output: Router with combined middleware created
}

// Example 4: Only adding middleware without changing route
func ExampleServiceRouterOptions_middlewareOnly() {
	opts := router.DefaultServiceRouterOptions().
		WithPrefix("/users").
		// Just add middleware, use convention for path and HTTP method
		WithMethodMiddleware("GetUser", "cache").
		WithMethodMiddleware("ListUsers", "cache").
		WithMethodMiddleware("DeleteUser", "auth", "admin_check", "audit")

	r := router.NewFromService(&UserService{}, opts)
	r.Build()

	fmt.Println("Router with middleware-only overrides created")
	// Output: Router with middleware-only overrides created
}

// Example 5: Override parent middleware
func ExampleServiceRouterOptions_overrideParentMiddleware() {
	opts := router.DefaultServiceRouterOptions().
		WithPrefix("/api").
		WithMiddlewares("auth", "logging"). // Global middleware
		WithRouteOverride("ListUsers", router.RouteMeta{
			OverrideParentMw: true,                // Skip parent middleware
			Middlewares:      []any{"rate_limit"}, // Only this
		})

	r := router.NewFromService(&UserService{}, opts)
	r.Build()

	fmt.Println("Router with parent middleware override created")
	// Output: Router with parent middleware override created
}

// Example 6: Full route customization
func ExampleServiceRouterOptions_fullCustomization() {
	opts := router.DefaultServiceRouterOptions().
		WithPrefix("/api/v1").
		WithRouteOverride("GetUser", router.RouteMeta{
			Path:             "/user/{id}/profile",
			Name:             "user-profile",
			Description:      "Get user profile with extended information",
			OverrideParentMw: false,
			Middlewares:      []any{"auth", "cache", "metrics"},
		})

	r := router.NewFromService(&UserService{}, opts)
	r.Build()

	fmt.Println("Router with full customization created")
	// Output: Router with full customization created
}
