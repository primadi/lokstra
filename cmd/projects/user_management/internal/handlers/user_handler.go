package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/primadi/lokstra/cmd/projects/user_management/internal/models"
	"github.com/primadi/lokstra/cmd/projects/user_management/internal/repository"
	"github.com/primadi/lokstra/core/dsl"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
	"github.com/primadi/lokstra/serviceapi/auth"
)

// UserHandler handles user-related HTTP requests using DSL
type UserHandler struct {
	serviceVars *dsl.ServiceVar[UserHandlerParams]
	userRepo    auth.UserRepository
}

// UserHandlerParams contains parameters for handler operations
type UserHandlerParams struct {
	RequestBody []byte
	TenantID    string
	UserID      string
	Username    string

	// Request/Response objects
	CreateUserReq *models.CreateUserRequest
	UpdateUserReq *models.UpdateUserRequest
	ListUsersReq  *models.ListUsersRequest
	User          *auth.User
	Users         []*auth.User

	// HTTP specific
	StatusCode   int
	ResponseBody []byte
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	dbPool serviceapi.DbPool,
	logger serviceapi.Logger,
	metrics serviceapi.Metrics,
	i18n serviceapi.I18n,
	userRepo auth.UserRepository,
) *UserHandler {
	return &UserHandler{
		serviceVars: dsl.NewServiceVar(
			dbPool,
			"public",
			logger,
			metrics,
			i18n,
			&UserHandlerParams{StatusCode: 200},
			make(map[string]any),
		),
		userRepo: userRepo,
	}
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// Create request context
	reqCtx := &request.Context{Context: r.Context()}

	// Extract tenant ID from header or path
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default" // Default tenant
	}
	h.serviceVars.Param.TenantID = tenantID

	// Create DSL flow
	flow := dsl.NewFlow("CreateUser", h.serviceVars)

	// Read and parse request body
	flow.Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
		buf := make([]byte, r.ContentLength)
		_, err := r.Body.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return dsl.ErrValidationFailed("request_body", err.Error())
		}
		ctx.GetParam().RequestBody = buf
		return nil
	}).

		// Parse JSON request
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			var createReq models.CreateUserRequest
			if err := json.Unmarshal(params.RequestBody, &createReq); err != nil {
				return dsl.ErrValidationFailed("json_format", err.Error())
			}
			params.CreateUserReq = &createReq
			return nil
		}).

		// Validate request
		Validate(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			req := ctx.GetParam().CreateUserReq
			if err := req.Validate(); err != nil {
				return err
			}
			return nil
		}).

		// Convert to auth.User
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			req := params.CreateUserReq

			user := &auth.User{
				TenantID:     params.TenantID,
				Username:     req.Username,
				Email:        req.Email,
				PasswordHash: h.hashPassword(req.Password),
				IsActive:     true,
				Metadata:     req.Metadata,
			}
			params.User = user
			return nil
		}).

		// Create user via repository
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			user := ctx.GetParam().User
			if err := h.userRepo.CreateUser(ctx.GetContext().Context, user); err != nil {
				return err
			}
			return nil
		}).

		// Prepare response
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			user := params.User

			response := models.UserResponse{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				IsActive:  user.IsActive,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Metadata:  user.Metadata,
			}

			respBytes, err := json.Marshal(response)
			if err != nil {
				return dsl.ErrDatabaseOperation("json_marshal", err)
			}

			params.ResponseBody = respBytes
			params.StatusCode = 201
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.serviceVars.Param.StatusCode)
	w.Write(h.serviceVars.Param.ResponseBody)
}

