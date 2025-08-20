package dsl

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

type FlowContext[TParam any] struct {
	reqCtx *request.Context

	dbConn     serviceapi.DbConn
	dbTx       serviceapi.DbTx
	serviceVar *ServiceVar[TParam]

	// Error handling
	lastError error

	// Resource cleanup tracking
	needsCleanup bool
}

func NewFlowContext[TParam any](reqCtx *request.Context,
	serviceVar *ServiceVar[TParam]) *FlowContext[TParam] {
	return &FlowContext[TParam]{
		reqCtx:       reqCtx,
		serviceVar:   serviceVar,
		needsCleanup: false,
	}
}

func (ctx *FlowContext[TParam]) GetExecutor() (serviceapi.DbExecutor, error) {
	if ctx.dbTx != nil {
		return ctx.dbTx, nil
	}
	if ctx.dbConn == nil {
		var err error
		ctx.dbConn, err = ctx.serviceVar.DbPool.Acquire(ctx.reqCtx.Context,
			ctx.serviceVar.DbSchemaName)
		if err != nil {
			return nil, err
		}
		ctx.needsCleanup = true
	}
	return ctx.dbConn, nil
}

func (ctx *FlowContext[TParam]) SetVar(name string, value any) {
	if ctx.serviceVar.Vars == nil {
		ctx.serviceVar.Vars = make(map[string]any)
	}
	ctx.serviceVar.Vars[name] = value
}

func (ctx *FlowContext[TParam]) GetVar(name string) (any, bool) {
	value, ok := ctx.serviceVar.Vars[name]
	return value, ok
}

// GetParam returns the service parameters
func (ctx *FlowContext[TParam]) GetParam() *TParam {
	return ctx.serviceVar.Param
}

// GetContext returns the request context
func (ctx *FlowContext[TParam]) GetContext() *request.Context {
	return ctx.reqCtx
}

// SetError sets the last error in context
func (ctx *FlowContext[TParam]) SetError(err error) {
	ctx.lastError = err
}

// GetError returns the last error
func (ctx *FlowContext[TParam]) GetError() error {
	return ctx.lastError
}

// Cleanup releases resources
func (ctx *FlowContext[TParam]) Cleanup() error {
	if ctx.needsCleanup && ctx.dbConn != nil {
		return ctx.dbConn.Release()
	}
	return nil
}

// GetLocalizedMessage returns a localized message using I18n service
func (ctx *FlowContext[TParam]) GetLocalizedMessage(code string, params map[string]any) string {
	if ctx.serviceVar.I18n == nil {
		return code // Fallback to code if I18n not available
	}

	// Try to get language from request context
	lang := "en" // Default language
	if ctx.reqCtx != nil {
		acceptLang := ctx.reqCtx.GetHeader("Accept-Language")
		if acceptLang != "" && len(acceptLang) >= 2 {
			lang = acceptLang[:2] // Simple extraction, could be more sophisticated
		}
	}

	return ctx.serviceVar.I18n.T(lang, code, params)
}
