package main

import (
	"lokstra"
	"lokstra/middleware/ratelimit"
	"lokstra/middleware/requestid"
	"lokstra/middleware/security"
	"lokstra/middleware/timeout"
	"time"
)

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "custom-middleware-app", ":8080")

	ctx.RegisterMiddlewareModule(requestid.GetModule())
	ctx.RegisterMiddlewareModule(security.GetModule())
	ctx.RegisterMiddlewareModule(ratelimit.GetModule())
	ctx.RegisterMiddlewareModule(timeout.GetModule())

	app.Use("lokstra.requestid")
	app.Use("lokstra.security")
	app.Use("lokstra.ratelimit")
	app.Use("lokstra.timeout")

	app.GET("/test", func(ctx *lokstra.Context) error {
		requestID := ctx.Get("request_id")
		
		return ctx.Ok(map[string]any{
			"message":    "Request processed successfully",
			"request_id": requestID,
		})
	})

	app.GET("/slow", func(ctx *lokstra.Context) error {
		time.Sleep(2 * time.Second)
		return ctx.Ok("This took 2 seconds")
	})

	app.GET("/rate-test", func(ctx *lokstra.Context) error {
		return ctx.Ok("Rate limit test - try calling this rapidly")
	})

	lokstra.Logger.Infof("Custom middleware example started on :8080")
	lokstra.Logger.Infof("Middleware stack: Request ID -> Security Headers -> Rate Limit -> Timeout")
	lokstra.Logger.Infof("Try /test, /slow, and /rate-test endpoints")
	app.Start()
}
