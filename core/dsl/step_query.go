package dsl

import "github.com/primadi/lokstra/serviceapi"

type stepQuerySelectMany[TParam any] struct {
	query string
	args  []any
	BaseStep[TParam]
}

func newStepQuerySelectMany[TParam any](query string, args ...any) *stepQuerySelectMany[TParam] {
	return &stepQuerySelectMany[TParam]{
		query: query,
		args:  args,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Execute Query Select Many",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepQuerySelectMany[TParam]) Run(ctx *FlowContext[TParam]) error {
	conn, err := ctx.GetExecutor()
	if err != nil {
		return err
	}
	result, err := conn.SelectManyRowMap(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		return err
	}
	if s.saveAs != "" {
		ctx.SetVar(s.saveAs, result)
	}
	return nil
}

// ---------------

type stepQuerySelectOne[TParam any] struct {
	query string
	args  []any
	BaseStep[TParam]
}

func newStepQuerySelectOne[TParam any](query string, args ...any) *stepQuerySelectOne[TParam] {
	return &stepQuerySelectOne[TParam]{
		query: query,
		args:  args,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Execute Query Select One",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepQuerySelectOne[TParam]) Run(ctx *FlowContext[TParam]) error {
	conn, err := ctx.GetExecutor()
	if err != nil {
		return err
	}
	result, err := conn.SelectOneRowMap(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		return err
	}
	if s.saveAs != "" {
		ctx.SetVar(s.saveAs, result)
	}
	return nil
}

// --------------------

type stepQueryForEach[TParam any] struct {
	flow             *Flow[TParam]
	query            string
	args             []any
	fn               func(serviceapi.Row) error
	BaseStep[TParam] // BaseStep is used to implement the Step interface}]
}

func newStepQueryForEach[TParam any](flow *Flow[TParam], query string, args ...any) *stepQueryForEach[TParam] {
	return &stepQueryForEach[TParam]{
		flow:  flow,
		query: query,
		args:  args,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Execute Query For Each",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepQueryForEach[TParam]) Run(ctx *FlowContext[TParam]) error {
	conn, err := ctx.GetExecutor()
	if err != nil {
		return err
	}
	rows, err := conn.Query(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row serviceapi.Row
		if err := rows.Scan(&row); err != nil {
			return err
		}
		if err := s.fn(row); err != nil {
			return err
		}
	}
	return nil
}

func (s *stepQueryForEach[TParam]) ForEach(fn func(serviceapi.Row) error) *Flow[TParam] {
	s.fn = fn
	return &Flow[TParam]{
		steps: []Step[TParam]{s},
	}
}

type stepQueryForEachIface[TParam any] interface {
	ForEach(fn func(serviceapi.Row) error) *Flow[TParam]
}

// --------------------

type stepQueryErrorIfNotExists[TParam any] struct {
	query string
	args  []any
	err   error
	BaseStep[TParam]
}

func newStepQueryErrorIfNotExists[TParam any](err error, query string,
	args ...any) *stepQueryErrorIfNotExists[TParam] {
	return &stepQueryErrorIfNotExists[TParam]{
		err:   err,
		query: query,
		args:  args,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Execute Query Error No Rows",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepQueryErrorIfNotExists[TParam]) Run(ctx *FlowContext[TParam]) error {
	conn, err := ctx.GetExecutor()
	if err != nil {
		return err
	}
	exists, err := conn.IsExists(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		return err
	}
	if !exists {
		return s.err
	}

	return nil
}

// --------------------

type stepQueryErrorIfExists[TParam any] struct {
	query string
	args  []any
	err   error
	BaseStep[TParam]
}

func newStepQueryErrorIfExists[TParam any](err error, query string,
	args ...any) *stepQueryErrorIfExists[TParam] {
	return &stepQueryErrorIfExists[TParam]{
		err:   err,
		query: query,
		args:  args,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Execute Query Error Has Rows",
				Kind: StepNormal,
			},
		},
	}
}

func (s *stepQueryErrorIfExists[TParam]) Run(ctx *FlowContext[TParam]) error {
	conn, err := ctx.GetExecutor()
	if err != nil {
		return err
	}
	exists, err := conn.IsExists(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		return err
	}
	if exists {
		return s.err
	}

	return nil
}
