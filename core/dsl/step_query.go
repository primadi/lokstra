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

	// Metrics: Database operation start
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_started", map[string]string{
			"operation": "select_many",
		})
	}

	result, err := conn.SelectManyRowMap(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		// Metrics: Database operation failed
		if ctx.serviceVar.Metrics != nil {
			ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_failed", map[string]string{
				"operation": "select_many",
			})
		}
		return err
	}

	// Metrics: Database operation success
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_succeeded", map[string]string{
			"operation": "select_many",
		})

		// Track rows returned
		ctx.serviceVar.Metrics.ObserveHistogram("dsl_db_rows_returned", float64(len(result)), map[string]string{
			"operation": "select_many",
		})
	}

	if s.saveAs != "" {
		ctx.SetVar(s.saveAs, result)
	}
	return nil
}

func (s *stepQuerySelectMany[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
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

	// Metrics: Database operation start
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_started", map[string]string{
			"operation": "select_one",
		})
	}

	result, err := conn.SelectOneRowMap(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		// Metrics: Database operation failed
		if ctx.serviceVar.Metrics != nil {
			ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_failed", map[string]string{
				"operation": "select_one",
			})
		}
		return err
	}

	// Metrics: Database operation success
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_succeeded", map[string]string{
			"operation": "select_one",
		})
	}

	if s.saveAs != "" {
		ctx.SetVar(s.saveAs, result)
	}
	return nil
}

func (s *stepQuerySelectOne[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
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
		// Create a row scanner that implements serviceapi.Row interface
		row := &rowScanner{rows: rows}
		if err := s.fn(row); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// rowScanner implements serviceapi.Row interface for stepQueryForEach
type rowScanner struct {
	rows serviceapi.Rows
}

func (r *rowScanner) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (s *stepQueryForEach[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
}

func (s *stepQueryForEach[TParam]) ForEach(fn func(serviceapi.Row) error) *Flow[TParam] {
	s.fn = fn
	// Return the flow that contains this step
	return s.flow.Then(s)
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

func (s *stepQueryErrorIfNotExists[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
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

func (s *stepQueryErrorIfExists[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
}