// GetUser handles GET /users/{username}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter required", http.StatusBadRequest)
		return
	}

	h.serviceVars.Param.TenantID = tenantID
	h.serviceVars.Param.Username = username

	// Create request context
	reqCtx := &request.Context{Context: r.Context()}

	// Create DSL flow
	flow := dsl.NewFlow("GetUser", h.serviceVars)

	// Get user from repository
	flow.Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
		params := ctx.GetParam()
		user, err := h.userRepo.GetUserByName(ctx.GetContext().Context, params.TenantID, params.Username)
		if err != nil {
			return err
		}
		if user == nil {
			return dsl.ErrNotFound("user", params.Username)
		}
		params.User = user
		return nil
	}).

		// Prepare response
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			user := params.User

			response := models.UserResponse{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				IsActive:  user.IsActive,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				LastLogin: user.LastLogin,
				Metadata:  user.Metadata,
			}

			respBytes, err := json.Marshal(response)
			if err != nil {
				return dsl.ErrDatabaseOperation("json_marshal", err)
			}

			params.ResponseBody = respBytes
			params.StatusCode = 200
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.serviceVars.Param.StatusCode)
	w.Write(h.serviceVars.Param.ResponseBody)
}

// UpdateUser handles PUT /users/{username}
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter required", http.StatusBadRequest)
		return
	}

	h.serviceVars.Param.TenantID = tenantID
	h.serviceVars.Param.Username = username

	// Create request context
	reqCtx := &request.Context{Context: r.Context()}

	// Create DSL flow
	flow := dsl.NewFlow("UpdateUser", h.serviceVars)

	// Read and parse request body
	flow.Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
		buf := make([]byte, r.ContentLength)
		_, err := r.Body.Read(buf)
		if err != nil && err.Error() != "EOF" {
			return dsl.ErrValidationFailed("request_body", err.Error())
		}
		ctx.GetParam().RequestBody = buf
		return nil
	}).

		// Parse JSON request
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			var updateReq models.UpdateUserRequest
			if err := json.Unmarshal(params.RequestBody, &updateReq); err != nil {
				return dsl.ErrValidationFailed("json_format", err.Error())
			}
			params.UpdateUserReq = &updateReq
			return nil
		}).

		// Get existing user
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			user, err := h.userRepo.GetUserByName(ctx.GetContext().Context, params.TenantID, params.Username)
			if err != nil {
				return err
			}
			if user == nil {
				return dsl.ErrNotFound("user", params.Username)
			}
			params.User = user
			return nil
		}).

		// Update user fields
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			user := params.User
			req := params.UpdateUserReq

			// Update only provided fields
			if req.Email != nil {
				user.Email = *req.Email
			}
			if req.Password != nil {
				user.PasswordHash = h.hashPassword(*req.Password)
			}
			if req.IsActive != nil {
				user.IsActive = *req.IsActive
			}
			if req.Metadata != nil {
				user.Metadata = req.Metadata
			}

			return nil
		}).

		// Update user via repository
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			user := ctx.GetParam().User
			return h.userRepo.UpdateUser(ctx.GetContext().Context, user)
		}).

		// Prepare response
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			user := params.User

			response := models.UserResponse{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				IsActive:  user.IsActive,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				LastLogin: user.LastLogin,
				Metadata:  user.Metadata,
			}

			respBytes, err := json.Marshal(response)
			if err != nil {
				return dsl.ErrDatabaseOperation("json_marshal", err)
			}

			params.ResponseBody = respBytes
			params.StatusCode = 200
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.serviceVars.Param.StatusCode)
	w.Write(h.serviceVars.Param.ResponseBody)
}

// DeleteUser handles DELETE /users/{username}
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Extract parameters
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username parameter required", http.StatusBadRequest)
		return
	}

	h.serviceVars.Param.TenantID = tenantID
	h.serviceVars.Param.Username = username

	// Create request context
	reqCtx := &request.Context{Context: r.Context()}

	// Create DSL flow
	flow := dsl.NewFlow("DeleteUser", h.serviceVars)

	// Delete user via repository
	flow.Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
		params := ctx.GetParam()
		return h.userRepo.DeleteUser(ctx.GetContext().Context, params.TenantID, params.Username)
	}).

		// Prepare success response
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()

			response := map[string]any{
				"message": "User deleted successfully",
				"status":  "success",
			}

			respBytes, err := json.Marshal(response)
			if err != nil {
				return dsl.ErrDatabaseOperation("json_marshal", err)
			}

			params.ResponseBody = respBytes
			params.StatusCode = 200
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.serviceVars.Param.StatusCode)
	w.Write(h.serviceVars.Param.ResponseBody)
}

