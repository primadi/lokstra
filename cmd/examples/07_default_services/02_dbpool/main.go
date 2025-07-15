package main

import (
	"context"
	"lokstra"
	"lokstra/services/dbpool_pg"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	ctx.RegisterServiceModule(dbpool_pg.GetModule())

	app := lokstra.NewApp(ctx, "dbpool-app", ":8080")

	app.GET("/users", func(ctx *lokstra.Context) error {
		service, err := ctx.GetService("dbpool_pg")
		if err != nil {
			return ctx.ErrorInternal("Database service not available")
		}

		dbService := service.(interface {
			Acquire(schema string) (interface{}, error)
		})

		conn, err := dbService.Acquire("public")
		if err != nil {
			return ctx.ErrorInternal("Failed to acquire database connection")
		}

		return ctx.Ok(map[string]any{
			"message": "Database connection acquired successfully",
			"status":  "connected",
		})
	})

	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{
			"status": "healthy",
			"service": "dbpool",
		})
	})

	lokstra.Logger.Infof("Database pool example started on :8080")
	app.Start()
}
