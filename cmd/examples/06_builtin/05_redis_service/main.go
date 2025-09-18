package main

import (
	"log"
	"time"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/services/redis"
)

// Request types for smart binding
type CacheRequest struct {
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

func main() {
	regCtx := lokstra.NewGlobalRegistrationContext()
	app := lokstra.NewApp(regCtx, "redis-service-app", ":8081")

	// Register the Redis service module
	regCtx.RegisterModule(redis.GetModule)

	// Basic Redis connection
	basicRedisConfig := "redis://localhost:6379/0"
	_, err := regCtx.CreateService("redis", "basic-redis", true, basicRedisConfig)
	if err != nil {
		log.Printf("Warning: Failed to create Redis service: %v", err)
	}

	// Health check
	app.GET("/health", func(ctx *lokstra.Context) error {
		return ctx.Ok(map[string]any{"status": "healthy"})
	})

	// Basic cache GET
	app.GET("/cache/:key", func(ctx *lokstra.Context) error {
		key := ctx.GetPathParam("key")

		redisService, err := serviceapi.GetService[serviceapi.Redis](regCtx, "basic-redis")
		if err != nil {
			return ctx.ErrorBadRequest("Redis service not available")
		}

		client := redisService.Client()
		value, err := client.Get(ctx.Request.Context(), key).Result()
		if err != nil {
			return ctx.ErrorNotFound("Key not found")
		}

		return ctx.Ok(map[string]any{"key": key, "value": value})
	})

	// Basic cache SET
	app.POST("/cache/:key", func(ctx *lokstra.Context, req *CacheRequest) error {
		key := ctx.GetPathParam("key")

		redisService, err := serviceapi.GetService[serviceapi.Redis](regCtx, "basic-redis")
		if err != nil {
			return ctx.ErrorBadRequest("Redis service not available")
		}

		client := redisService.Client()

		var expiration time.Duration
		if req.TTL > 0 {
			expiration = time.Duration(req.TTL) * time.Second
		}

		err = client.Set(ctx.Request.Context(), key, req.Value, expiration).Err()
		if err != nil {
			return ctx.ErrorInternal("Redis operation failed")
		}

		return ctx.Ok(map[string]any{"message": "Key set successfully", "key": key})
	})

	app.ListenAndServe()
}
