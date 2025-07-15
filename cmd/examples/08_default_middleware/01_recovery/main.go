package main

import (
	"lokstra"
	"lokstra/middleware/recovery"
)

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "recovery-app", ":8080")

	ctx.RegisterMiddlewareModule(recovery.GetModule())

	app.Use("lokstra.recovery")

	app.GET("/safe", func(ctx *lokstra.Context) error {
		return ctx.Ok("This endpoint is safe")
	})

	app.GET("/panic", func(ctx *lokstra.Context) error {
		panic("This is a test panic!")
	})

	app.GET("/nil-pointer", func(ctx *lokstra.Context) error {
		var ptr *string
		return ctx.Ok(*ptr)
	})

	lokstra.Logger.Infof("Recovery middleware example started on :8080")
	lokstra.Logger.Infof("Try /safe (works), /panic (recovers), /nil-pointer (recovers)")
	app.Start()
}
