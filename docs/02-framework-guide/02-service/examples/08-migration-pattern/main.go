package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Demonstrates migration from monolith to microservices

// Step 1: Monolith Service
type MonolithService struct{}

func (*MonolithService) GetUser(id int) map[string]any {
	return map[string]any{"id": id, "name": "User from Monolith", "source": "monolith"}
}

func (*MonolithService) GetOrder(id int) map[string]any {
	return map[string]any{"id": id, "product": "Order from Monolith", "source": "monolith"}
}

func MonolithServiceFactory(deps map[string]any, config map[string]any) any {
	return &MonolithService{}
}

// Step 2: Extracted Microservice (User Service)
type UserMicroservice struct{}

func (*UserMicroservice) GetUser(id int) map[string]any {
	return map[string]any{"id": id, "name": "User from Microservice", "source": "microservice"}
}

func UserMicroserviceFactory(deps map[string]any, config map[string]any) any {
	return &UserMicroservice{}
}

// Step 3: Facade Pattern (gradual migration)
type FacadeService struct {
	useMonolith bool
	monolith    *MonolithService
	userService *UserMicroservice
}

func NewFacadeService(useMonolith bool, monolith *MonolithService, userService *UserMicroservice) *FacadeService {
	return &FacadeService{
		useMonolith: useMonolith,
		monolith:    monolith,
		userService: userService,
	}
}

func (s *FacadeService) GetUser(id int) map[string]any {
	if s.useMonolith {
		fmt.Println("→ Routing to monolith")
		return s.monolith.GetUser(id)
	}
	fmt.Println("→ Routing to microservice")
	return s.userService.GetUser(id)
}

func (s *FacadeService) GetOrder(id int) map[string]any {
	// Orders still in monolith
	return s.monolith.GetOrder(id)
}

func FacadeServiceFactory(deps map[string]any, config map[string]any) any {
	useMonolith := false
	if use, ok := config["use_monolith"].(bool); ok {
		useMonolith = use
	}

	monolith := lokstra_registry.GetService[*MonolithService]("monolith-service")
	userService := lokstra_registry.GetService[*UserMicroservice]("user-microservice")

	return NewFacadeService(useMonolith, monolith, userService)
}

// HTTP Handlers
type GetUserParams struct {
	ID int `path:"id"`
}

func GetUser(params *GetUserParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*FacadeService]("facade-service")
	result := svc.GetUser(params.ID)
	return response.NewApiOk(result)
}

func GetOrder(params *GetUserParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*FacadeService]("facade-service")
	result := svc.GetOrder(params.ID)
	return response.NewApiOk(result)
}

func Home() *response.Response {
	return response.NewHtmlResponse(`<!DOCTYPE html>
<html><head><title>Migration Pattern</title></head>
<body><h1>🔄 Migration Pattern Example</h1>
<p>Demonstrates gradual migration from monolith to microservices using facade pattern.</p>
<ul>
<li>GET /users/:id - Get user (can route to monolith or microservice)</li>
<li>GET /orders/:id - Get order (still in monolith)</li>
</ul>
<p><strong>Strategy:</strong> Use config flag to gradually shift traffic to microservice.</p>
</body></html>`)
}

func main() {
	lokstra_registry.RegisterServiceType("monolith-service", MonolithServiceFactory, nil)
	lokstra_registry.RegisterServiceType("user-microservice", UserMicroserviceFactory, nil)
	lokstra_registry.RegisterServiceType("facade-service", FacadeServiceFactory, nil)

	lokstra_registry.RegisterLazyService("monolith-service", MonolithServiceFactory, nil)
	lokstra_registry.RegisterLazyService("user-microservice", UserMicroserviceFactory, nil)

	// Toggle this to switch between monolith and microservice
	lokstra_registry.RegisterLazyService("facade-service", FacadeServiceFactory, map[string]any{
		"use_monolith": false, // Set to true to use monolith
	})

	router := lokstra.NewRouter("migration-pattern")
	router.GET("/", Home)
	router.GET("/users/:id", GetUser)
	router.GET("/orders/:id", GetOrder)

	app := lokstra.NewApp("migration-pattern", ":3000", router)
	fmt.Println("🚀 Migration Pattern Example")
	fmt.Println("📍 http://localhost:3000")
	fmt.Println("\n💡 Toggle 'use_monolith' config to switch routing")

	if err := app.Run(0); err != nil {
		fmt.Println("Error:", err)
	}
}
