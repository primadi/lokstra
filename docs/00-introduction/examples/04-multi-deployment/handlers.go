package main

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples/04-multi-deployment/appservice"
)

// ========================================
// User Handler (struct-based DI)
// ========================================

type UserHandler struct {
	userService *service.Cached[appservice.UserService]
}

func NewUserHandler(userService *service.Cached[appservice.UserService]) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) list(ctx *request.Context) error {
	users, err := h.userService.MustGet().List(&appservice.ListUsersParams{})
	if err != nil {
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}
	return ctx.Api.Ok(users)
}

func (h *UserHandler) get(ctx *request.Context) error {
	var params appservice.GetUserParams
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	user, err := h.userService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(user)
}

// ========================================
// Order Handler (struct-based DI)
// ========================================

type OrderHandler struct {
	orderService *service.Cached[appservice.OrderService]
}

func NewOrderHandler(orderService *service.Cached[appservice.OrderService]) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) get(ctx *request.Context) error {
	var params appservice.GetOrderParams
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid order ID")
	}

	orderWithUser, err := h.orderService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(orderWithUser)
}

func (h *OrderHandler) getUserOrders(ctx *request.Context) error {
	var params appservice.GetUserOrdersParams
	if err := ctx.Req.BindAll(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	orders, err := h.orderService.MustGet().GetByUserID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(orders)
}