// ListUsers handles GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Extract parameters
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	h.serviceVars.Param.TenantID = tenantID
	h.serviceVars.Param.ListUsersReq = &models.ListUsersRequest{
		Limit:  limit,
		Offset: offset,
	}

	// Create request context
	reqCtx := &request.Context{Context: r.Context()}

	// Create DSL flow
	flow := dsl.NewFlow("ListUsers", h.serviceVars)

	// Get users from repository
	flow.Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
		params := ctx.GetParam()
		users, err := h.userRepo.ListUsers(ctx.GetContext().Context, params.TenantID)
		if err != nil {
			return err
		}
		params.Users = users
		return nil
	}).

		// Apply pagination and prepare response
		Do(func(ctx *dsl.FlowContext[UserHandlerParams]) error {
			params := ctx.GetParam()
			users := params.Users
			req := params.ListUsersReq

			// Apply offset and limit
			total := len(users)
			start := req.Offset
			end := start + req.Limit

			if start >= total {
				users = []*auth.User{}
			} else {
				if end > total {
					end = total
				}
				users = users[start:end]
			}

			// Convert to response format
			userResponses := make([]models.UserResponse, len(users))
			for i, user := range users {
				userResponses[i] = models.UserResponse{
					ID:        user.ID,
					Username:  user.Username,
					Email:     user.Email,
					IsActive:  user.IsActive,
					CreatedAt: user.CreatedAt,
					UpdatedAt: user.UpdatedAt,
					LastLogin: user.LastLogin,
					Metadata:  user.Metadata,
				}
			}

			response := models.ListUsersResponse{
				Users:  userResponses,
				Total:  total,
				Limit:  req.Limit,
				Offset: req.Offset,
			}

			respBytes, err := json.Marshal(response)
			if err != nil {
				return dsl.ErrDatabaseOperation("json_marshal", err)
			}

			params.ResponseBody = respBytes
			params.StatusCode = 200
			return nil
		})

	// Execute flow
	if err := flow.Run(reqCtx); err != nil {
		h.handleError(w, err)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(h.serviceVars.Param.StatusCode)
	w.Write(h.serviceVars.Param.ResponseBody)
}

// Helper methods

// handleError handles errors and sends appropriate HTTP responses
func (h *UserHandler) handleError(w http.ResponseWriter, err error) {
	var statusCode int
	var message string

	if localizedErr, ok := err.(*dsl.LocalizedError); ok {
		switch localizedErr.Code {
		case "validation.required_field":
			statusCode = http.StatusBadRequest
			message = localizedErr.Message
		case "validation.invalid_value":
			statusCode = http.StatusBadRequest
			message = localizedErr.Message
		case "resource.not_found":
			statusCode = http.StatusNotFound
			message = localizedErr.Message
		case "database.operation_failed":
			statusCode = http.StatusInternalServerError
			message = "Internal server error"
		default:
			statusCode = http.StatusInternalServerError
			message = "Internal server error"
		}
	} else {
		statusCode = http.StatusInternalServerError
		message = "Internal server error"
	}

	// Log error
	if h.serviceVars.Logger != nil {
		h.serviceVars.Logger.Errorf("Request failed: %v, status_code: %d", err.Error(), statusCode)
	}

	response := map[string]any{
		"error":  message,
		"status": "error",
		"code":   statusCode,
	}

	respBytes, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(respBytes)
}

// hashPassword hashes a password (simple implementation, use bcrypt in production)
func (h *UserHandler) hashPassword(password string) string {
	// In production, use bcrypt or similar
	return repository.NewUserRepository(nil, nil, nil, nil).(*repository.UserRepository).HashPassword(password)
}
