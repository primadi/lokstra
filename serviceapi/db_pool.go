package serviceapi

import "context"

// DbPool defines a connection pool interface
// supporting schema-aware connection acquisition
// and future multi-backend support.
type DbPool interface {
	// Acquire a connection for the specified schema.
	// If schema is empty, a default connection is provided.
	// it will set the search_path to the specified schema after acquiring the connection.
	// For multi-tenant, use AcquireMultiTenant with tenantID.
	Acquire(ctx context.Context, schema string) (DbConn, error)

	// AcquireMultiTenant acquires a connection for the specified schema and tenantID.
	// If tenantID is empty, a default connection is provided.
	// it will set the search_path to the specified schema after acquiring the connection.
	// and set Row Level Security (RLS) context to the specified tenantID.
	// This is useful for multi-tenant applications.
	AcquireMultiTenant(ctx context.Context, schema string, tenantID string) (DbConn, error)
}

type RowMap = map[string]any

// DbConn represents a live DB connection (e.g. from pgxpool)
type DbConn interface {
	Begin(ctx context.Context) (DbTx, error)
	Transaction(ctx context.Context, fn func(tx DbExecutor) error) error
	Ping(context context.Context) error
	Release() error
	DbExecutor
}

// DbTx represents an ongoing transaction
type DbTx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	DbExecutor
}

type DbExecutor interface {
	Exec(ctx context.Context, query string, args ...any) (CommandResult, error)
	Query(ctx context.Context, query string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) Row

	SelectOne(ctx context.Context, query string, args []any, dest ...any) error
	SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error

	SelectOneRowMap(ctx context.Context, query string, args ...any) (RowMap, error)
	SelectManyRowMap(ctx context.Context, query string, args ...any) ([]RowMap, error)

	SelectManyWithMapper(ctx context.Context,
		fnScan func(Row) (any, error), query string, args ...any) (any, error)

	IsExists(ctx context.Context, query string, args ...any) (bool, error)
	IsErrorNoRows(err error) bool
}

// CommandResult abstracts result from Exec()
type CommandResult interface {
	RowsAffected() int64
}

// Rows abstracts rows from Query()
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// Row abstracts single row result from QueryRow()
type Row interface {
	Scan(dest ...any) error
}

// --------------------
// CommandResultImpl is a concrete implementation of CommandResult
// --------------------

type CommandResultImpl struct {
	fnRowsAffected func() int64
}

// RowsAffected implements CommandResult.
func (c *CommandResultImpl) RowsAffected() int64 {
	return c.fnRowsAffected()
}

var _ CommandResult = (*CommandResultImpl)(nil)

func NewCommandResult(fnRowsAffected func() int64) CommandResult {
	return &CommandResultImpl{
		fnRowsAffected: fnRowsAffected,
	}
}
