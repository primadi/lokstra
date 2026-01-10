package main

type GetUserRequest struct {
	ID int `path:"id" validate:"required"`
}

type ListUsersRequest struct{}

// @EndpointService name="user-service", prefix="/api/users"
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
