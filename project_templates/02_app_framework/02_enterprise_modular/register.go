package main

import (
	"github.com/google/uuid"
	"github.com/primadi/lokstra/common/logger"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/lokstra_registry"
	"github.com/primadi/lokstra/middleware/recovery"

	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/order"
	"github.com/primadi/lokstra/project_templates/02_app_framework/02_enterprise_modular/modules/user"
)

func registerServiceTypes() {
	// Register modules (each module owns its service type registration)
	user.Register()
	order.Register()
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
		logger.LogInfo("→ [%s] %s %s", reqID, ctx.R.Method, ctx.R.URL.Path)
		ctx.Set("request_id", reqID)

		// Process request
		err := ctx.Next()

		// After request
		if err != nil {
			logger.LogError("← [%s] ERROR: %v", reqID, err)
		} else {
			logger.LogInfo("← [%s] SUCCESS (status: %d)", reqID, ctx.Resp.RespStatusCode)
		}

		return err
	}
}
