package main

import (
	"github.com/primadi/lokstra/docs/00-introduction/examples_old/04-multi-deployment/appservice"
	"github.com/primadi/lokstra/old_registry"
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
	old_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
	old_registry.RegisterServiceType("usersFactory", appservice.NewUserService)
	old_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)

	// register lazy service for all services
	old_registry.RegisterLazyService("db", "dbFactory", nil)
	old_registry.RegisterLazyService("users", "usersFactory", nil)
	old_registry.RegisterLazyService("orders", "ordersFactory", nil)
}

func registerUserServices() {
	// register only user-related service type
	old_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
	old_registry.RegisterServiceType("usersFactory", appservice.NewUserService)

	// register lazy service for user and its dependencies
	old_registry.RegisterLazyService("db", "dbFactory", nil)
	old_registry.RegisterLazyService("users", "usersFactory", nil)
}

func registerOrderServices() {
	// register only order-related service type
	old_registry.RegisterServiceType("dbFactory", appservice.NewDatabase)
	old_registry.RegisterServiceType("ordersFactory", appservice.NewOrderService)
	// register remote user service type
	old_registry.RegisterServiceTypeRemote("usersFactory",
		appservice.NewUserServiceRemote)

	old_registry.RegisterLazyService("db", "dbFactory", nil)
	old_registry.RegisterLazyService("orders", "ordersFactory", nil)
	old_registry.RegisterLazyService("users", "usersFactory", nil)
}
