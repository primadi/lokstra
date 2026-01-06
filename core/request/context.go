package request

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/primadi/lokstra/core/response"
	"github.com/primadi/lokstra/serviceapi"
)

// ConfigResolver is a function type for resolving config values
// Used to avoid circular dependency with lokstra_registry
type ConfigResolver func(key string, defaultValue any) any

// Global config resolver set by lokstra_registry at initialization
var globalConfigResolver ConfigResolver

// SetConfigResolver sets the global config resolver
// Called by lokstra_registry during initialization to avoid circular dependency
func SetConfigResolver(resolver ConfigResolver) {
	globalConfigResolver = resolver
}

// StatusCodeError represents an error state based on HTTP status code
// Used to trigger transaction rollback for non-2xx status codes
type StatusCodeError struct {
	StatusCode int
}

func (e *StatusCodeError) Error() string {
	return fmt.Sprintf("HTTP status code %d indicates error", e.StatusCode)
}

type Context struct {
	// Embedding standard context for easy access
	context.Context

	// Helper to access request methods and fields
	Req *RequestHelper
	// Helper to access response methods and fields
	Resp *response.Response
	// Helper for opinionated API responses (wraps data in ApiResponse)
	Api *response.ApiHelper

	// Direct access to primitives (for advanced usage)
	W *writerWrapper
	R *http.Request

	// Internal index to track middleware/handler execution
	index    int
	handlers []HandlerFunc

	value map[string]any

	// Transaction finalizers to be called automatically in FinalizeResponse
	// Map of poolName -> finalizer function
	txFinalizers map[string]func(*error)
	// Track order of transaction creation for proper LIFO finalization
	txPoolOrder []string
}

func NewContext(w http.ResponseWriter, r *http.Request, handlers []HandlerFunc) *Context {
	api := response.NewApiHelper()

	ctx := &Context{
		Context:  context.Background(),
		W:        newWriterWrapper(w),
		R:        r,
		handlers: handlers,
		Resp:     api.Resp(), // Direct assignment to Resp
		Api:      api,        // Initialize API helper
	}

	// Initialize request helper
	ctx.Req = newRequestHelper(ctx)

	return ctx
}

// Call inside middleware
func (c *Context) Next() error {
	if c.index >= len(c.handlers) {
		return nil
	}
	h := c.handlers[c.index]
	c.index++
	return h(c)
}

// Begins a transaction for the specified pool name
// The transaction will be automatically finalized (commit/rollback) when FinalizeResponse is called
// No need to defer the returned function anymore - it's handled automatically
//
// Supports indirection with @ prefix:
//   - BeginTransaction("my-pool")           - Direct pool name
//   - BeginTransaction("@auth.db-pool")     - Pool name from config
//
// Example with indirection:
//
//	config.yaml:
//	  configs:
//	    auth:
//	      db-pool: "postgres-auth-pool"
//
//	BeginTransaction("@auth.db-pool")  // Resolves to "postgres-auth-pool"
func (c *Context) BeginTransaction(poolName string) {
	// Resolve indirection if @ prefix is present
	actualPoolName := poolName
	if strings.HasPrefix(poolName, "@") && globalConfigResolver != nil {
		configKey := strings.TrimPrefix(poolName, "@")
		if resolvedName, ok := globalConfigResolver(configKey, "").(string); ok && resolvedName != "" {
			actualPoolName = resolvedName
		} else {
			// Config key not found or not a string, use original (will likely fail at runtime)
			// This matches the behavior of @Inject where invalid config keys fail at runtime
			actualPoolName = poolName
		}
	}

	newCtx, finalizeCtx := serviceapi.BeginTransaction(c.Context, actualPoolName)
	c.Context = newCtx // Update embedded context with transaction context

	// Initialize map if needed
	if c.txFinalizers == nil {
		c.txFinalizers = make(map[string]func(*error))
	}

	// Store finalizer by actual pool name (not the @ prefixed name)
	c.txFinalizers[actualPoolName] = finalizeCtx
	c.txPoolOrder = append(c.txPoolOrder, actualPoolName)
}

// RollbackTransaction manually rolls back a specific transaction
// Use this for edge cases like dry-run or testing where you want to return 200 OK but rollback changes
//
// ⚠️ WARNING: Only call this in the SAME handler that called BeginTransaction.
// Do NOT use manual commit/rollback if your handler calls other handlers (nested calls),
// as it can cause transaction state inconsistency.
//
// Safe usage:
//   - Dry-run operations (single handler, no nested calls)
//   - Testing/validation (single handler, no nested calls)
//
// Unsafe usage:
//   - Handler A commits manually, then calls Handler B (nested)
//   - Service layer that might be called by multiple handlers
func (c *Context) RollbackTransaction(poolName string) {
	if finalizer, exists := c.txFinalizers[poolName]; exists {
		// Force rollback by passing a non-nil error
		rollbackErr := error(&StatusCodeError{StatusCode: http.StatusInternalServerError})
		finalizer(&rollbackErr)
		// Remove from finalizers to prevent double-finalization
		delete(c.txFinalizers, poolName)
		// Also remove from order tracking
		c.removeTxFromOrder(poolName)
	}
}

