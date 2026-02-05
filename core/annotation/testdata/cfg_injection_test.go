package testapp

// Domain interface
type TenantRepository interface {
	GetTenant(id string) (*Tenant, error)
	SaveTenant(tenant *Tenant) error
}

type Tenant struct {
	ID   string
	Name string
}

// @Service "postgres-tenant-repository"
type PostgresTenantRepository struct {
	// @Inject "db-pool"
	DB any
}

var _ TenantRepository = (*PostgresTenantRepository)(nil)

func (s *PostgresTenantRepository) GetTenant(id string) (*Tenant, error) {
	return &Tenant{ID: id, Name: "Postgres Tenant"}, nil
}

func (s *PostgresTenantRepository) SaveTenant(tenant *Tenant) error {
	return nil
}

// @Service "mysql-tenant-repository"
type MySQLTenantRepository struct {
	// @Inject "db-pool"
	DB any
}

var _ TenantRepository = (*MySQLTenantRepository)(nil)

func (s *MySQLTenantRepository) GetTenant(id string) (*Tenant, error) {
	return &Tenant{ID: id, Name: "MySQL Tenant"}, nil
}

func (s *MySQLTenantRepository) SaveTenant(tenant *Tenant) error {
	return nil
}

// @Handler name="tenant-service", prefix="/api/tenants", middlewares=["recovery"]
type TenantService struct {
	// @Inject "cfg:repository.tenant-repository"
	Repository TenantRepository
}

// @Route "GET /{id}"
func (s *TenantService) GetTenant(id string) (*Tenant, error) {
	return s.Repository.GetTenant(id)
}

// @Route "POST /"
func (s *TenantService) CreateTenant(t *Tenant) (*Tenant, error) {
	return t, s.Repository.SaveTenant(t)
}
