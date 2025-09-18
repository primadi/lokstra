package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/models"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/services"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ListUsers handles GET /users - list all users with pagination
func (h *UserHandler) ListUsers(ctx *lokstra.Context) error {
	// Parse query parameters
	listReq := &models.ListUsersRequest{
		Page:     1,
		PageSize: 10,
	}

	// Parse page
	if pageStr := ctx.GetQueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			listReq.Page = page
		}
	}

	// Parse page_size
	if pageSizeStr := ctx.GetQueryParam("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			listReq.PageSize = pageSize
		}
	}

	// Parse search
	listReq.Search = ctx.GetQueryParam("search")

	// Call service
	result, err := h.userService.ListUsers(ctx.Context, listReq)
	if err != nil {
		return ctx.ErrorInternal("Failed to list users: " + err.Error())
	}

	return ctx.Ok(result)
}

// GetUser handles GET /users/:id - get user by ID
func (h *UserHandler) GetUser(ctx *lokstra.Context) error {
	// Parse user ID from path parameter
	idStr := ctx.GetPathParam("id")
	if idStr == "" {
		return ctx.ErrorBadRequest("User ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.ErrorBadRequest("Invalid user ID format")
	}

	// Call service
	user, err := h.userService.GetUser(ctx.Context, id)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			return ctx.ErrorNotFound("User not found")
		}

		return ctx.ErrorInternal("Failed to get user: " + err.Error())
	}

	return ctx.Ok(&models.UserResponse{User: user})
}

// CreateUser handles POST /users - create new user
func (h *UserHandler) CreateUser(ctx *lokstra.Context) error {
	// Parse request body
	var createReq models.CreateUserRequest
	body, err := ctx.GetRawBody()
	if err != nil {
		return ctx.ErrorBadRequest("Failed to read request body")
	}

	if err := json.Unmarshal(body, &createReq); err != nil {
		return ctx.ErrorBadRequest("Invalid JSON in request body")
	}

	// Call service
	user, err := h.userService.CreateUser(ctx.Context, &createReq)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with email '%s' already exists", createReq.Email) {
			return ctx.ErrorDuplicate("User with this email already exists")
		}

		// Check if it's a validation error
		if validationErr, ok := err.(*models.ValidationError); ok {
			return ctx.ErrorValidation("Validation failed", map[string]string{
				validationErr.Field: validationErr.Message,
			})
		}

		return ctx.ErrorInternal("Failed to create user: " + err.Error())
	}

	return ctx.OkCreated(&models.UserResponse{User: user})
}

// UpdateUser handles PUT /users/:id - update user
func (h *UserHandler) UpdateUser(ctx *lokstra.Context) error {
	// Parse user ID from path parameter
	idStr := ctx.GetPathParam("id")
	if idStr == "" {
		return ctx.ErrorBadRequest("User ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.ErrorBadRequest("Invalid user ID format")
	}

	// Parse request body
	var updateReq models.UpdateUserRequest
	body, err := ctx.GetRawBody()
	if err != nil {
		return ctx.ErrorBadRequest("Failed to read request body")
	}

	if err := json.Unmarshal(body, &updateReq); err != nil {
		return ctx.ErrorBadRequest("Invalid JSON in request body")
	}

	// Call service
	user, err := h.userService.UpdateUser(ctx.Context, id, &updateReq)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			return ctx.ErrorNotFound("User not found")
		}

		if updateReq.Email != nil && err.Error() == fmt.Sprintf("user with email '%s' already exists", *updateReq.Email) {
			return ctx.ErrorDuplicate("User with this email already exists")
		}

		return ctx.ErrorInternal("Failed to update user: " + err.Error())
	}

	return ctx.OkUpdated(&models.UserResponse{User: user})
}

// DeleteUser handles DELETE /users/:id - delete user
func (h *UserHandler) DeleteUser(ctx *lokstra.Context) error {
	// Parse user ID from path parameter
	idStr := ctx.GetPathParam("id")
	if idStr == "" {
		return ctx.ErrorBadRequest("User ID is required")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.ErrorBadRequest("Invalid user ID format")
	}

	// Call service
	err = h.userService.DeleteUser(ctx.Context, id)
	if err != nil {
		if err.Error() == fmt.Sprintf("user with ID %d not found", id) {
			return ctx.ErrorNotFound("User not found")
		}

		return ctx.ErrorInternal("Failed to delete user: " + err.Error())
	}

	return ctx.Ok(map[string]any{
		"message": "User deleted successfully",
	})
}
