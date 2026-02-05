package annotation_test

// Example: Config-based service injection with cfg: prefix
// This demonstrates the new @Inject "cfg:..." feature

// Domain interface
type Repository interface {
	GetUser(id string) (*User, error)
	SaveUser(user *User) error
}

type User struct {
	ID   string
	Name string
}

// @Service "postgres-repository"
type PostgresRepository struct {
	// @Inject "db-pool"
	DB any
}

var _ Repository = (*PostgresRepository)(nil)

func (s *PostgresRepository) GetUser(id string) (*User, error) {
	return &User{ID: id, Name: "User from Postgres"}, nil
}

func (s *PostgresRepository) SaveUser(user *User) error {
	return nil
}

// @Service "mysql-repository"
type MySQLRepository struct {
	// @Inject "db-pool"
	DB any
}

var _ Repository = (*MySQLRepository)(nil)

func (s *MySQLRepository) GetUser(id string) (*User, error) {
	return &User{ID: id, Name: "User from MySQL"}, nil
}

func (s *MySQLRepository) SaveUser(user *User) error {
	return nil
}

// @Handler name="user-service", prefix="/api/users"
type UserService struct {
	// Config-based injection - service name from config!
	// @Inject "cfg:repository.implementation"
	Repository Repository

	// Direct injection (existing behavior)
	// @Inject "logger"
	Logger any

	// Config value injection
	// @Inject "cfg:app.name"
	AppName string
}

// @Route "GET /{id}"
func (s *UserService) GetUser(id string) (*User, error) {
	return s.Repository.GetUser(id)
}

/*
Expected generated code in zz_generated.lokstra.go:

func UserServiceFactory(deps map[string]any, config map[string]any) any {
	svc := &UserService{
		// Config-based: registry resolves config["repository.implementation"] -> "postgres-repository"
		// Then auto-injects deps["cfg:repository.implementation"] (already resolved!)
		Repository: deps["cfg:repository.implementation"].(Repository),

		// Direct injection (as before)
		Logger: deps["logger"].(any),

		// Config value (as before)
		AppName: config["app.name"].(string),
	}
	return svc
}

func RegisterUserService() {
	lokstra_registry.RegisterRouterServiceType("user-service-factory",
		UserServiceFactory,
		UserServiceRemoteFactory,
		&deploy.ServiceTypeConfig{...})

	lokstra_registry.RegisterLazyService("user-service",
		"user-service-factory",
		map[string]any{
			// Both direct and config-based dependencies in depends-on
			"depends-on": []string{"cfg:repository.implementation", "logger"},

			// Config value that specifies which service to inject
			"repository.implementation": lokstra_registry.GetConfig("repository.implementation", ""),

			// Config values
			"app.name": lokstra_registry.GetConfig("app.name", ""),
		})
}

config.yaml:
---
configs:
  repository:
    implementation: "postgres-repository"  # Change to "mysql-repository" to switch!
  app:
    name: "MyApp"

service-definitions:
  db-pool:
    type: db-pool

  postgres-repository:
    type: postgres-repository

  mysql-repository:
    type: mysql-repository

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]
*/
