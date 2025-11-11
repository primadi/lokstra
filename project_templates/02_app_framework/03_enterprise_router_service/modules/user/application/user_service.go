package application

import (
	"github.com/primadi/lokstra/core/service"
	"github.com/primadi/lokstra/project_templates/02_app_framework/03_enterprise_router_service/modules/user/domain"
)

// @RouterService name="user-service", prefix="/api", middlewares=["recovery", "request-logger"]
type UserServiceImpl struct {
	// @Inject "user-repository"
	UserRepo *service.Cached[domain.UserRepository]
}

// Ensure implementation
var _ domain.UserService = (*UserServiceImpl)(nil)

// @Route "GET /users/{id}"
func (s *UserServiceImpl) GetByID(p *domain.GetUserRequest) (*domain.User, error) {
	return s.UserRepo.MustGet().GetByID(p.ID)
}

// @Route "GET /users"
func (s *UserServiceImpl) List(p *domain.ListUsersRequest) ([]*domain.User, error) {
	return s.UserRepo.MustGet().List()
}

// @Route "POST /users"
func (s *UserServiceImpl) Create(p *domain.CreateUserRequest) (*domain.User, error) {
	u := &domain.User{
		Name:   p.Name,
		Email:  p.Email,
		RoleID: p.RoleID,
		Status: "active",
	}
	return s.UserRepo.MustGet().Create(u)
}

// @Route "PUT /users/{id}"
func (s *UserServiceImpl) Update(p *domain.UpdateUserRequest) (*domain.User, error) {
	u := &domain.User{
		ID:     p.ID,
		Name:   p.Name,
		Email:  p.Email,
		RoleID: p.RoleID,
	}
	return s.UserRepo.MustGet().Update(u)
}

// @Route "POST /users/{id}/suspend"
func (s *UserServiceImpl) Suspend(p *domain.SuspendUserRequest) error {
	user, err := s.UserRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return err
	}
	user.Status = "suspended"
	_, err = s.UserRepo.MustGet().Update(user)
	return err
}

// @Route "POST /users/{id}/activate"
func (s *UserServiceImpl) Activate(p *domain.ActivateUserRequest) error {
	user, err := s.UserRepo.MustGet().GetByID(p.ID)
	if err != nil {
		return err
	}
	user.Status = "active"
	_, err = s.UserRepo.MustGet().Update(user)
	return err
}

// @Route "DELETE /users/{id}"
func (s *UserServiceImpl) Delete(p *domain.DeleteUserRequest) error {
	return s.UserRepo.MustGet().Delete(p.ID)
}

func Register() {
	// do nothing, just to make sure the package is loaded
}
