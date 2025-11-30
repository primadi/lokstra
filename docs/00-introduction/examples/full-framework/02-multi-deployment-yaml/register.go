package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/repository"
	"github.com/primadi/lokstra/docs/00-introduction/examples/full-framework/02-multi-deployment-yaml/service"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
)

func registerServiceTypes() {
	// Clean Architecture: Separate layers for contract, model, service, repository

	// Register repositories (infrastructure layer - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.NewUserRepositoryMemory)

	lokstra_registry.RegisterServiceType("order-repository-factory",
		repository.NewOrderRepositoryMemory)

	// Register services (application layer - local and remote)
	// Metadata provided via RegisterServiceType options (not in factory structs)
	lokstra_registry.RegisterRouterServiceType("user-service-factory",
		service.UserServiceFactory,
		service.UserServiceRemoteFactory,
		nil, // No custom config - uses default REST routing
	)

	lokstra_registry.RegisterRouterServiceType("order-service-factory",
		service.OrderServiceFactory,
		service.OrderServiceRemoteFactory,
		&deploy.ServiceTypeConfig{
			RouteOverrides: map[string]deploy.RouteConfig{
				"GetByUserID": {Path: "/users/{user_id}/orders"},
			},
		},
	)
}

func registerMiddlewareTypes() {
	// Register recovery middleware (built-in middleware)
	recovery.Register()

	// register custom middleware
	lokstra_registry.RegisterMiddlewareFactory("before-after-logger", beforeAfterLoggerFactory)
}

func beforeAfterLoggerFactory() request.HandlerFunc {
	return func(ctx *request.Context) error {
		// 1. Before request
		reqId := uuid.New().String()
		log.Printf("Starting request id=%s %s %s\n", reqId, ctx.R.Method, ctx.R.URL.Path)
		ctx.Set("request_id", reqId)

		// 2. Proceed to next middleware/handler
		err := ctx.Next()

		// 3. After request
		if err != nil {
			log.Printf("Request id=%s failed: %v\n", reqId, err)
		} else {
			log.Printf("Request id=%s completed successfully\n", reqId)
		}
		return err
	}
}
