package main

import (
	"context"
	"lokstra"
	"lokstra/services/redis"
	"time"
)

func main() {
	ctx := lokstra.NewGlobalContext()

	ctx.RegisterServiceModule(redis.GetModule())

	app := lokstra.NewApp(ctx, "redis-app", ":8080")

	app.POST("/cache/:key", func(ctx *lokstra.Context) error {
		key := ctx.Param("key")
		
		var body map[string]any
		if err := ctx.BindJSON(&body); err != nil {
			return ctx.ErrorBadRequest("Invalid JSON body")
		}

		value, ok := body["value"].(string)
		if !ok {
			return ctx.ErrorBadRequest("Value must be a string")
		}

		service, err := ctx.GetService("redis")
		if err != nil {
			return ctx.ErrorInternal("Redis service not available")
		}

		redisService := service.(*redis.RedisService)
		
		err = redisService.Set(context.Background(), key, value, 5*time.Minute)
		if err != nil {
			return ctx.ErrorInternal("Failed to set cache value")
		}

		return ctx.Ok(map[string]any{
			"message": "Value cached successfully",
			"key":     key,
		})
	})

	app.GET("/cache/:key", func(ctx *lokstra.Context) error {
		key := ctx.Param("key")

		service, err := ctx.GetService("redis")
		if err != nil {
			return ctx.ErrorInternal("Redis service not available")
		}

		redisService := service.(*redis.RedisService)
		
		value, err := redisService.Get(context.Background(), key)
		if err != nil {
			return ctx.ErrorNotFound("Key not found")
		}

		return ctx.Ok(map[string]any{
			"key":   key,
			"value": value,
		})
	})

	app.DELETE("/cache/:key", func(ctx *lokstra.Context) error {
		key := ctx.Param("key")

		service, err := ctx.GetService("redis")
		if err != nil {
			return ctx.ErrorInternal("Redis service not available")
		}

		redisService := service.(*redis.RedisService)
		
		err = redisService.Del(context.Background(), key)
		if err != nil {
			return ctx.ErrorInternal("Failed to delete cache value")
		}

		return ctx.Ok(map[string]any{
			"message": "Key deleted successfully",
			"key":     key,
		})
	})

	lokstra.Logger.Infof("Redis service example started on :8080")
	app.Start()
}
