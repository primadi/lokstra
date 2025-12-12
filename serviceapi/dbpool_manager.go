package serviceapi

import (
	"context"
)

type DbPoolManager interface {
	// get or create DbPool for the given dsn
	GetDbPool(dsn, schema, rlsID string) (DbPool, error)

	// acquire connection for the given dsn, schema, and rlsID
	// if pool for dsn does not exist, create it
	// if schema and rlsID are provided, set them accordingly
	AcquireConn(ctx context.Context, dsn string, schema string, rlsID string) (DbConn, error)

	// ---------------------------------------
	// Named based pool management
	//----------------------------------------

	// set name for the given dsn, schema, and rlsID
	SetNamedDbPool(name string, dsn string, schema string, rlsID string)
	// get dsn, schema, rlsID for the given name
	GetNamedDbPoolInfo(name string) (string, string, string, error)
	// get DbPool for the given name
	GetNamedDbPool(name string) (DbPool, error)
	// remove name mapping
	RemoveNamedDbPool(name string)
	// acquire connection for the given name
	AcquireNamedConn(ctx context.Context, name string) (DbConn, error)

	Shutdownable
}
