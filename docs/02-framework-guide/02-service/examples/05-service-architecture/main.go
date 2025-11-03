package main

import (
	"fmt"

	"github.com/primadi/lokstra"
	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/lokstra_registry"
)

// Demonstrates clean architecture and DDD patterns

// Domain Layer
type User struct {
	ID    int
	Name  string
	Email string
}

// Repository Interface (Port)
type UserRepository interface {
	FindByID(id int) (*User, error)
	Save(user *User) error
}

// Service Layer (Use Case)
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

// Infrastructure Layer (Adapter)
type InMemoryUserRepository struct {
	users map[int]*User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: map[int]*User{
			1: {ID: 1, Name: "John", Email: "john@example.com"},
			2: {ID: 2, Name: "Jane", Email: "jane@example.com"},
		},
	}
}

func (r *InMemoryUserRepository) FindByID(id int) (*User, error) {
	if user, ok := r.users[id]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (r *InMemoryUserRepository) Save(user *User) error {
	r.users[user.ID] = user
	return nil
}

// Factories
func UserRepositoryFactory(deps map[string]any, config map[string]any) any {
	return NewInMemoryUserRepository()
}

func UserServiceFactory(deps map[string]any, config map[string]any) any {
	repo := lokstra_registry.GetService[UserRepository]("user-repository")
	return NewUserService(repo)
}

// HTTP Handlers
type GetUserParams struct {
	ID int `path:"id"`
}

func GetUser(params *GetUserParams) *response.ApiHelper {
	svc := lokstra_registry.GetService[*UserService]("user-service")
	user, err := svc.GetUser(params.ID)
	if err != nil {
		return response.NewApiNotFound(err.Error())
	}
	return response.NewApiOk(user)
}

func Home() *response.Response {
	return response.NewHtmlResponse(`<!DOCTYPE html>
<html><head><title>Service Architecture</title></head>
<body><h1>🏛️ Service Architecture Example</h1>
<p>Demonstrates clean architecture and DDD patterns.</p>
<ul><li>GET /users/:id - Get user (clean architecture)</li></ul>
</body></html>`)
}

func main() {
	lokstra_registry.RegisterServiceType("user-repository", UserRepositoryFactory, nil)
	lokstra_registry.RegisterServiceType("user-service", UserServiceFactory, nil)

	lokstra_registry.RegisterLazyService("user-repository", UserRepositoryFactory, nil)
	lokstra_registry.RegisterLazyService("user-service", UserServiceFactory, nil)

	router := lokstra.NewRouter("service-architecture")
	router.GET("/", Home)
	router.GET("/users/:id", GetUser)

	app := lokstra.NewApp("service-architecture", ":3000", router)
	fmt.Println("🚀 Service Architecture Example")
	fmt.Println("📍 http://localhost:3000")

	if err := app.Run(0); err != nil {
		fmt.Println("Error:", err)
	}
}
