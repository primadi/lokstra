package dsl

import (
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

type FlowPhase = int

const (
	PhaseValidation FlowPhase = iota
	PhaseExecution
	PhasePostCommit
	PhaseFinalization
)

type Flow[TParam any] struct {
	name  string
	steps []Step[TParam]
	sv    *ServiceVar[TParam]

	dbConn        serviceapi.DbConn
	dbTx          serviceapi.DbTx
	forceRollback bool

	phase FlowPhase
}

func NewFlow[TParam any](name string, serviceVars *ServiceVar[TParam]) *Flow[TParam] {
	return &Flow[TParam]{
		name:  name,
		sv:    serviceVars,
		phase: PhaseValidation,
	}
}

func (f *Flow[TParam]) Then(step Step[TParam]) *Flow[TParam] {
	switch f.phase {
	case PhaseValidation:
		switch step.Meta().Kind {
		case StepTxBegin:
			f.phase = PhaseExecution
		case StepTxEnd:
			panic("cannot add StepTxEnd in validation phase")
		case StepAfterCommit:
			panic("cannot add StepAfterCommit in validation phase")
		case StepFinalize:
			f.phase = PhaseFinalization
		}
	case PhaseExecution:
		switch step.Meta().Kind {
		case StepTxBegin:
			panic("cannot add StepTxBegin in execution phase")
		case StepTxEnd:
			f.phase = PhasePostCommit
		case StepAfterCommit:
			panic("cannot add StepAfterCommit in execution phase")
		case StepFinalize:
			panic("cannot add StepFinalize in execution phase")
		}
	case PhasePostCommit:
		switch step.Meta().Kind {
		case StepTxBegin:
			panic("cannot add StepTxBegin in post-commit phase")
		case StepTxEnd:
			panic("cannot add StepTxEnd in post-commit phase")
		case StepFinalize:
			f.phase = PhaseFinalization
		}
	case PhaseFinalization:
		switch step.Meta().Kind {
		case StepTxBegin:
			panic("cannot add StepTxBegin in finalization phase")
		case StepTxEnd:
			panic("cannot add StepTxEnd in finalization phase")
		case StepAfterCommit:
			panic("cannot add StepAfterCommit in finalization phase")
		}
	}
	f.steps = append(f.steps, step)
	return f
}

func (f *Flow[TParam]) Steps(step ...Step[TParam]) *Flow[TParam] {
	for _, s := range step {
		f.Then(s)
	}
	return f
}

func (f *Flow[TParam]) Name() string {
	return f.name
}

func (f *Flow[TParam]) GetSteps() []Step[TParam] {
	return f.steps
}

func (f *Flow[TParam]) BeginTx() *Flow[TParam] {
	return f.Then(&BaseStep[TParam]{
		meta: StepMeta{
			Name: "Begin Transaction",
			Kind: StepTxBegin,
		},
	})
}

func (f *Flow[TParam]) CommitOrRollback() *Flow[TParam] {
	return f.Then(&BaseStep[TParam]{
		meta: StepMeta{
			Name: "Commit or Rollback Transaction",
			Kind: StepTxEnd,
		},
	})
}

func (f *Flow[TParam]) ExecSql(query string, args ...any) *Flow[TParam] {
	return f.Then(newExecSqlStep[TParam](query, args...))
}

func (f *Flow[TParam]) Query(query string, args ...any) *Flow[TParam] {
	return f.Then(newStepQuerySelectMany[TParam](query, args...))
}

func (f *Flow[TParam]) QueryOne(query string, args ...any) *Flow[TParam] {
	return f.Then(newStepQuerySelectOne[TParam](query, args...))
}

func (f *Flow[TParam]) QueryForEach(query string, args ...any) stepQueryForEachIface[TParam] {
	return newStepQueryForEach(f, query, args...)
}

func (f *Flow[TParam]) ErrorIfExists(err error, query string, args ...any) *Flow[TParam] {
	return f.Then(newStepQueryErrorIfExists[TParam](err, query, args...))
}

func (f *Flow[TParam]) ErrorIfNotExists(err error, query string, args ...any) *Flow[TParam] {
	return f.Then(newStepQueryErrorIfNotExists[TParam](err, query, args...))
}

func (f *Flow[TParam]) RunStep(step Step[TParam], flowCtx *FlowContext[TParam]) error {
	switch step.Meta().Kind {
	case StepTxBegin:
		if flowCtx.dbConn == nil {
			flowCtx.dbConn, _ = f.sv.DbPool.Acquire(flowCtx.reqCtx.Context,
				f.sv.DbSchemaName)
		}
		flowCtx.dbTx, _ = flowCtx.dbConn.Begin(flowCtx.reqCtx.Context)
	case StepTxEnd:
		if f.forceRollback {
			if err := flowCtx.dbTx.Rollback(flowCtx.reqCtx.Context); err != nil {
				return err
			}
		} else {
			if err := flowCtx.dbTx.Commit(flowCtx.reqCtx.Context); err != nil {
				return err
			}
		}
	}

	return step.Run(flowCtx)
}

func (f *Flow[TParam]) Run(reqContext *request.Context) error {
	flowCtx := NewFlowContext(reqContext, f.sv)
	for _, step := range f.steps {
		var err error
		if err = f.RunStep(step, flowCtx); err != nil {
			return err
		}
	}
	return nil
}
