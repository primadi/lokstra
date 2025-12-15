package serviceapi

import (
	"context"
)

// DbPoolInfo holds database DSN, schema name, and RLS context
type DbPoolInfo struct {
	Dsn        string
	Schema     string
	RlsContext map[string]string
}

type DbPoolManager interface {
	// Get all named DbPools
	GetAllNamedDbPools() map[string]*DbPoolInfo

	// get or create DbPool for the given dsn
	GetDbPool(dsn, schema string, rlsContext map[string]string) (DbPool, error)

	// acquire connection for the given dsn, schema, and rlsContext
	// if pool for dsn does not exist, create it
	// if schema and rlsContext are provided, set them accordingly
	AcquireConn(ctx context.Context, dsn string, schema string, rlsContext map[string]string) (DbConn, error)

	// ---------------------------------------
	// Named based pool management
	//----------------------------------------

	// set name for the given dsn, schema, and rlsContext
	SetNamedDbPool(name string, dsn string, schema string, rlsContext map[string]string)
	// get dsn, schema, rlsContext for the given name
	GetNamedDbPoolInfo(name string) (string, string, map[string]string, error)
	// get DbPool for the given name
	GetNamedDbPool(name string) (DbPool, error)
	// remove name mapping
	RemoveNamedDbPool(name string)
	// acquire connection for the given name
	AcquireNamedConn(ctx context.Context, name string) (DbConn, error)

	Shutdownable
}
