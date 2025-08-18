package flow

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/serviceapi"
)

type Context struct {
	stdContext context.Context
	pool       serviceapi.DbPool
	schema     string
	dbConn     serviceapi.DbConn
	dbTx       serviceapi.DbTx

	vars map[string]any // runtime variables
}

func NewContext(stdContext context.Context, pool serviceapi.DbPool, schema string) *Context {
	return &Context{
		stdContext: stdContext,
		pool:       pool,
		schema:     schema,
		vars:       make(map[string]any),
	}
}

func (c *Context) StdContext() context.Context {
	return c.stdContext
}

func (c *Context) CurrentExecutor() (serviceapi.DbExecutor, error) {
	if c.dbTx != nil {
		return c.dbTx, nil
	}
	if c.dbConn != nil {
		return c.dbConn, nil
	}
	conn, err := c.pool.Acquire(c.stdContext, c.schema)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire DB connection: %s", err.Error())
	}
	c.dbConn = conn
	return c.dbConn, nil
}

func (c *Context) Set(name string, value any) {
	c.vars[name] = value
}

func (c *Context) Get(name string) (any, bool) {
	value, exists := c.vars[name]
	return value, exists
}

// GetAll returns all variables in the context
func (c *Context) GetAll() map[string]any {
	result := make(map[string]any)
	for key, value := range c.vars {
		result[key] = value
	}
	return result
}

func (c *Context) Exec(query string, args ...any) (serviceapi.CommandResult, error) {
	exec, err := c.CurrentExecutor()
	if err != nil {
		return nil, err
	}
	return exec.Exec(c.StdContext(), query, args...)
}

func (c *Context) QueryRow(query string, args ...any) (serviceapi.Row, error) {
	exec, err := c.CurrentExecutor()
	if err != nil {
		return nil, err
	}
	return exec.QueryRow(c.StdContext(), query, args...), nil
}

func (c *Context) Query(query string, args ...any) (serviceapi.Rows, error) {
	exec, err := c.CurrentExecutor()
	if err != nil {
		return nil, err
	}
	return exec.Query(c.StdContext(), query, args...)
}

func (c *Context) ForEach(fn func(serviceapi.Row) error, query string, args ...any) (int, error) {
	rows, err := c.Query(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		rowCount++
		if err := fn(rows); err != nil {
			return rowCount, fmt.Errorf("failed to process row: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return rowCount, fmt.Errorf("error iterating rows: %w", err)
	}
	return rowCount, nil
}

func (c *Context) QueryRowMap(query string, args ...any) (serviceapi.RowMap, error) {
	exec, err := c.CurrentExecutor()
	if err != nil {
		return nil, err
	}
	return exec.SelectOneRowMap(c.StdContext(), query, args...)
}

func (c *Context) QueryManyRowMap(query string, args ...any) ([]serviceapi.RowMap, error) {
	exec, err := c.CurrentExecutor()
	if err != nil {
		return nil, err
	}
	return exec.SelectManyRowMap(c.StdContext(), query, args...)
}
