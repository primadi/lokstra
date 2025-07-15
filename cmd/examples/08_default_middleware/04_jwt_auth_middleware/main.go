package main

import (
	"fmt"
	"lokstra"
	"lokstra/modules/jwt_auth_basic"
)

type SimpleValidator struct{}

func (s *SimpleValidator) ValidateCredentials(username, password string) (string, string, error) {
	if username == "admin" && password == "secret" {
		return "user123", "admin", nil
	}
	return "", "", fmt.Errorf("invalid credentials")
}

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "jwt-middleware-app", ":8080")

	jwtService, err := jwt_auth_basic.NewJWTAuthService("jwt_auth", map[string]any{
		"secret":        "my-secret-key",
		"expires_hours": 1,
	})
	if err != nil {
		panic(err)
	}

	jwtService.SetValidator(&SimpleValidator{})
	ctx.RegisterService("jwt_auth", jwtService)

	ctx.RegisterMiddlewareModule(jwt_auth_basic.GetMiddlewareModule())

	app.POST("/login", func(ctx *lokstra.Context) error {
		token, err := jwtService.Login("admin", "secret")
		if err != nil {
			return ctx.ErrorUnauthorized("Invalid credentials")
		}

		return ctx.Ok(map[string]any{
			"token": token,
			"type":  "Bearer",
		})
	})

	app.GET("/public", func(ctx *lokstra.Context) error {
		return ctx.Ok("This is a public endpoint")
	})

	protectedGroup := app.Group("/protected")
	protectedGroup.Use("lokstra.jwt_auth")

	protectedGroup.GET("/profile", func(ctx *lokstra.Context) error {
		userID := ctx.Get("user_id")
		roleID := ctx.Get("role_id")

		return ctx.Ok(map[string]any{
			"message": "Protected endpoint accessed",
			"user_id": userID,
			"role_id": roleID,
		})
	})

	lokstra.Logger.Infof("JWT Auth middleware example started on :8080")
	lokstra.Logger.Infof("1. POST /login to get token")
	lokstra.Logger.Infof("2. GET /protected/profile with Authorization: Bearer <token>")
	app.Start()
}
