package services

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/models"
	"github.com/primadi/lokstra/examples/application_architecture/modules/user_management/repository"
	"github.com/primadi/lokstra/serviceapi"
)

// UserService interface defines the business logic contract for user operations
type UserService interface {
	// CreateUser creates a new user with validation
	CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error)

	// GetUser retrieves a user by ID
	GetUser(ctx context.Context, id int64) (*models.User, error)

	// GetUserByEmail retrieves a user by email
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)

	// UpdateUser updates a user with validation
	UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error)

	// DeleteUser soft deletes a user
	DeleteUser(ctx context.Context, id int64) error

	// ListUsers retrieves users with pagination and search
	ListUsers(ctx context.Context, req *models.ListUsersRequest) (*models.ListUsersResponse, error)
}

// userService implements UserService
type userService struct {
	userRepo          repository.UserRepository
	logger            serviceapi.Logger
	validationEnabled bool
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, logger serviceapi.Logger, validationEnabled bool) UserService {
	return &userService{
		userRepo:          userRepo,
		logger:            logger,
		validationEnabled: validationEnabled,
	}
}

// CreateUser creates a new user with validation
func (s *userService) CreateUser(ctx context.Context, req *models.CreateUserRequest) (*models.User, error) {
	s.logger.Infof("Creating user with email: %s", req.Email)

	// Validate request
	if s.validationEnabled {
		if err := req.Validate(); err != nil {
			s.logger.Errorf("User creation validation failed for %s: %s", req.Email, err.Error())
			return nil, fmt.Errorf("validation failed: %w", err)
		}
	}

	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		s.logger.Errorf("Failed to check user existence for %s: %s", req.Email, err.Error())
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if exists {
		s.logger.Warnf("Attempted to create user with existing email: %s", req.Email)
		return nil, fmt.Errorf("user with email '%s' already exists", req.Email)
	}

	// Create user
	user, err := s.userRepo.Create(ctx, req)
	if err != nil {
		s.logger.Errorf("Failed to create user %s: %s", req.Email, err.Error())
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.logger.Infof("User created successfully with ID %d: %s", user.ID, user.Email)
	return user, nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	s.logger.Debugf("Getting user by ID: %d", id)

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Errorf("Failed to get user %d: %s", id, err.Error())
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	s.logger.Debugf("User retrieved successfully: ID %d, Email %s", user.ID, user.Email)
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	s.logger.Debugf("Getting user by email: %s", email)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Errorf("Failed to get user by email %s: %s", email, err.Error())
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	s.logger.Debugf("User retrieved successfully by email: ID %d, Email %s", user.ID, user.Email)
	return user, nil
}

// UpdateUser updates a user with validation
func (s *userService) UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	s.logger.Infof("Updating user: %d", id)

	// Check if user exists
	existingUser, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Errorf("Failed to find user for update %d: %s", id, err.Error())
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate email uniqueness if email is being updated
	if req.Email != nil && *req.Email != existingUser.Email {
		exists, err := s.userRepo.ExistsByEmailExcludingID(ctx, *req.Email, id)
		if err != nil {
			s.logger.Errorf("Failed to check email uniqueness for %s (ID %d): %s", *req.Email, id, err.Error())
			return nil, fmt.Errorf("failed to validate email uniqueness: %w", err)
		}

		if exists {
			s.logger.Warnf("Attempted to update user %d with existing email: %s", id, *req.Email)
			return nil, fmt.Errorf("user with email '%s' already exists", *req.Email)
		}
	}

	// Update user
	user, err := s.userRepo.Update(ctx, id, req)
	if err != nil {
		s.logger.Errorf("Failed to update user %d: %s", id, err.Error())
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	s.logger.Infof("User updated successfully: ID %d, Email %s", user.ID, user.Email)
	return user, nil
}

// DeleteUser soft deletes a user
func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	s.logger.Infof("Deleting user: %d", id)

	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.logger.Errorf("Failed to find user for deletion %d: %s", id, err.Error())
		return fmt.Errorf("user not found: %w", err)
	}

	// Delete user
	err = s.userRepo.Delete(ctx, id)
	if err != nil {
		s.logger.Errorf("Failed to delete user %d: %s", id, err.Error())
		return fmt.Errorf("failed to delete user: %w", err)
	}

	s.logger.Infof("User deleted successfully: %d", id)
	return nil
}

// ListUsers retrieves users with pagination and search
func (s *userService) ListUsers(ctx context.Context, req *models.ListUsersRequest) (*models.ListUsersResponse, error) {
	s.logger.Debugf("Listing users: page=%d, page_size=%d, search=%s", req.Page, req.PageSize, req.Search)

	users, totalItems, err := s.userRepo.List(ctx, req)
	if err != nil {
		s.logger.Errorf("Failed to list users: %s", err.Error())
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int(totalItems) / req.PageSize
	if int(totalItems)%req.PageSize > 0 {
		totalPages++
	}

	response := &models.ListUsersResponse{
		Users: users,
		Pagination: &models.PaginationMeta{
			Page:       req.Page,
			PageSize:   req.PageSize,
			TotalItems: totalItems,
			TotalPages: totalPages,
		},
	}

	s.logger.Debugf("Users listed successfully: count=%d, total=%d", len(users), totalItems)
	return response, nil
}
