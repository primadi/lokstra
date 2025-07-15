package main

import (
	"lokstra"
	"lokstra/middleware/request_logger"
	"time"
)

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "request-logger-app", ":8080")

	ctx.RegisterMiddlewareModule(request_logger.GetModule())

	app.Use("lokstra.request_logger")

	app.GET("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"users": []string{"user1", "user2"},
		})
	})

	app.POST("/users", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "User created",
		})
	})

	app.GET("/slow", func(ctx *lokstra.Context) error {
		time.Sleep(200 * time.Millisecond)
		return ctx.Ok("Slow response")
	})

	lokstra.Logger.Infof("Request logger middleware example started on :8080")
	lokstra.Logger.Infof("All requests will be logged with method, path, status, and duration")
	app.Start()
}
