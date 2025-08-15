package flow

import "github.com/primadi/lokstra/serviceapi"

type SqlStepBuilder struct {
	h *Handler
	s *sqlStep
}

func (h *Handler) ExecSql(query string, args ...any) *SqlStepBuilder {
	s := &sqlStep{
		kind:  kindExec,
		query: query,
		args:  args,
	}
	h = h.Then(s)
	return &SqlStepBuilder{
		h: h,
		s: s,
	}
}

func (h *Handler) QueryRowSql(query string, args ...any) *SqlStepBuilder {
	s := &sqlStep{
		kind:  kindQueryRow,
		query: query,
		args:  args,
	}
	h = h.Then(s)
	return &SqlStepBuilder{
		h: h,
		s: s,
	}
}

func (h *Handler) QuerySql(query string, args ...any) *SqlStepBuilder {
	s := &sqlStep{
		kind:  kindQuery,
		query: query,
		args:  args,
	}
	h = h.Then(s)
	return &SqlStepBuilder{
		h: h,
		s: s,
	}
}

// ExecReturning is a convenience method for EXEC queries with RETURNING clause
func (h *Handler) ExecReturning(query string, args ...any) *SqlStepBuilder {
	s := &sqlStep{
		kind:         kindExec,
		query:        query,
		args:         args,
		hasReturning: true,
	}
	h = h.Then(s)
	return &SqlStepBuilder{
		h: h,
		s: s,
	}
}

func (b *SqlStepBuilder) Args(args ...any) *SqlStepBuilder {
	b.s.args = args
	return b
}

func (b *SqlStepBuilder) ArgsFn(fn func(*Context) []any) *SqlStepBuilder {
	b.s.argsFn = fn
	return b
}

// WithName sets a custom name for this SQL step (useful for debugging/telemetry)
func (b *SqlStepBuilder) WithName(name string) *SqlStepBuilder {
	b.s.customName = name
	return b
}

// WithReturning marks this EXEC query as having a RETURNING clause
// This changes the execution strategy to capture returned data instead of just RowsAffected
func (b *SqlStepBuilder) WithReturning() *SqlStepBuilder {
	if b.s.kind != kindExec {
		panic("WithReturning() can only be used with ExecSql queries")
	}
	b.s.hasReturning = true
	return b
}

// Guards (EXEC)

func (b *SqlStepBuilder) AffectOne() *SqlStepBuilder {
	b.s.guardExactOne = true
	return b
}

func (b *SqlStepBuilder) AffectAtLeast(n int64) *SqlStepBuilder {
	b.s.guardAtLeast = &n
	return b
}

func (b *SqlStepBuilder) AffectAtMost(n int64) *SqlStepBuilder {
	b.s.guardAtMost = &n
	return b
}

func (b *SqlStepBuilder) AffectBetween(atLeast, atMost int64) *SqlStepBuilder {
	if atLeast > atMost {
		panic("atLeast cannot be greater than atMost")
	}
	b.s.guardAtLeast = &atLeast
	b.s.guardAtMost = &atMost
	return b
}

// Guards (QUERY)

func (b *SqlStepBuilder) EnsureExists(err error) *SqlStepBuilder {
	if b.s.ensureNotExists != nil {
		panic("ensureNotExists is already set, cannot set ensureExists")
	}
	b.s.ensureExists = err
	return b
}

func (b *SqlStepBuilder) EnsureNotExists(err error) *SqlStepBuilder {
	if b.s.ensureExists != nil {
		panic("ensureExists is already set, cannot set ensureNotExists")
	}
	b.s.ensureNotExists = err
	return b
}

// Finishing SqlTepBuilder

func (b *SqlStepBuilder) ScanTo(fn func(serviceapi.Row) error) *Handler {
	b.s.scanTo = fn
	return b.h
}

func (b *SqlStepBuilder) Collect(fn func(serviceapi.Rows) error) *Handler {
	b.s.collect = fn
	return b.h
}

func (b *SqlStepBuilder) ForEach(fn func(serviceapi.Row) error) *Handler {
	b.s.forEach = fn
	return b.h
}

func (b *SqlStepBuilder) SaveAs(name string) *Handler {
	b.s.saveAs = name
	return b.h
}

func (b *SqlStepBuilder) Done() *Handler { return b.h }
