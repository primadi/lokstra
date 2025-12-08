package testapp

// Domain interface
type TenantStore interface {
	GetTenant(id string) (*Tenant, error)
	SaveTenant(tenant *Tenant) error
}

type Tenant struct {
	ID   string
	Name string
}

// @Service "postgres-tenant-store"
type PostgresTenantStore struct {
	// @Inject "db-pool"
	DB any
}

var _ TenantStore = (*PostgresTenantStore)(nil)

func (s *PostgresTenantStore) GetTenant(id string) (*Tenant, error) {
	return &Tenant{ID: id, Name: "Postgres Tenant"}, nil
}

func (s *PostgresTenantStore) SaveTenant(tenant *Tenant) error {
	return nil
}

// @Service "mysql-tenant-store"
type MySQLTenantStore struct {
	// @Inject "db-pool"
	DB any
}

var _ TenantStore = (*MySQLTenantStore)(nil)

func (s *MySQLTenantStore) GetTenant(id string) (*Tenant, error) {
	return &Tenant{ID: id, Name: "MySQL Tenant"}, nil
}

func (s *MySQLTenantStore) SaveTenant(tenant *Tenant) error {
	return nil
}

// @RouterService name="tenant-service", prefix="/api/tenants", middlewares=["recovery"]
type TenantService struct {
	// @Inject "cfg:store.tenant-store"
	Store TenantStore
}

// @Route "GET /{id}"
func (s *TenantService) GetTenant(id string) (*Tenant, error) {
	return s.Store.GetTenant(id)
}

// @Route "POST /"
func (s *TenantService) CreateTenant(t *Tenant) (*Tenant, error) {
	return t, s.Store.SaveTenant(t)
}
