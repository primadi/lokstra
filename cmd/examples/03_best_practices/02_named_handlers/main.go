package main

import (
	"lokstra"
)

func main() {
	ctx := lokstra.NewGlobalContext()
	app := lokstra.NewApp(ctx, "named-handlers-app", ":8080")

	registerHandlers(ctx)

	app.GET("/users", "user.list")
	app.GET("/users/:id", "user.get")
	app.POST("/users", "user.create")
	app.PUT("/users/:id", "user.update")
	app.DELETE("/users/:id", "user.delete")

	app.GET("/products", "product.list")
	app.GET("/products/:id", "product.get")

	lokstra.Logger.Infof("Named handlers example started on :8080")
	app.Start()
}

func registerHandlers(ctx *lokstra.GlobalContext) {
	ctx.RegisterHandler("user.list", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "List all users",
			"users":   []string{"user1", "user2", "user3"},
		})
	})

	ctx.RegisterHandler("user.get", func(ctx *lokstra.Context) error {
		userID := ctx.Param("id")
		return ctx.Ok(map[string]any{
			"message": "Get user by ID",
			"user_id": userID,
			"user":    map[string]any{"id": userID, "name": "John Doe"},
		})
	})

	ctx.RegisterHandler("user.create", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message": "Create new user",
			"status":  "created",
		})
	})

	ctx.RegisterHandler("user.update", func(ctx *lokstra.Context) error {
		userID := ctx.Param("id")
		return ctx.Ok(map[string]any{
			"message": "Update user",
			"user_id": userID,
			"status":  "updated",
		})
	})

	ctx.RegisterHandler("user.delete", func(ctx *lokstra.Context) error {
		userID := ctx.Param("id")
		return ctx.Ok(map[string]any{
			"message": "Delete user",
			"user_id": userID,
			"status":  "deleted",
		})
	})

	ctx.RegisterHandler("product.list", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"message":  "List all products",
			"products": []string{"product1", "product2", "product3"},
		})
	})

	ctx.RegisterHandler("product.get", func(ctx *lokstra.Context) error {
		productID := ctx.Param("id")
		return ctx.Ok(map[string]any{
			"message":    "Get product by ID",
			"product_id": productID,
			"product":    map[string]any{"id": productID, "name": "Sample Product"},
		})
	})
}
