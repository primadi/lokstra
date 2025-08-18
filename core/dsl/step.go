package dsl

type StepKind = int

const (
	StepNormal StepKind = iota
	StepTxBegin
	StepTxEnd
	StepAfterCommit
	StepFinalize
)

type StepMeta struct {
	Name string
	Kind StepKind
}

type Step[TParam any] interface {
	Run(ctx *FlowContext[TParam]) error // do the work
	Meta() StepMeta                     // used by executor for orchestration/telemetry
	SaveAs(name string) Step[TParam]    // save result to FlowContext.Vars[name]
}

type BaseStep[TParam any] struct {
	meta   StepMeta
	saveAs string
}

// Meta implements Step.
func (b *BaseStep[TParam]) Meta() StepMeta {
	return b.meta
}

// Run implements Step.
func (b *BaseStep[TParam]) Run(ctx *FlowContext[TParam]) error {
	return nil
}

// SaveAs implements Step.
func (b *BaseStep[TParam]) SaveAs(name string) Step[TParam] {
	b.saveAs = name
	return b
}
