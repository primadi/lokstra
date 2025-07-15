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
	if username == "user" && password == "password" {
		return "user456", "user", nil
	}
	return "", "", fmt.Errorf("invalid credentials")
}

func main() {
	ctx := lokstra.NewGlobalContext()

	jwtService, err := jwt_auth_basic.NewJWTAuthService("jwt_auth", map[string]any{
		"secret":        "my-secret-key",
		"expires_hours": 24,
	})
	if err != nil {
		panic(err)
	}

	jwtService.SetValidator(&SimpleValidator{})
	ctx.RegisterService("jwt_auth", jwtService)

	ctx.RegisterMiddlewareModule(jwt_auth_basic.GetMiddlewareModule())

	app := lokstra.NewApp(ctx, "jwt-app", ":8080")

	app.POST("/login", func(ctx *lokstra.Context) error {
		var body map[string]any
		if err := ctx.BindJSON(&body); err != nil {
			return ctx.ErrorBadRequest("Invalid JSON body")
		}

		username, ok := body["username"].(string)
		if !ok {
			return ctx.ErrorBadRequest("Username is required")
		}

		password, ok := body["password"].(string)
		if !ok {
			return ctx.ErrorBadRequest("Password is required")
		}

		service, err := ctx.GetService("jwt_auth")
		if err != nil {
			return ctx.ErrorInternal("JWT service not available")
		}

		jwtService := service.(*jwt_auth_basic.JWTAuthService)
		
		token, err := jwtService.Login(username, password)
		if err != nil {
			return ctx.ErrorUnauthorized("Invalid credentials")
		}

		return ctx.Ok(map[string]any{
			"token": token,
			"type":  "Bearer",
		})
	})

	protectedGroup := app.Group("/protected", "lokstra.jwt_auth")

	protectedGroup.GET("/profile", func(ctx *lokstra.Context) error {
		userID := ctx.Get("user_id")
		roleID := ctx.Get("role_id")

		return ctx.Ok(map[string]any{
			"user_id": userID,
			"role_id": roleID,
			"message": "This is a protected endpoint",
		})
	})

	protectedGroup.GET("/admin", func(ctx *lokstra.Context) error {
		roleID := ctx.Get("role_id")
		if roleID != "admin" {
			return ctx.ErrorForbidden("Admin access required")
		}

		return ctx.Ok(map[string]any{
			"message": "Admin dashboard",
		})
	})

	lokstra.Logger.Infof("JWT Auth example started on :8080")
	lokstra.Logger.Infof("Try: POST /login with {\"username\":\"admin\",\"password\":\"secret\"}")
	lokstra.Logger.Infof("Then: GET /protected/profile with Authorization: Bearer <token>")
	app.Start()
}
