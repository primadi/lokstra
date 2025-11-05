package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"

	userApp "github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/application"
	userRepo "github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user/infrastructure/repository"

	orderApp "github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order/application"
	orderRepo "github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order/infrastructure/repository"
)

func registerServiceTypes() {
	// ==================== USER MODULE ====================
	// Register user repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		userRepo.UserRepositoryFactory, nil)

	// Register user service (application - local and remote)
	lokstra_registry.RegisterServiceType("user-service-factory",
		userApp.UserServiceFactory,
		nil, // Remote factory would go here for microservices
		deploy.WithResource("user", "users"),
		deploy.WithConvention("rest"),
	)

	// ==================== ORDER MODULE ====================
	// Register order repository (infrastructure - local only)
	lokstra_registry.RegisterServiceType("order-repository-factory",
		orderRepo.OrderRepositoryFactory, nil)

	// Register order service (application - local and remote)
	lokstra_registry.RegisterServiceType("order-service-factory",
		orderApp.OrderServiceFactory,
		nil, // Remote factory would go here for microservices
		deploy.WithResource("order", "orders"),
		deploy.WithConvention("rest"),
	)
}

func registerMiddlewareTypes() {
	// Register recovery middleware (built-in)
	recovery.Register()

	// Register custom middleware
	lokstra_registry.RegisterMiddlewareFactory("request-logger", requestLoggerFactory)
}

func requestLoggerFactory() request.HandlerFunc {
	return func(ctx *request.Context) error {
		// Before request
		reqID := uuid.New().String()
		log.Printf("→ [%s] %s %s", reqID, ctx.R.Method, ctx.R.URL.Path)
		ctx.Set("request_id", reqID)

		// Process request
		err := ctx.Next()

		// After request
		if err != nil {
			log.Printf("← [%s] ERROR: %v", reqID, err)
		} else {
			log.Printf("← [%s] SUCCESS (status: %d)", reqID, ctx.Resp.RespStatusCode)
		}

		return err
	}
}
