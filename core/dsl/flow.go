package dsl

import (
	"time"

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

func (f *Flow[TParam]) QuerySaveAs(query string, saveAs string, args ...any) *Flow[TParam] {
	step := newStepQuerySelectMany[TParam](query, args...)
	step.SaveAs(saveAs)
	return f.Then(step)
}

func (f *Flow[TParam]) QueryOneSaveAs(query string, saveAs string, args ...any) *Flow[TParam] {
	step := newStepQuerySelectOne[TParam](query, args...)
	step.SaveAs(saveAs)
	return f.Then(step)
}

func (f *Flow[TParam]) ExecSqlSaveAs(query string, saveAs string, args ...any) *Flow[TParam] {
	step := newExecSqlStep[TParam](query, args...)
	step.SaveAs(saveAs)
	return f.Then(step)
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

func (f *Flow[TParam]) Do(fn func(*FlowContext[TParam]) error) *Flow[TParam] {
	return f.Then(newStepCustom(fn))
}

func (f *Flow[TParam]) If(condition func(*FlowContext[TParam]) bool, thenStep Step[TParam]) *Flow[TParam] {
	return f.Then(newStepConditional(condition, thenStep))
}

func (f *Flow[TParam]) Retry(step Step[TParam], maxRetries int) *Flow[TParam] {
	return f.Then(newStepRetry(step, maxRetries))
}

func (f *Flow[TParam]) Validate(validationFn func(*FlowContext[TParam]) error) *Flow[TParam] {
	return f.Then(ValidateStep(validationFn))
}

func (f *Flow[TParam]) Error(err error) *Flow[TParam] {
	return f.Then(ErrorStep[TParam](err))
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

// getStepKindString converts StepKind to string for metrics
func (f *Flow[TParam]) getStepKindString(kind StepKind) string {
	switch kind {
	case StepNormal:
		return "normal"
	case StepTxBegin:
		return "tx_begin"
	case StepTxEnd:
		return "tx_end"
	case StepAfterCommit:
		return "after_commit"
	case StepFinalize:
		return "finalize"
	default:
		return "unknown"
	}
}

func (f *Flow[TParam]) Run(reqContext *request.Context) error {
	flowCtx := NewFlowContext(reqContext, f.sv)

	// Metrics: Flow execution start
	if f.sv.Metrics != nil {
		f.sv.Metrics.IncCounter("dsl_flow_started", serviceapi.Labels{
			"flow_name": f.name,
		})
	}

	flowStartTime := time.Now()

	defer func() {
		// Always cleanup resources
		if err := flowCtx.Cleanup(); err != nil {
			// Log cleanup error but don't override main error
			if f.sv.Logger != nil {
				f.sv.Logger.Errorf("Failed to cleanup flow context: %v", err)
			}
		}

		// Metrics: Flow execution duration
		if f.sv.Metrics != nil {
			duration := time.Since(flowStartTime).Seconds()
			f.sv.Metrics.ObserveHistogram("dsl_flow_duration_seconds", duration, serviceapi.Labels{
				"flow_name": f.name,
			})
		}
	}()

	for i, step := range f.steps {
		stepStartTime := time.Now()
		stepName := step.Meta().Name

		// Log step execution start
		if f.sv.Logger != nil {
			f.sv.Logger.Debugf("Executing step '%s' at index %d", stepName, i)
		}

		// Metrics: Step execution start
		if f.sv.Metrics != nil {
			f.sv.Metrics.IncCounter("dsl_step_started", serviceapi.Labels{
				"flow_name": f.name,
				"step_name": stepName,
				"step_kind": f.getStepKindString(step.Meta().Kind),
			})
		}

		var err error
		if err = f.RunStep(step, flowCtx); err != nil {
			// Log error and set in context
			if f.sv.Logger != nil {
				f.sv.Logger.Errorf("Step execution failed '%s' at index %d: %v", stepName, i, err)
			}
			flowCtx.SetError(err)

			// Metrics: Step execution failed
			if f.sv.Metrics != nil {
				f.sv.Metrics.IncCounter("dsl_step_failed", serviceapi.Labels{
					"flow_name": f.name,
					"step_name": stepName,
					"step_kind": f.getStepKindString(step.Meta().Kind),
				})

				f.sv.Metrics.IncCounter("dsl_flow_failed", serviceapi.Labels{
					"flow_name": f.name,
				})
			}

			// If we have a transaction, mark for rollback
			if flowCtx.dbTx != nil {
				f.forceRollback = true
			}

			return err
		}

		// Metrics: Step execution success
		if f.sv.Metrics != nil {
			stepDuration := time.Since(stepStartTime).Seconds()
			f.sv.Metrics.ObserveHistogram("dsl_step_duration_seconds", stepDuration, serviceapi.Labels{
				"flow_name": f.name,
				"step_name": stepName,
				"step_kind": f.getStepKindString(step.Meta().Kind),
			})

			f.sv.Metrics.IncCounter("dsl_step_succeeded", serviceapi.Labels{
				"flow_name": f.name,
				"step_name": stepName,
				"step_kind": f.getStepKindString(step.Meta().Kind),
			})
		}

		// Log successful execution
		if f.sv.Logger != nil {
			f.sv.Logger.Debugf("Step executed successfully '%s' at index %d", stepName, i)
		}
	}

	// Metrics: Flow execution success
	if f.sv.Metrics != nil {
		f.sv.Metrics.IncCounter("dsl_flow_succeeded", serviceapi.Labels{
			"flow_name": f.name,
		})
	}

	return nil
}
