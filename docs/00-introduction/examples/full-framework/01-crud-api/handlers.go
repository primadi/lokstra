package main

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
)

// ========================================
// Handler Struct with Dependency Injection
// ========================================

// UserHandler handles HTTP requests for user CRUD operations
type UserHandler struct {
	userService *service.Cached[*UserService]
}

// NewUserHandler creates a new UserHandler with injected service
func NewUserHandler(userService *service.Cached[*UserService]) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) listUsers(ctx *request.Context) error {
	users, err := h.userService.MustGet().GetAll()
	if err != nil {
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Ok(users)
}

func (h *UserHandler) getUser(ctx *request.Context) error {
	var params GetByIDParams
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	user, err := h.userService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}

	return ctx.Api.Ok(user)
}

func (h *UserHandler) createUser(ctx *request.Context) error {
	var params CreateParams
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request body")
	}

	user, err := h.userService.MustGet().Create(&params)
	if err != nil {
		if err.Error() == "email already exists" {
			return ctx.Api.Error(409, "DUPLICATE", err.Error())
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.Created(user, "User created successfully")
}

func (h *UserHandler) updateUser(ctx *request.Context) error {
	var params UpdateParams
	// Use BindAll to bind both path parameter AND body
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_INPUT", "Invalid request data")
	}

	user, err := h.userService.MustGet().Update(&params)
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

func (h *UserHandler) deleteUser(ctx *request.Context) error {
	var params DeleteParams
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	err := h.userService.MustGet().Delete(&params)
	if err != nil {
		if err.Error() == "failed to delete user: user not found" {
			return ctx.Api.Error(404, "NOT_FOUND", "User not found")
		}
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}

	return ctx.Api.OkWithMessage(nil, "User deleted successfully")
}
