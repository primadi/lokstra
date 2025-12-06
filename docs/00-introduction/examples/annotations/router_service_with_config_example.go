package application

import (
	"fmt"
	"time"
)

// Request DTOs
type GetUserRequest struct {
	ID string `path:"id"`
}

type ListUsersRequest struct {
	Page     int `query:"page"`
	PageSize int `query:"page_size"`
}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type UpdateUserRequest struct {
	ID    string `path:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type DeleteUserRequest struct {
	ID string `path:"id"`
}

// Response DTOs
type UserResponse struct {
	User     *User          `json:"user"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type UserListResponse struct {
	Users    []*User        `json:"users"`
	PageSize int            `json:"page_size"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Example: @RouterService with @Inject (optional), @InjectCfg, and @Route

// @RouterService name="user-api-service", prefix="/api/v1/users", middlewares=["recovery", "request-logger"]
type UserAPIService struct {
	// Required dependency
	// @Inject "user-repository"
	UserRepo UserRepository

	// Cache dependency for performance
	// @Inject "cache-service"
	Cache CacheService

	// Configuration: API rate limiting
	// @InjectCfg key="api.rate-limit.enabled", default=true
	RateLimitEnabled bool

	// @InjectCfg key="api.rate-limit.max-requests", default=100
	MaxRequests int

	// @InjectCfg "api.rate-limit.window", "1m"
	RateLimitWindow time.Duration

	// Configuration: Pagination
	// @InjectCfg key="api.pagination.default-page-size", default="20"
	DefaultPageSize int

	// @InjectCfg key="api.pagination.max-page-size", default="100"
	MaxPageSize int

	// Configuration: Response
	// @InjectCfg key="api.response.include-metadata", default="true"
	IncludeMetadata bool

	// Configuration: Authentication
	// @InjectCfg "api.jwt-secret"
	JwtSecret string

	// @InjectCfg key="api.token-expiry", default="24h"
	TokenExpiry time.Duration
}

// @Route "GET /{id}"
func (s *UserAPIService) GetByID(p *GetUserRequest) (*UserResponse, error) {
	// Check cache if available
	if s.Cache != nil {
		cacheKey := "user:" + p.ID
		if cached, err := s.Cache.Get(cacheKey); err == nil {
			return cached.(*UserResponse), nil
		}
	}

	user, err := s.UserRepo.GetByID(p.ID)
	if err != nil {
		return nil, err
	}

	response := &UserResponse{
		User:     user,
		Metadata: s.buildMetadata(),
	}

	// Cache if available
	if s.Cache != nil {
		_ = s.Cache.Set("user:"+p.ID, response, 5*time.Minute)
	}

	return response, nil
}

// @Route "GET /"
func (s *UserAPIService) List(p *ListUsersRequest) (*UserListResponse, error) {
	// Apply default page size
	if p.PageSize == 0 {
		p.PageSize = s.DefaultPageSize
	}

	// Enforce max page size
	if p.PageSize > s.MaxPageSize {
		p.PageSize = s.MaxPageSize
	}

	users, err := s.UserRepo.List()
	if err != nil {
		return nil, err
	}

	response := &UserListResponse{
		Users:    users,
		PageSize: p.PageSize,
	}

	if s.IncludeMetadata {
		response.Metadata = s.buildMetadata()
	}

	return response, nil
}

// @Route "POST /", ["auth"]
func (s *UserAPIService) Create(p *CreateUserRequest) (*UserResponse, error) {
	// Check rate limit
	if s.RateLimitEnabled {
		// Rate limiting logic here
		_ = s.MaxRequests
		_ = s.RateLimitWindow
	}

	user := &User{
		Name:  p.Name,
		Email: p.Email,
	}

	created, err := s.UserRepo.Create(user)
	if err != nil {
		return nil, err
	}

	response := &UserResponse{
		User:     created,
		Metadata: s.buildMetadata(),
	}

	return response, nil
}

// @Route "PUT /{id}", ["auth"]
func (s *UserAPIService) Update(p *UpdateUserRequest) (*UserResponse, error) {
	user := &User{
		ID:    p.ID,
		Name:  p.Name,
		Email: p.Email,
	}

	updated, err := s.UserRepo.Update(user)
	if err != nil {
		return nil, err
	}

	// Invalidate cache if available
	if s.Cache != nil {
		_ = s.Cache.Delete("user:" + p.ID)
	}

	response := &UserResponse{
		User:     updated,
		Metadata: s.buildMetadata(),
	}

	return response, nil
}

// @Route "DELETE /{id}", ["auth", "admin"]
func (s *UserAPIService) Delete(p *DeleteUserRequest) error {
	err := s.UserRepo.Delete(p.ID)
	if err != nil {
		return err
	}

	// Invalidate cache if available
	if s.Cache != nil {
		_ = s.Cache.Delete("user:" + p.ID)
	}

	return nil
}

func (s *UserAPIService) buildMetadata() map[string]any {
	if !s.IncludeMetadata {
		return nil
	}

	return map[string]any{
		"timestamp":         time.Now().Unix(),
		"rate_limit_window": s.RateLimitWindow.String(),
		"max_page_size":     fmt.Sprintf("%d", s.MaxPageSize),
	}
}

func Register() {
	// Package auto-loaded by code generation
}
