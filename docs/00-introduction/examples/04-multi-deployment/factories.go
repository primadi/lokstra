package main

import (
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/appservice"
)

// ========================================
// Service Factories (for deploy.Global)
// ========================================

// DatabaseFactory creates a new database instance
func DatabaseFactory(deps map[string]any, config map[string]any) any {
	return appservice.NewDatabase()
}

// UserServiceFactory creates a new user service instance
func UserServiceFactory(deps map[string]any, config map[string]any) any {
	return &appservice.UserServiceImpl{
		DB: service.Cast[*appservice.Database](deps["database"]),
	}
}

// OrderServiceFactory creates a new order service instance
func OrderServiceFactory(deps map[string]any, config map[string]any) any {
	return &appservice.OrderServiceImpl{
		DB:    service.Cast[*appservice.Database](deps["database"]),
		Users: service.Cast[appservice.UserService](deps["user-service"]),
	}
}

// UserServiceRemoteFactory creates a remote user service client
// func UserServiceRemoteFactory(deps map[string]any, config map[string]any) any {
// 	// TODO: Get base URL from remote service config
// 	baseURL := "http://localhost:3004"
// 	return appservice.NewUserServiceRemote(baseURL)
// }
