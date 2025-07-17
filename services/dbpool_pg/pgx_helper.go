package pg_dbpool

import (
	"context"

	"github.com/primadi/lokstra/serviceapi"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// dbExecutor is an interface that defines methods for executing SQL commands
// and queries using the pgx library.
// It abstracts the common operations needed to interact with a PostgreSQL database.
type dbExecutor interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// pgxCommandResult implements CommandResult for pgx
// It wraps the pgconn.CommandTag to provide RowsAffected method.
// It is used to return the number of rows affected by an Exec operation.
type pgxCommandResult struct {
	fnRowsAffected func() int64
}

func (r *pgxCommandResult) RowsAffected() int64 {
	return r.fnRowsAffected()
}

var _ serviceapi.CommandResult = (*pgxCommandResult)(nil)

// pgxRowIterator implements RowIterator for pgx
// It wraps pgx.Rows to provide methods for iterating over query results.
// It provides methods to check if there are more rows, scan the current row into destination variables,
// close the iterator, and check for errors.
// It is used to iterate over the results of a Query operation.
type pgxRowIterator struct {
	rows pgx.Rows
}

func (r *pgxRowIterator) Next() bool {
	return r.rows.Next()
}

func (r *pgxRowIterator) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *pgxRowIterator) Close() error {
	r.rows.Close()
	return nil
}

func (r *pgxRowIterator) Err() error {
	return r.rows.Err()
}

var _ serviceapi.RowIterator = (*pgxRowIterator)(nil)
