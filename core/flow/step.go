package flow

type StepKind int

const (
	StepNormal StepKind = iota
	StepTxBegin
	StepTxEnd
	StepAfterCommit
	StepRespond
)

// Meta info so executor can orchestrate (tx, after-commit, respond, tracing)
type StepMeta struct {
	Name string
	Kind StepKind
}

type Step interface {
	Run(*Context) error // do the work
	Meta() StepMeta     // used by executor for orchestration/telemetry
}
