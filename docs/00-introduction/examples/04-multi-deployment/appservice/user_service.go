package appservice

import "github.com/primadi/lokstra/core/service"

type UserService interface {
	GetByID(p *GetUserParams) (*User, error)
	List(p *ListUsersParams) ([]*User, error)
}

type UserServiceImpl struct {
	DB *service.Cached[*Database]
}

var _ UserService = (*UserServiceImpl)(nil) // Ensure implementation

type GetUserParams struct {
	ID int `path:"id"`
}

type ListUsersParams struct{}

func (s *UserServiceImpl) GetByID(p *GetUserParams) (*User, error) {
	return s.DB.MustGet().GetUser(p.ID)
}

func (s *UserServiceImpl) List(p *ListUsersParams) ([]*User, error) {
	return s.DB.MustGet().GetAllUsers()
}

func NewUserService() UserService {
	return &UserServiceImpl{
		DB: service.LazyLoad[*Database]("db"),
	}
}
