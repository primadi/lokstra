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
}

func NewFlowContext[TParam any](reqCtx *request.Context,
	serviceVar *ServiceVar[TParam]) *FlowContext[TParam] {
	return &FlowContext[TParam]{
		reqCtx:     reqCtx,
		serviceVar: serviceVar,
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
