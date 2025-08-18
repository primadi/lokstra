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
	result, err := conn.Exec(ctx.reqCtx.Context, s.query, s.args...)
	if err != nil {
		return err
	}
	if s.saveAs != "" {
		ctx.SetVar(s.saveAs, result.RowsAffected())
	}
	return nil
}
