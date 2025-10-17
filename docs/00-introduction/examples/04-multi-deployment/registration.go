package main

import (
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/appservice"
	"github.com/primadi/lokstra/lokstra_registry"
)

// ========================================
// Service Registration - Manual Approach
// ========================================
//
// NOTE: This file demonstrates MANUAL service registration per deployment.
//
// In production, you would use automated patterns:
//   - Define services once with both local AND remote factories
//   - Config-driven deployment (YAML or code)
//   - Automatic selection of local vs remote based on server config
//   - See EVOLUTION.md for automated patterns
//
// Manual approach shown here for educational purposes:
//   - Understand deployment-specific registration
//   - Learn local vs remote factory selection
//   - See how service wiring differs per deployment
//   - Foundation before automation
//

func registerMonolithServices() {
	// register all service type
	lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
	lokstra_registry.RegisterServiceType("usersFactory", appservice.NewUserService)
	lokstra_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)

	// register lazy service for all services
	lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
	lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
	lokstra_registry.RegisterLazyService("orders", "ordersFactory", nil)
}

func registerUserServices() {
	// register only user-related service type
	lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
	lokstra_registry.RegisterServiceType("usersFactory", appservice.NewUserService)

	// register lazy service for user and its dependencies
	lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
	lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
}

func registerOrderServices() {
	// register only order-related service type
	lokstra_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
	lokstra_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)
	// register remote user service type
	lokstra_registry.RegisterServiceTypeRemote("usersFactory",
		appservice.NewUserServiceRemote)

	lokstra_registry.RegisterLazyService("db", "dbFactory", nil)
	lokstra_registry.RegisterLazyService("orders", "ordersFactory", nil)
	lokstra_registry.RegisterLazyService("users", "usersFactory", nil)
}
