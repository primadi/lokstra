package annotation_test

// Example: Config-based service injection with cfg: prefix
// This demonstrates the new @Inject "cfg:..." feature

// Domain interface
type Store interface {
	GetUser(id string) (*User, error)
	SaveUser(user *User) error
}

type User struct {
	ID   string
	Name string
}

// @Service "postgres-store"
type PostgresStore struct {
	// @Inject "db-pool"
	DB any
}

var _ Store = (*PostgresStore)(nil)

func (s *PostgresStore) GetUser(id string) (*User, error) {
	return &User{ID: id, Name: "User from Postgres"}, nil
}

func (s *PostgresStore) SaveUser(user *User) error {
	return nil
}

// @Service "mysql-store"
type MySQLStore struct {
	// @Inject "db-pool"
	DB any
}

var _ Store = (*MySQLStore)(nil)

func (s *MySQLStore) GetUser(id string) (*User, error) {
	return &User{ID: id, Name: "User from MySQL"}, nil
}

func (s *MySQLStore) SaveUser(user *User) error {
	return nil
}

// @RouterService name="user-service", prefix="/api/users"
type UserService struct {
	// Config-based injection - service name from config!
	// @Inject "cfg:store.implementation"
	Store Store

	// Direct injection (existing behavior)
	// @Inject "logger"
	Logger any

	// Config value injection (existing @InjectCfgValue)
	// @InjectCfgValue "app.name"
	AppName string
}

// @Route "GET /{id}"
func (s *UserService) GetUser(id string) (*User, error) {
	return s.Store.GetUser(id)
}

/*
Expected generated code in zz_generated.lokstra.go:

func UserServiceFactory(deps map[string]any, config map[string]any) any {
	svc := &UserService{
		// Config-based: reads config["store.implementation"] -> "postgres-store"
		// Then injects deps["postgres-store"]
		Store: deps[config["store.implementation"].(string)].(Store),

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
			// Only direct dependencies
			"depends-on": []string{"logger"},

			// Config-based dependency resolved at runtime
			"store.implementation": lokstra_registry.GetConfig("store.implementation", ""),

			// Config values
			"app.name": lokstra_registry.GetConfig("app.name", ""),
		})
}

config.yaml:
---
configs:
  store:
    implementation: "postgres-store"  # Change to "mysql-store" to switch!
  app:
    name: "MyApp"

service-definitions:
  db-pool:
    type: db-pool

  postgres-store:
    type: postgres-store

  mysql-store:
    type: mysql-store

deployments:
  development:
    servers:
      api:
        addr: ":8080"
        published-services: [user-service]
*/
