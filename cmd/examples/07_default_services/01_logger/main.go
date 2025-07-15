package main

import (
	"lokstra"
	"lokstra/services/logger"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	ctx.RegisterServiceModule(logger.GetModule())

	app := lokstra.NewApp(ctx, "logger-app", ":8080")

	app.GET("/debug", func(ctx *lokstra.Context) error {
		lokstra.Logger.Debugf("Debug message from handler")
		return ctx.Ok("Debug message logged")
	})

	app.GET("/info", func(ctx *lokstra.Context) error {
		lokstra.Logger.Infof("Info message from handler")
		return ctx.Ok("Info message logged")
	})

	app.GET("/warn", func(ctx *lokstra.Context) error {
		lokstra.Logger.Warnf("Warning message from handler")
		return ctx.Ok("Warning message logged")
	})

	app.GET("/error", func(ctx *lokstra.Context) error {
		lokstra.Logger.Errorf("Error message from handler")
		return ctx.Ok("Error message logged")
	})

	app.GET("/with-fields", func(ctx *lokstra.Context) error {
		lokstra.Logger.WithField("user_id", "123").
			WithField("action", "test").
			Infof("Action performed by user")
		return ctx.Ok("Message with fields logged")
	})

	lokstra.Logger.Infof("Logger service example started on :8080")
	app.Start()
}
