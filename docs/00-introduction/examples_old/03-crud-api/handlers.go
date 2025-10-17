package main

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// ========================================
// Custom Handlers (Alternative to Service-as-Router)
// ========================================

// Package-level cached service for optimal performance
// - Loaded once on first access (lazy initialization)
// - Cached for all subsequent calls (zero registry lookup cost)
// - Thread-safe via sync.Once
// - MustGet() panics with clear error if service not found (fail-fast)
var userService = service.LazyLoad[*UserService]("users")

func listUsersHandler(ctx *request.Context) error {
	users, err := userService.MustGet().GetAll()
	if err != nil {
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
	var params GetByIDParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	user, err := userService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}

	return ctx.Api.Ok(user)
}

func createUserHandler(ctx *request.Context) error {
	var params CreateParams
	if err := ctx.Req.BindBody(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request body")
	}

	user, err := userService.MustGet().Create(&params)
	if err != nil {
		if err.Error() == "email already exists" {
			return ctx.Api.Error(409, "DUPLICATE", err.Error())
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Created(user, "User created successfully")
}

func updateUserHandler(ctx *request.Context) error {
	var params UpdateParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}
	if err := ctx.Req.BindBody(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request body")
	}

	user, err := userService.MustGet().Update(&params)
	if err != nil {
		if err.Error() == "user not found" {
			return ctx.Api.Error(404, "NOT_FOUND", err.Error())
		}
		if err.Error() == "email already exists" {
			return ctx.Api.Error(409, "DUPLICATE", err.Error())
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Ok(user)
}

func deleteUserHandler(ctx *request.Context) error {
	var params DeleteParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	err := userService.MustGet().Delete(&params)
	if err != nil {
		if err.Error() == "failed to delete user: user not found" {
			return ctx.Api.Error(404, "NOT_FOUND", "User not found")
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.OkWithMessage(nil, "User deleted successfully")
}
