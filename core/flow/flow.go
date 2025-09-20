package flow

import (
	"github.com/primadi/lokstra/core/registration"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

// HandledError represents an error that has already been handled (response already set)
// This signals to AsHandler that it should return nil instead of the error
type HandledError struct {
	Message string
}

func (e HandledError) Error() string {
	return e.Message
}

type Flow[T any] struct {
	name  string
	steps []Step[T]

	DbPool          serviceapi.DbPool
	DbSchemaName    string
	CurrentStepName string
}

type StepAction[T any] = func(flowCtx *Context[T]) error

type Step[T any] struct {
	name   string
	action StepAction[T]
}

// NewStep creates a new step with given name and action
func NewStep[T any](name string, action StepAction[T]) Step[T] {
	return Step[T]{
		name:   name,
		action: action,
	}
}

func NewFlow[T any](name string) *Flow[T] {
	return &Flow[T]{
		name:  name,
		steps: []Step[T]{},
	}
}

func (f *Flow[T]) AddAction(name string, action StepAction[T]) *Flow[T] {
	f.steps = append(f.steps, Step[T]{
		name:   name,
		action: action,
	})
	return f
}

func (f *Flow[T]) AddSteps(steps ...Step[T]) *Flow[T] {
	f.steps = append(f.steps, steps...)
	return f
}

func (f *Flow[T]) SetDbSchemaName(schemaName string) *Flow[T] {
	f.DbSchemaName = schemaName
	return f
}

func (f *Flow[T]) SetDbPoolService(dbPool serviceapi.DbPool) *Flow[T] {
	f.DbPool = dbPool
	return f
}

func (f *Flow[T]) SetDbPool(regCtx registration.Context, name string) *Flow[T] {
	var err error

	f.DbPool, err = registration.GetService[serviceapi.DbPool](regCtx, name)
	if err != nil {
		panic(err)
	}

	return f
}

// -------------------

func (f *Flow[T]) run(flowCtx *Context[T]) error {
	var stepErr error

	for _, step := range f.steps {
		f.CurrentStepName = step.name
		if err := step.action(flowCtx); err != nil {
			stepErr = err
			break
		}
	}

	// Always release database resources (commit if no error, rollback if error)
	flowCtx.releaseDb(stepErr)

	if stepErr != nil {
		// Check if this is a handled error (response already set)
		// If so, return nil to indicate successful HTTP handling
		if _, isHandled := stepErr.(HandledError); isHandled {
			return nil
		}

		// Unhandled error - let it bubble up
		return stepErr
	}

	return nil
}

func (f *Flow[T]) AsHandler() request.HandlerFunc {
	return func(reqCtx *request.Context) error {
		var params T

		err := reqCtx.BindAll(&params)
		if err != nil {
			return err
		}
		flowCtx := newContext(f, reqCtx)
		flowCtx.Params = &params
		return f.run(flowCtx)
	}
}

// AsHandlerSmart creates a handler with smart content-type binding
// Supports both JSON and form-urlencoded request bodies
func (f *Flow[T]) AsHandlerSmart() request.HandlerFunc {
	return func(reqCtx *request.Context) error {
		var params T

		err := reqCtx.BindAllSmart(&params)
		if err != nil {
			return err
		}
		flowCtx := newContext(f, reqCtx)
		flowCtx.Params = &params
		return f.run(flowCtx)
	}
}
