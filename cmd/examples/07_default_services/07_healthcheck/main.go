package main

import (
	"context"
	"lokstra"
	"lokstra/services/health"
	"time"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	ctx.RegisterServiceModule(health.GetModule())

	app := lokstra.NewApp(ctx, "health-app", ":8080")

	service, err := ctx.GetService("health")
	if err != nil {
		panic("Health service not available")
	}

	healthService := service.(*health.HealthService)

	healthService.RegisterCheck("database", func(ctx context.Context) error {
		return nil
	}, 5*time.Second)

	healthService.RegisterCheck("redis", func(ctx context.Context) error {
		return nil
	}, 3*time.Second)

	healthService.RegisterCheck("external_api", func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}, 2*time.Second)

	app.GET("/health", func(ctx *lokstra.Context) error {
		service, err := ctx.GetService("health")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		healthService := service.(*health.HealthService)
		
		results := healthService.CheckHealth(context.Background())
		
		allHealthy := true
		for _, status := range results {
			if status != health.StatusHealthy {
				allHealthy = false
				break
			}
		}

		statusCode := 200
		if !allHealthy {
			statusCode = 503
		}

		return ctx.JSON(statusCode, map[string]any{
			"status": func() string {
				if allHealthy {
					return "healthy"
				}
				return "unhealthy"
			}(),
			"checks": results,
		})
	})

	app.GET("/health/ready", func(ctx *lokstra.Context) error {
		service, err := ctx.GetService("health")
		if err != nil {
			return ctx.ErrorInternal("Health service not available")
		}

		healthService := service.(*health.HealthService)
		
		if healthService.IsHealthy(context.Background()) {
			return ctx.Ok(map[string]any{
				"status": "ready",
			})
		}

		return ctx.JSON(503, map[string]any{
			"status": "not ready",
		})
	})

	app.GET("/health/live", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status": "alive",
		})
	})

	lokstra.Logger.Infof("Health check example started on :8080")
	lokstra.Logger.Infof("Visit /health for detailed health status")
	lokstra.Logger.Infof("Visit /health/ready for readiness probe")
	lokstra.Logger.Infof("Visit /health/live for liveness probe")
	app.Start()
}
