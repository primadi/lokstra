package handlers

import (
	"github.com/primadi/lokstra/cmd/projects/user_management/repository"
	"github.com/primadi/lokstra/common/utils"
	"github.com/primadi/lokstra/core/flow"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi/auth"
)

func CreateNewUserHandler() request.HandlerFunc {
	return flow.NewFlow[CreateUserRequestDTO]("CreateNewUser").
		AddValidateRequired("Username", "Email", "Password").
		AddAction("create_user", createUserAction).AsHandler()
}

func CreateUpdateUserHandler() request.HandlerFunc {
	return flow.NewFlow[UpdateUserRequestDTO]("UpdateUser").
		AddValidateRequired("ID"). // ID dari path harus ada
		AddValidateRequest([]flow.FieldValidator{
			{
				Field: "Email",
				Rules: []flow.ValidationRule{flow.Email()}, // Optional field, validate only if provided
			},
			{
				Field: "Password",
				Rules: []flow.ValidationRule{flow.MinLength(8)}, // Optional field, validate only if provided
			},
		}).
		AddAction("update_user", updateUserAction).AsHandler()
}

func CreateDeleteUserHandler() request.HandlerFunc {
	return flow.NewFlow[DeleteUserRequestDTO]("DeleteUser").
		AddValidateRequired("UserID").
		AddAction("delete_user", deleteUserAction).AsHandler()
}

func CreateListUserHandler() request.HandlerFunc {
	return flow.NewFlow[ListUserRequestDTO]("ListUsers").
		AddPaginationQueryAction().
		AddAction("list_users", listUsersAction).AsHandler()
}

func CreateGetUserByNameHandler() request.HandlerFunc {
	return flow.NewFlow[GetUserByNameRequestDTO]("GetUserByName").
		AddValidateRequired("Username").
		AddAction("get_user_by_name", getUserByNameAction).AsHandler()
}

func CreateGetUserByIDHandler() request.HandlerFunc {
	return flow.NewFlow[GetUserByIDRequestDTO]("GetUserByID").
		AddValidateRequired("ID").
		AddAction("get_user_by_id", getUserByIDAction).AsHandler()
}

// Extract to separate function for better readability
func createUserAction(fctx *flow.Context[CreateUserRequestDTO]) error {
	// Map DTO to entity with cleaner struct initialization
	user := auth.User{
		Username: fctx.Params.Username,
		Email:    fctx.Params.Email,
		Metadata: fctx.Params.Metadata,
		IsActive: fctx.Params.IsActive != nil && *fctx.Params.IsActive,
		TenantID: "default", // TODO: Get from request context
	}

	// Handle password hashing with error check
	var err error
	user.PasswordHash, err = utils.HashPassword(fctx.Params.Password)
	if err != nil {
		return fctx.ErrorInternal("Failed to process password")
	}

	// Create user via repository
	repo := repository.NewUserRepository(fctx.GetDbExecutor())
	return repo.CreateUser(fctx, &user)
}

func updateUserAction(fctx *flow.Context[UpdateUserRequestDTO]) error {
	// ID sudah otomatis di-bind dari path parameter
	userID := fctx.Params.ID

	// Get existing user first
	repo := repository.NewUserRepository(fctx.GetDbExecutor())
	// TODO: Replace "default" with actual tenant ID from request context
	existingUser, err := repo.GetUserByName(fctx, "default", userID)
	if err != nil {
		return fctx.ErrorNotFound("User not found")
	}

	// Update fields that are provided
	if fctx.Params.Email != "" {
		existingUser.Email = fctx.Params.Email
	}

	if fctx.Params.Password != "" {
		hashedPassword, err := utils.HashPassword(fctx.Params.Password)
		if err != nil {
			return fctx.ErrorInternal("Failed to process password")
		}
		existingUser.PasswordHash = hashedPassword
	}

	if fctx.Params.IsActive != nil {
		existingUser.IsActive = *fctx.Params.IsActive
	}

	if fctx.Params.Metadata != nil {
		existingUser.Metadata = fctx.Params.Metadata
	}

	// Update user via repository
	return repo.UpdateUser(fctx, existingUser)
}

func deleteUserAction(fctx *flow.Context[DeleteUserRequestDTO]) error {
	userID := fctx.Params.UserID

	// Delete user via repository
	repo := repository.NewUserRepository(fctx.GetDbExecutor())
	// TODO: Replace "default" with actual tenant ID from request context
	return repo.DeleteUser(fctx, "default", userID)
}

func listUsersAction(fctx *flow.Context[ListUserRequestDTO]) error {
	// Get pagination from context (set by AddPaginationQueryAction)
	pagination, exists := fctx.GetPagination()
	if !exists {
		return fctx.ErrorInternal("Pagination context not found")
	}

	// Get users via repository with pagination
	repo := repository.NewUserRepository(fctx.GetDbExecutor())
	// TODO: Replace "default" with actual tenant ID from request context
	users, total, err := repo.ListUsersWithPagination(fctx, "default", pagination.Page, pagination.PageSize, pagination.Filter)
	if err != nil {
		return fctx.ErrorInternal("Failed to retrieve users")
	}

	// Return paginated response using helper
	return fctx.PaginatedOk(users, total)
}

func getUserByNameAction(fctx *flow.Context[GetUserByNameRequestDTO]) error {
	repo := repository.NewUserRepository(fctx.GetDbExecutor())

	// TODO: Replace "default" with actual tenant ID from request context
	user, err := repo.GetUserByName(fctx, "default", fctx.Params.Username)
	if err != nil {
		// Check if user not found or other error
		return fctx.ErrorNotFound("User not found with username: " + fctx.Params.Username)
	}

	return fctx.Ok(user)
}

func getUserByIDAction(fctx *flow.Context[GetUserByIDRequestDTO]) error {
	repo := repository.NewUserRepository(fctx.GetDbExecutor())

	// TODO: Replace "default" with actual tenant ID from request context
	user, err := repo.GetUserByID(fctx, "default", fctx.Params.ID)
	if err != nil {
		// Check if user not found or other error
		return fctx.ErrorNotFound("User not found with ID: " + fctx.Params.ID)
	}

	return fctx.Ok(user)
}
