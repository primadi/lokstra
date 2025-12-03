package main

// RegisterMiddleware demonstrates annotation usage.
// This is a documentation comment explaining how to use annotations.
//
// Example of annotation in code (this should be IGNORED - TAB indented):
//
//	@RouterService name="example-service", prefix="/api/example"
//
// The above annotation is indented with TAB, so it's treated as code example.
type RegisterMiddleware struct{}

func (r *RegisterMiddleware) Handle() {
	// Implementation
}

// AnotherFunction shows another example.
//
// You can use annotations like this (this should be IGNORED - multi-space indented):
//
//    @Route "GET /test"
//
// The above is indented with multiple spaces, treated as code example.
func AnotherFunction() {}

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
	// @Inject "user-repository"
	UserRepo UserRepository
}

// @Route "GET /{id}"
func (s *UserService) GetByID(p *GetUserParams) (*User, error) {
	return s.UserRepo.GetByID(p.ID)
}

// @Route "POST /"
func (s *UserService) Create(p *CreateUserParams) (*User, error) {
	return s.UserRepo.Create(p)
}

// Dummy types for compilation
type UserRepository interface {
	GetByID(id string) (*User, error)
	Create(p *CreateUserParams) (*User, error)
}

type User struct {
	ID   string
	Name string
}

type GetUserParams struct {
	ID string
}

type CreateUserParams struct {
	Name string
}
