package main

import (
	"lokstra"
	"lokstra/services/metrics"
	"net/http"
	"time"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	ctx.RegisterServiceModule(metrics.GetModule())

	app := lokstra.NewApp(ctx, "metrics-app", ":8080")

	app.GET("/metrics", func(ctx *lokstra.Context) error {
		service, err := ctx.GetService("metrics")
		if err != nil {
			return ctx.ErrorInternal("Metrics service not available")
		}

		metricsService := service.(*metrics.MetricsService)
		handler := metricsService.GetHandler()
		
		handler.ServeHTTP(ctx.ResponseWriter, ctx.Request)
		return nil
	})

	app.GET("/test", func(ctx *lokstra.Context) error {
		service, err := ctx.GetService("metrics")
		if err != nil {
			return ctx.ErrorInternal("Metrics service not available")
		}

		metricsService := service.(*metrics.MetricsService)
		
		start := time.Now()
		
		time.Sleep(100 * time.Millisecond)
		
		duration := time.Since(start)
		metricsService.RecordHTTPRequest("GET", "/test", "200", duration)

		return ctx.Ok(map[string]any{
			"message":  "Test endpoint",
			"duration": duration.String(),
		})
	})

	app.GET("/slow", func(ctx *lokstra.Context) error {
		service, err := ctx.GetService("metrics")
		if err != nil {
			return ctx.ErrorInternal("Metrics service not available")
		}

		metricsService := service.(*metrics.MetricsService)
		
		start := time.Now()
		
		time.Sleep(500 * time.Millisecond)
		
		duration := time.Since(start)
		metricsService.RecordHTTPRequest("GET", "/slow", "200", duration)

		return ctx.Ok(map[string]any{
			"message":  "Slow endpoint",
			"duration": duration.String(),
		})
	})

	lokstra.Logger.Infof("Metrics service example started on :8080")
	lokstra.Logger.Infof("Visit /metrics to see Prometheus metrics")
	lokstra.Logger.Infof("Visit /test and /slow to generate metrics")
	app.Start()
}
