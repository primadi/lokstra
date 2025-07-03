package pg_dbpool

import "context"

// DBPool defines a connection pool interface
// supporting schema-aware connection acquisition
// and future multi-backend support.
type DBPool interface {
	Acquire(schema string) (DBConn, error)
}

// DBConn represents a live DB connection (e.g. from pgxpool)
type DBConn interface {
	Exec(ctx context.Context, query string, args ...any) (CommandResult, error)
	Query(ctx context.Context, query string, args ...any) (RowIterator, error)
	QueryRow(ctx context.Context, query string, args ...any) RowScanner

	SelectOne(ctx context.Context, query string, args []any, dest ...any) error
	SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error
	SelectMany(ctx context.Context, query string, args ...any) (any, error)
	SelectManyWithMapper(ctx context.Context,
		fnScan func(RowScanner) (any, error), query string, args ...any) (any, error)

	IsExists(ctx context.Context, query string, args ...any) (bool, error)

	Begin(ctx context.Context) (Tx, error)
	Transaction(ctx context.Context, fn func(tx DBConn) error) error

	Release() error
}

// Tx represents an ongoing transaction
type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	DBConn
}

// CommandResult abstracts result from Exec()
type CommandResult interface {
	RowsAffected() int64
}

// RowIterator abstracts rows from Query()
type RowIterator interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// RowScanner abstracts single row result from QueryRow()
type RowScanner interface {
	Scan(dest ...any) error
}
