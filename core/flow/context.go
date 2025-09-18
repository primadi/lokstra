package flow

import (
	"fmt"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

type Context[T any] struct {
	*request.Context
	Params *T
	flow   *Flow[T]
	dbConn serviceapi.DbConn
	dbTx   serviceapi.DbTx

	// Store for passing data between flow steps
	store map[string]any
}

func newContext[T any](f *Flow[T], reqCtx *request.Context) *Context[T] {
	return &Context[T]{
		flow:    f,
		store:   make(map[string]any),
		Context: reqCtx,
	}
}

func (c *Context[T]) GetDbExecutor() serviceapi.DbExecutor {
	if c.dbTx != nil {
		return c.dbTx
	}
	if c.dbConn == nil {
		if c.flow.DbPool == nil {
			c.flow.DbPool = defaultDbPool
		}
		if c.flow.DbPool == nil {
			// Return an error if no database pool is available
			panic("no database pool available - make sure to configure flow_dbPool in global settings")
		}
		if c.flow.DbSchemaName == "" {
			c.flow.DbSchemaName = defaultDbSchemaName
		}

		var err error
		c.dbConn, err = c.flow.DbPool.Acquire(c.Context, c.flow.DbSchemaName)
		if err != nil {
			panic(fmt.Sprintf("failed to acquire database connection: %v", err))
		}
	}
	return c.dbConn
}

func (c *Context[T]) RollbackTx() error {
	if c.dbTx == nil {
		return fmt.Errorf("no transaction to rollback")
	}

	// Rollback the transaction and handle potential errors
	if err := c.dbTx.Rollback(c.Context); err != nil {
		c.cleanupDbResources()
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	// Clean up resources after rollback
	c.cleanupDbResources()
	return nil
}

func (c *Context[T]) CommitTx() error {
	if c.dbTx == nil {
		return fmt.Errorf("no transaction to commit")
	}

	// Commit the transaction and handle potential errors
	if err := c.dbTx.Commit(c.Context); err != nil {
		// Even if commit fails, we should still clean up resources
		c.cleanupDbResources()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Clean up resources after successful commit
	c.cleanupDbResources()
	return nil
}

// cleanupDbResources centralizes the cleanup logic for database resources
func (c *Context[T]) cleanupDbResources() {
	if c.dbTx != nil {
		c.dbTx = nil
	}
	if c.dbConn != nil {
		c.dbConn.Release()
		c.dbConn = nil
	}
}

func (c *Context[T]) releaseDb(err error) {
	if err != nil && defaultLogger != nil {
		defaultLogger.Errorf("Step [%s] Error occurred: %v", c.flow.CurrentStepName, err)
	}

	if c.dbTx != nil {
		if err != nil {
			if err = c.dbTx.Rollback(c.Context); err != nil {
				if defaultLogger != nil {
					defaultLogger.Errorf("Step [%s] Failed to rollback transaction: %v",
						c.flow.CurrentStepName, err)
				}
			}
		} else {
			// Log commit errors but don't fail the cleanup process
			if err = c.dbTx.Commit(c.Context); err != nil {
				if defaultLogger != nil {
					defaultLogger.Errorf("Step [%s] Failed to commit transaction: %v",
						c.flow.CurrentStepName, err)
				}
			}
		}
	}

	c.cleanupDbResources()
}

// Store/Get methods for passing data between flow steps
func (c *Context[T]) Set(key string, value any) {
	c.store[key] = value
}

func (c *Context[T]) Get(key string) (any, bool) {
	value, exists := c.store[key]
	return value, exists
}

func (c *Context[T]) MustGet(key string) any {
	if value, exists := c.store[key]; exists {
		return value
	}
	panic("Key '" + key + "' does not exist in flow context")
}
