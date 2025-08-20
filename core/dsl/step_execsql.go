package dsl

type execSqlStep[TParam any] struct {
	query string
	args  []any
	BaseStep[TParam]
}

func newExecSqlStep[TParam any](query string, args ...any) *execSqlStep[TParam] {
	return &execSqlStep[TParam]{
		query: query,
		args:  args,
		BaseStep: BaseStep[TParam]{
			meta: StepMeta{
				Name: "Execute SQL",
				Kind: StepNormal,
			},
		},
	}
}

func (s *execSqlStep[TParam]) Run(ctx *FlowContext[TParam]) error {
	conn, err := ctx.GetExecutor()
	if err != nil {
		return err
	}

	// Metrics: Database operation start
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_started", map[string]string{
			"operation": "exec",
		})
	}

	result, err := conn.Exec(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		// Metrics: Database operation failed
		if ctx.serviceVar.Metrics != nil {
			ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_failed", map[string]string{
				"operation": "exec",
			})
		}
		return err
	}

	// Metrics: Database operation success
	if ctx.serviceVar.Metrics != nil {
		ctx.serviceVar.Metrics.IncCounter("dsl_db_operation_succeeded", map[string]string{
			"operation": "exec",
		})

		// Track rows affected
		rowsAffected := result.RowsAffected()
		ctx.serviceVar.Metrics.ObserveHistogram("dsl_db_rows_affected", float64(rowsAffected), map[string]string{
			"operation": "exec",
		})
	}

	if s.saveAs != "" {
		ctx.SetVar(s.saveAs, result.RowsAffected())
	}
	return nil
}

func (s *execSqlStep[TParam]) SaveAs(name string) Step[TParam] {
	s.saveAs = name
	return s
}
