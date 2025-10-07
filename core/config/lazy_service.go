package config

// GenericLazyService is a placeholder for lazy service injection during config processing.
// It only holds the service name - actual service resolution happens in factories.
//
// This is used when processing layered service configs with depends-on.
// Config values that reference other services get wrapped in GenericLazyService.
//
// Example:
//
//	// In config YAML:
//	services:
//	  infrastructure:
//	    - name: db
//	  repository:
//	    - name: user-repo
//	      depends-on: [db]
//	      config:
//	        db_service: db  # This becomes GenericLazyService("db")
//
//	// In factory:
//	func NewUserRepo(cfg map[string]interface{}) (*UserRepo, error) {
//	    dbLazy := utils.GetLazyService[Database](cfg, "db_service")
//	    return &UserRepo{
//	        db: dbLazy,  // Lazy[Database]
//	    }, nil
//	}
type GenericLazyService struct {
	serviceName string
}

// NewGenericLazyService creates a new generic lazy service placeholder.
func NewGenericLazyService(serviceName string) *GenericLazyService {
	return &GenericLazyService{
		serviceName: serviceName,
	}
}

// ServiceName returns the name of the service to be lazily loaded.
func (g *GenericLazyService) ServiceName() string {
	return g.serviceName
}
