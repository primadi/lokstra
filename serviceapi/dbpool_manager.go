package serviceapi

import (
	"context"

	"github.com/primadi/lokstra/lokstra_registry"
)

type DbPoolManager interface {
	// get or create DbPool for the given dsn
	GetDsnPool(dsn string) (DbPool, error)

	// ---------------------------------------
	// Tenant based pool management
	//----------------------------------------

	// set dsn and schema for the given tenant
	SetTenantDsn(tenant string, dsn string, schema string)
	// get dsn and schema for the given tenant
	GetTenantDsn(tenant string) (string, string, error)
	// get DbPool for the given tenant
	GetTenantPool(tenant string) (DbPoolWithTenant, error)
	// remove tenant mapping
	RemoveTenant(tenant string)
	// acquire connection for the given tenant
	AcquireTenantConn(ctx context.Context, tenant string) (DbConn, error)

	// ---------------------------------------
	// Named based pool management
	//----------------------------------------

	// set dsn and schema for the given name
	SetNamedDsn(name string, dsn string, schema string)
	// get dsn and schema for the given name
	GetNamedDsn(name string) (string, string, error)
	// get DbPool for the given name
	GetNamedPool(name string) (DbPoolWithSchema, error)
	// remove name mapping
	RemoveNamed(name string)
	// acquire connection for the given name
	AcquireNamedConn(ctx context.Context, name string) (DbConn, error)

	lokstra_registry.Shutdownable
}
