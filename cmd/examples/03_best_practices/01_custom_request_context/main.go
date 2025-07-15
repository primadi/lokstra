package main

import (
	"lokstra"
)

type CustomContext struct {
	*lokstra.Context
	UserID   string
	TenantID string
}

func (c *CustomContext) GetUserID() string {
	return c.UserID
}

func (c *CustomContext) GetTenantID() string {
	return c.TenantID
}

func (c *CustomContext) IsAdmin() bool {
	roleID := c.Get("role_id")
	return roleID == "admin"
}

func customContextMiddleware(next lokstra.HandlerFunc) lokstra.HandlerFunc {
	return func(ctx *lokstra.Context) error {
		customCtx := &CustomContext{
			Context:  ctx,
			UserID:   "user123",
			TenantID: "tenant456",
		}

		return next(customCtx.Context)
	}
}

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "custom-context-app", ":8080")

	app.Use(customContextMiddleware)

	app.GET("/profile", func(ctx *lokstra.Context) error {
		customCtx := &CustomContext{Context: ctx}
		
		return ctx.Ok(map[string]any{
			"message":   "User profile",
			"user_id":   customCtx.GetUserID(),
			"tenant_id": customCtx.GetTenantID(),
			"is_admin":  customCtx.IsAdmin(),
		})
	})

	app.GET("/admin", func(ctx *lokstra.Context) error {
		customCtx := &CustomContext{Context: ctx}
		
		if !customCtx.IsAdmin() {
			return ctx.ErrorForbidden("Admin access required")
		}

		return ctx.Ok(map[string]any{
			"message": "Admin dashboard",
		})
	})

	lokstra.Logger.Infof("Custom context example started on :8080")
	app.Start()
}
