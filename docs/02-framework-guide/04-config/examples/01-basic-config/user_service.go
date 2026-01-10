package main

// Request DTOs
type GetUserRequest struct {
	ID int `path:"id" validate:"required"`
}

type ListUsersRequest struct{}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

// @EndpointService name="user-service", prefix="/api/users", middlewares=["recovery", "request-logger"]
type UserService struct {
	// @Inject "user-repository"
	UserRepo UserRepository
}

// @Route "GET /{id}"
func (s *UserService) GetByID(p *GetUserRequest) (*User, error) {
	return s.UserRepo.GetByID(p.ID)
}

// @Route "GET /"
func (s *UserService) List(p *ListUsersRequest) ([]*User, error) {
	return s.UserRepo.List()
}

// @Route "POST /"
func (s *UserService) Create(p *CreateUserRequest) (*User, error) {
	user := &User{
		Name:  p.Name,
		Email: p.Email,
	}
	return s.UserRepo.Create(user)
}
