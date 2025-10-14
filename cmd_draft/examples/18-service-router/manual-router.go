package main

import (
	"fmt"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/router"
)

// CreateManualRouter demonstrates traditional manual router creation
// This is the old way of creating routes - manually registering each endpoint
func CreateManualRouter(service *UserService) router.Router {
	r := router.New("manual-user-router")
	r.SetPathPrefix("/api/v1/manual")

	// Manually register each route with wrapper handlers
	r.GET("/users", func(ctx *request.Context) error {
		users, err := service.ListUsers(ctx)
		if err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return err
		}
		ctx.Resp.Json(map[string]interface{}{
			"success": true,
			"data":    users,
		})
		return nil
	})

	r.GET("/users/{id}", func(ctx *request.Context) error {
		id := ctx.Req.PathParam("id", "")
		user, err := service.GetUser(ctx, id)
		if err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return err
		}
		ctx.Resp.Json(map[string]interface{}{
			"success": true,
			"data":    user,
		})
		return nil
	})

	r.POST("/users", func(ctx *request.Context) error {
		var req CreateUserRequest
		if err := ctx.Req.BindBody(&req); err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   "invalid request",
			})
			return err
		}

		user, err := service.CreateUser(ctx, &req)
		if err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return err
		}
		ctx.Resp.Json(map[string]interface{}{
			"success": true,
			"data":    user,
		})
		return nil
	})

	r.PUT("/users/{id}", func(ctx *request.Context) error {
		id := ctx.Req.PathParam("id", "")
		var req UpdateUserRequest
		if err := ctx.Req.BindBody(&req); err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   "invalid request",
			})
			return err
		}

		user, err := service.UpdateUser(ctx, id, &req)
		if err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return err
		}
		ctx.Resp.Json(map[string]interface{}{
			"success": true,
			"data":    user,
		})
		return nil
	})

	r.DELETE("/users/{id}", func(ctx *request.Context) error {
		id := ctx.Req.PathParam("id", "")
		err := service.DeleteUser(ctx, id)
		if err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return err
		}
		ctx.Resp.Json(map[string]interface{}{
			"success": true,
		})
		return nil
	})

	r.GET("/users/search", func(ctx *request.Context) error {
		users, err := service.SearchUsers(ctx)
		if err != nil {
			ctx.Resp.WithStatus(400).Json(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return err
		}
		ctx.Resp.Json(map[string]interface{}{
			"success": true,
			"data":    users,
		})
		return nil
	})

	fmt.Println("✓ Manual Router created successfully")
	return r
}
