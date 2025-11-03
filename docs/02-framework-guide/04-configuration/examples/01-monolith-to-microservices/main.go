package main

import (
	"fmt"
	"log"
	"os"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/config"
)

// User handlers
func GetUsersHandler() map[string]any {
	return map[string]any{
		"users": []map[string]any{
			{"id": 1, "name": "Alice"},
			{"id": 2, "name": "Bob"},
		},
	}
}

// Order handlers
func GetOrdersHandler() map[string]any {
	return map[string]any{
		"orders": []map[string]any{
			{"id": 1, "item": "Book", "user_id": 1},
			{"id": 2, "item": "Laptop", "user_id": 2},
		},
	}
}

// Home handler
func HomeHandler() string {
	return `
	<html>
	<body>
		<h1>Monolith to Microservices Example</h1>
		<p>Same codebase, different deployments</p>
		<h2>Monolith Deployment (Port 3010)</h2>
		<ul>
			<li><a href="http://localhost:3010/users" target="_blank">Users</a></li>
			<li><a href="http://localhost:3010/orders" target="_blank">Orders</a></li>
		</ul>
		<h2>Microservices Deployment</h2>
		<ul>
			<li><a href="http://localhost:3011/users" target="_blank">User Service (Port 3011)</a></li>
			<li><a href="http://localhost:3012/orders" target="_blank">Order Service (Port 3012)</a></li>
		</ul>
		<p>Note: Run different configs to test different deployments</p>
	</body>
	</html>
	`
}

func main() {
	// Determine which config to load
	deploymentType := "monolith" // Default
	if len(os.Args) > 1 {
		deploymentType = os.Args[1]
	}

	cfg := config.New()
	configFile := fmt.Sprintf("config-%s.yaml", deploymentType)

	log.Printf("Loading configuration: %s", configFile)
	if err := config.LoadConfigFile(configFile, cfg); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create routers based on deployment
	if deploymentType == "monolith" {
		router := lokstra.NewRouter("main")
		router.GET("/", HomeHandler)
		router.GET("/users", GetUsersHandler)
		router.GET("/orders", GetOrdersHandler)

		app := lokstra.NewApp("monolith", ":3010", router)
		log.Println("Starting monolith on :3010")
		if err := app.Run(0); err != nil {
			log.Fatal(err)
		}
	} else {
		// Microservices - start multiple servers
		userRouter := lokstra.NewRouter("user")
		userRouter.GET("/", HomeHandler)
		userRouter.GET("/users", GetUsersHandler)

		orderRouter := lokstra.NewRouter("order")
		orderRouter.GET("/", HomeHandler)
		orderRouter.GET("/orders", GetOrdersHandler)

		userApp := lokstra.NewApp("user-service", ":3011", userRouter)
		orderApp := lokstra.NewApp("order-service", ":3012", orderRouter)

		log.Println("Starting user service on :3011")
		log.Println("Starting order service on :3012")

		// Start services concurrently
		go func() {
			if err := userApp.Run(0); err != nil {
				log.Fatal(err)
			}
		}()

		if err := orderApp.Run(0); err != nil {
			log.Fatal(err)
		}
	}
}
