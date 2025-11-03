package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/repository"
	"github.com/primadi/lokstra/docs/00-introduction/examples/05-multi-deployment-pure-code/service"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
)

func registerServiceTypes() {
	// Clean Architecture: Separate layers for contract, model, service, repository

	// Register repositories (infrastructure layer - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.NewUserRepositoryMemory, nil)

	lokstra_registry.RegisterServiceType("order-repository-factory",
		repository.NewOrderRepositoryMemory, nil)

	// Register services (application layer - local and remote)
	// Metadata provided via RegisterServiceType options (not in factory structs)
	lokstra_registry.RegisterServiceType("user-service-factory",
		service.UserServiceFactory,
		service.UserServiceRemoteFactory,
		deploy.WithResource("user", "users"),
		deploy.WithConvention("rest"),
	)

	lokstra_registry.RegisterServiceType("order-service-factory",
		service.OrderServiceFactory,
		service.OrderServiceRemoteFactory,
		deploy.WithResource("order", "orders"),
		deploy.WithConvention("rest"),
		deploy.WithRouteOverride("GetByUserID", "/users/{user_id}/orders"),
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