// CommitTransaction manually commits a specific transaction
// Use this for edge cases where you need explicit control over transaction lifecycle
//
// ⚠️ WARNING: Only call this in the SAME handler that called BeginTransaction.
// Do NOT use manual commit/rollback if your handler calls other handlers (nested calls),
// as it can cause transaction state inconsistency.
//
// Safe usage:
//   - Conditional commit based on business logic (single handler)
//   - Partial success handling (single handler, no nested calls)
//
// Unsafe usage:
//   - Handler A commits manually, then calls Handler B (nested)
//   - Service layer that might be called by multiple handlers
func (c *Context) CommitTransaction(poolName string) {
	if finalizer, exists := c.txFinalizers[poolName]; exists {
		// Force commit by passing nil error
		var commitErr error
		finalizer(&commitErr)
		// Remove from finalizers to prevent double-finalization
		delete(c.txFinalizers, poolName)
		// Also remove from order tracking
		c.removeTxFromOrder(poolName)
	}
}

// removeTxFromOrder removes a pool name from the order tracking
func (c *Context) removeTxFromOrder(poolName string) {
	for i, name := range c.txPoolOrder {
		if name == poolName {
			c.txPoolOrder = append(c.txPoolOrder[:i], c.txPoolOrder[i+1:]...)
			break
		}
	}
}

// Finalizes the response, writing status code and body if not already written
// Also automatically finalizes all transactions (commit on success, rollback on error)
func (c *Context) FinalizeResponse(err error) {
	// IMPORTANT: Always finalize transactions, even if response was manually written
	// Use defer to ensure transactions are finalized in all code paths
	defer func() {
		// Finalize all remaining transactions (in reverse order - LIFO)
		// Skip transactions that were manually committed/rolled back
		// Determine if transaction should commit or rollback based on:
		// 1. Explicit error from handler
		// 2. Status code >= 400 (client/server errors)
		statusCode := c.StatusCode()
		var txErr error
		if err != nil {
			txErr = err
		} else if statusCode >= http.StatusBadRequest {
			// Create error from status code to trigger rollback
			txErr = &StatusCodeError{StatusCode: statusCode}
		}

		// Call finalizers in reverse order (LIFO) for remaining transactions
		for i := len(c.txPoolOrder) - 1; i >= 0; i-- {
			poolName := c.txPoolOrder[i]
			if finalizer, exists := c.txFinalizers[poolName]; exists {
				finalizer(&txErr)
			}
		}
	}()

	if c.W.ManualWritten() {
		// User already wrote directly to ResponseWriter, skip response writing
		// but still finalize transactions (via defer above)
		return
	}

	if err != nil {
		// Check if error is ValidationError
		if valErr, ok := err.(*ValidationError); ok {
			// Use Api helper to format validation error properly
			c.Api.ValidationError("Validation failed", valErr.FieldErrors)
		} else {
			// Handle other errors
			st := c.Resp.RespStatusCode
			if st == 0 || st < http.StatusBadRequest {
				c.Api.InternalError(err.Error())
				// c.Resp.WithStatus(http.StatusInternalServerError).
				//   Json(map[string]string{"error": err.Error()})
			}
		}
	}

	c.Resp.WriteHttp(c.W)
}

func (c *Context) executeHandler() error {
	return c.Next()
}

// Adds a value to the context storage
func (c *Context) Set(key string, value any) {
	if c.value == nil {
		c.value = make(map[string]any)
	}
	c.value[key] = value
}

// Retrieves a value from the context storage
func (c *Context) Get(key string) any {
	return c.value[key]
}

// Adds a value to the context
type contextKey string

func (c *Context) SetContextValue(key string, value any) {
	c.Context = context.WithValue(c.Context, contextKey(key), value)
}

// Retrieves a value from the context
func (c *Context) GetContextValue(key string) any {
	if c.Context == nil {
		return nil
	}
	return c.Context.Value(contextKey(key))
}

// StatusCode returns the HTTP status code from the response
// It checks multiple sources in order of priority:
// 1. Writer's status code (if manually written)
// 2. Response helper's status code (if set via Api/Resp)
// 3. Default 200 OK
func (c *Context) StatusCode() int {
	// First check writer's status code (manual writes)
	ret := c.W.StatusCode()
	if ret == 0 {
		// Then check response helper's status code
		ret = c.Resp.RespStatusCode
	}
	if ret == 0 {
		// Default to 200 OK
		ret = 200
	}
	return ret
}
