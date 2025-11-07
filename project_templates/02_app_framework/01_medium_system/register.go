package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/core/deploy"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/repository"
	"github.com/primadi/lokstra/project_templates/02_app_framework/01_medium_system/service"
)

func registerServiceTypes() {
	// Register repositories (infrastructure layer - local only)
	lokstra_registry.RegisterServiceType("user-repository-factory",
		repository.NewUserRepositoryMemory, nil)

	lokstra_registry.RegisterServiceType("order-repository-factory",
		repository.NewOrderRepositoryMemory, nil)

	// Register services (application layer - local and remote)
	lokstra_registry.RegisterServiceType("user-service-factory",
		service.UserServiceFactory,
		service.UserServiceRemoteFactory,
		deploy.WithRouter(&deploy.ServiceTypeRouter{
			PathPrefix:  "/api",
			Middlewares: []string{"recovery", "request-logger"},
		}),
	)

	lokstra_registry.RegisterServiceType("order-service-factory",
		service.OrderServiceFactory,
		service.OrderServiceRemoteFactory,
		deploy.WithRouter(&deploy.ServiceTypeRouter{
			CustomRoutes: map[string]string{
				"GetByUserID":  "/users/{user_id}/orders",
				"UpdateStatus": "PATCH /orders/{id}/status",
			},
		}),
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
