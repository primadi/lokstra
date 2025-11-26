package main

import (
	"log"

	"github.com/google/uuid"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"

	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/order"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user"
)

func registerServiceTypes() {
	// Register modules (each module owns its service type registration)
	user.Register()
	order.Register()
}

func registerRouters() {
	// Register manual routers (not generated from @RouterService)
	healthRouter := NewHealthRouter()
	lokstra_registry.RegisterRouter("health-router", healthRouter)
	log.Println("✅ Registered manual router: health-router")
}

func registerMiddlewareTypes() {
	// Register recovery middleware (built-in)
	recovery.Register()

	// Register custom middleware
	lokstra_registry.RegisterMiddlewareFactory("request-logger", requestLoggerFactory)
	lokstra_registry.RegisterMiddlewareFactory("mw-test", func(config map[string]any) request.HandlerFunc {
		return func(ctx *request.Context) error {
			log.Printf("→ [mw-test] Before request | Param1: %v, Param2: %v", config["param1"], config["param2"])
			err := ctx.Next()
			log.Println("← [mw-test] After request")
			return err
		}
	})
}

func requestLoggerFactory(config map[string]any) request.HandlerFunc {
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
