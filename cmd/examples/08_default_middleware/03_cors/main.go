package main

import (
	"lokstra"
	"lokstra/middleware/cors"
)

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "cors-app", ":8080")

	ctx.RegisterMiddlewareModule(cors.GetModule())

	app.Use("lokstra.cors")

	app.GET("/api/data", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"data": "This endpoint supports CORS",
		})
	})

	app.POST("/api/submit", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Data submitted successfully",
		})
	})

	app.OPTIONS("/api/data", func(ctx *lokstra.Context) error {
		return ctx.Ok("")
	})

	lokstra.Logger.Infof("CORS middleware example started on :8080")
	lokstra.Logger.Infof("Try making cross-origin requests to /api/data and /api/submit")
	app.Start()
}
