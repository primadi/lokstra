package main

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/docs/00-introduction/examples_old/04-multi-deployment/appservice"
)

// ========================================
// Handlers - Manual Approach
// ========================================
//
// NOTE: This example demonstrates MANUAL handler creation from service methods.
//
// In production, you would use automated patterns:
//   - router.NewFromService() to auto-generate handlers
//   - Convention-based routing (RESTful, RPC, etc)
//   - See EVOLUTION.md for automated patterns
//
// Manual approach shown here for educational purposes:
//   - Understand how service-to-handler works
//   - Learn request binding and error handling
//   - See the foundation before automation
//

var (
	userService  = service.LazyLoad[appservice.UserService]("users")
	orderService = service.LazyLoad[appservice.OrderService]("orders")
)

func listUsersHandler(ctx *request.Context) error {
	users, err := userService.MustGet().List(&appservice.ListUsersParams{})
	if err != nil {
		return ctx.Api.Error(500, "INTERNAL_ERROR", err.Error())
	}
	return ctx.Api.Ok(users)
}

func getUserHandler(ctx *request.Context) error {
	var params appservice.GetUserParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	user, err := userService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(user)
}

func getOrderHandler(ctx *request.Context) error {
	var params appservice.GetOrderParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid order ID")
	}

	orderWithUser, err := orderService.MustGet().GetByID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(orderWithUser)
}

func getUserOrdersHandler(ctx *request.Context) error {
	var params appservice.GetUserOrdersParams
	if err := ctx.Req.BindPath(&params); err != nil {
		return ctx.Api.BadRequest("INVALID_ID", "Invalid user ID")
	}

	orders, err := orderService.MustGet().GetByUserID(&params)
	if err != nil {
		return ctx.Api.Error(404, "NOT_FOUND", err.Error())
	}
	return ctx.Api.Ok(orders)
}
