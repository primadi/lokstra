package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/common/api_client"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ============================================
// Remote Service Pattern
// ============================================

type RemoteUserService struct {
	client *api_client.ClientRouter
}

func NewRemoteUserService(routerName, pathPrefix string) *RemoteUserService {
	// Note: In production, use lokstra_registry.GetRouter() and api_client.ClientRouter
	fmt.Printf("✓ Created RemoteUserService: router=%s, prefix=%s\n", routerName, pathPrefix)
	return &RemoteUserService{client: nil} // Placeholder for demonstration
}

type UserResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (s *RemoteUserService) GetUser(id int) (*UserResponse, error) {
	// Placeholder implementation - demonstrates the pattern
	// In production, use s.client.GET() or FetchAndCast
	return &UserResponse{
		ID:    id,
		Name:  "Demo User",
		Email: "demo@example.com",
	}, nil
}

func (s *RemoteUserService) CreateUser(name, email string) (*UserResponse, error) {
	// Placeholder implementation - demonstrates the pattern
	// In production, use s.client.POST() or FetchAndCast
	return &UserResponse{
		ID:    1,
		Name:  name,
		Email: email,
	}, nil
}

func RemoteUserServiceFactory(deps map[string]any, config map[string]any) any {
	routerName := "user-api"
	if router, ok := config["router"].(string); ok {
		routerName = router
	}

	pathPrefix := "/api/v1"
	if prefix, ok := config["path-prefix"].(string); ok {
		pathPrefix = prefix
	}

	return NewRemoteUserService(routerName, pathPrefix)
}

// ============================================
// HTTP Handlers
// ============================================

type GetUserParams struct {
	ID int `path:"id"`
}

func GetUser(params *GetUserParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*RemoteUserService]("remote-user-service")
	result, err := svc.GetUser(params.ID)
	if err != nil {
		return response.NewApiInternalError(fmt.Sprintf("Remote call failed: %v", err))
	}
	return response.NewApiOk(result)
}

type CreateUserParams struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

func CreateUser(params *CreateUserParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*RemoteUserService]("remote-user-service")
	result, err := svc.CreateUser(params.Name, params.Email)
	if err != nil {
		return response.NewApiInternalError(fmt.Sprintf("Remote call failed: %v", err))
	}
	return response.NewApiCreated(result, "User created successfully")
}

func Home() *response.Response {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Remote Services Example</title>
    <style>
        body { font-family: Arial; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #333; }
        .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
        .method { display: inline-block; padding: 3px 8px; border-radius: 3px; font-weight: bold; color: white; }
        .get { background: #61affe; }
        .post { background: #49cc90; }
        code { background: #eee; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🌐 Remote Services Example</h1>
    
    <p>This example demonstrates HTTP-based service communication using ClientRouter.</p>

    <h2>Features</h2>
    <ul>
        <li><strong>ClientRouter</strong> - HTTP client for remote API calls</li>
        <li><strong>FetchAndCast</strong> - Type-safe HTTP requests</li>
        <li><strong>Remote service pattern</strong> - Wrap remote APIs as services</li>
        <li><strong>Configuration-driven</strong> - Service URLs from config</li>
    </ul>

    <h2>Test Endpoints</h2>

    <div class="endpoint">
        <span class="method get">GET</span>
        <code>/users/:id</code> - Get user by ID (proxied to remote API)
    </div>

    <div class="endpoint">
        <span class="method post">POST</span>
        <code>/users</code> - Create new user (proxied to remote API)
    </div>

    <h2>📖 Documentation</h2>
    <p>See <code>index</code> for detailed explanation</p>
    <p>Use <code>test.http</code> for API testing</p>
    
    <p><strong>⚠ Note:</strong> This example requires a remote API server to be running.</p>
</body>
</html>`

	return response.NewHtmlResponse(html)
}

// ============================================
// Main
// ============================================

func main() {
	// Define remote service with configuration
	lokstra_registry.RegisterLazyService("remote-user-service", RemoteUserServiceFactory,
		map[string]any{
			"router":      "user-api",
			"path-prefix": "/api/v1",
		})

	// Setup router
	router := lokstra.NewRouter("remote-services")
	router.GET("/", Home)
	router.GET("/users/:id", GetUser)
	router.POST("/users", CreateUser)

	// Create and run app
	app := lokstra.NewApp("remote-services", ":3000", router)

	fmt.Println("🚀 Remote Services Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n📋 Available endpoints:")
	fmt.Println("   GET  /users/:id    - Get user (proxy)")
	fmt.Println("   POST /users        - Create user (proxy)")
	fmt.Println("\n⚠ Note: Configure remote API URL in service registration")
	fmt.Println("   or use ClientRouter with full URL")
	fmt.Println("\n🧪 Open test.http for API testing")

	if err := app.Run(0); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
