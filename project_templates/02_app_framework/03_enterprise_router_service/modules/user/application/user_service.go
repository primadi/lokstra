package application

import (
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user/domain"
)

// @RouterService name="user-service", prefix="/api", middlewares=["recovery", "request-logger", "simple-auth"]
type UserServiceImpl struct {
	// @Inject "user-repository"
	UserRepo domain.UserRepository
}

// Ensure implementation
var _ domain.UserService = (*UserServiceImpl)(nil)

// @Route "GET /users/{id}", ["mw-test param1=123, param2=abc"]
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
	return s.UserRepo.GetByID(p.ID)
}

// @Route "GET /users", ["mw-test"]
func (s *UserServiceImpl) List(p *domain.ListUsersRequest) ([]*domain.User, error) {
	return s.UserRepo.List()
}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
	u := &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		RoleID: p.RoleID,
		Status: "active",
	}
	return s.UserRepo.Create(u)
}

// @Route "PUT /users/{id}"
func (s *UserServiceImpl) Update(p *domain.UpdateUserRequest) (*domain.User, error) {
	u := &domain.User{
		ID:     p.ID,
		Name:   p.Name,
		Email:  p.Email,
		RoleID: p.RoleID,
	}
	return s.UserRepo.Update(u)
}

// @Route "POST /users/{id}/suspend"
func (s *UserServiceImpl) Suspend(p *domain.SuspendUserRequest) error {
	user, err := s.UserRepo.GetByID(p.ID)
	if err != nil {
		return err
	}
	user.Status = "suspended"
	_, err = s.UserRepo.Update(user)
	return err
}

// @Route "POST /users/{id}/activate"
func (s *UserServiceImpl) Activate(p *domain.ActivateUserRequest) error {
	user, err := s.UserRepo.GetByID(p.ID)
	if err != nil {
		return err
	}
	user.Status = "active"
	_, err = s.UserRepo.Update(user)
	return err
}

// @Route "DELETE /users/{id}"
func (s *UserServiceImpl) Delete(p *domain.DeleteUserRequest) error {
	return s.UserRepo.Delete(p.ID)
}
