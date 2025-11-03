package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/lokstra_registry"
)

// This example demonstrates auto-router registration with custom conventions
// See index for detailed explanation

type UserService struct{}

func (*UserService) GetResourceName() (string, string) {
	return "user", "users"
}

func (*UserService) GetConventionName() string {
	return "rest"
}

func (*UserService) List() *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"users": []map[string]any{
			{"id": 1, "name": "John"},
			{"id": 2, "name": "Jane"},
		},
	})
}

type GetByIDParams struct {
	ID int `path:"id"`
}

func (*UserService) GetByID(params *GetByIDParams) *response.ApiHelper {
	return response.NewApiOk(map[string]any{
		"id":   params.ID,
		"name": fmt.Sprintf("User %d", params.ID),
	})
}

func UserServiceFactory(deps map[string]any, config map[string]any) any {
	return &UserService{}
}

func Home() *response.Response {
	return response.NewHtmlResponse(`<!DOCTYPE html>
<html><head><title>Auto-Router Advanced</title></head>
<body><h1>🚦 Auto-Router Advanced Example</h1>
<p>This example demonstrates advanced auto-router features.</p>
<ul><li>GET /users - List users</li>
<li>GET /users/:id - Get user by ID</li></ul>
</body></html>`)
}

func main() {
	lokstra_registry.RegisterServiceType("user-service", UserServiceFactory, nil)
	lokstra_registry.RegisterLazyService("user-service", UserServiceFactory, nil)

	router := lokstra.NewRouter("auto-router-advanced")
	router.GET("/", Home)

	app := lokstra.NewApp("auto-router-advanced", ":3000", router)
	fmt.Println("🚀 Auto-Router Advanced Example")
	fmt.Println("📍 http://localhost:3000")

	if err := app.Run(0); err != nil {
		fmt.Println("Error:", err)
	}
}
