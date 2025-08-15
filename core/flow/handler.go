package flow

import (
	"context"
	"fmt"

	"github.com/primadi/lokstra/core/iface"
	"github.com/primadi/lokstra/serviceapi"
)

// Handler is a build-time container of steps (immutable at runtime).
type Handler struct {
	// build-time deps & defaults
	regCtx           iface.RegistrationContext
	requiredServices []string
	defaultDbService string
	defaultDbSchema  string

	// pipeline (ordered)
	steps []Step

	name string

	stdContext context.Context // standard context for steps

}

// NewHandler creates a new flow builder.
func NewHandler(regCtx iface.RegistrationContext, name string) *Handler {
	return &Handler{regCtx: regCtx, name: name}
}

func (h *Handler) Then(step Step) *Handler {
	h.steps = append(h.steps, step)
	return h
}

func (h *Handler) Steps(steps ...Step) *Handler {
	h.steps = append(h.steps, steps...)
	return h
}

func (h *Handler) SetDbService(name string) *Handler {
	h.defaultDbService = name
	return h
}

func (h *Handler) SetDbSchema(schema string) *Handler {
	h.defaultDbSchema = schema
	return h
}

func (h *Handler) SetRequiredServices(services ...string) *Handler {
	h.requiredServices = services
	return h
}

// BeginTx adds a transaction begin step to the handler pipeline.
//
// IMPORTANT: Nested transactions are not supported and will cause:
// - Build-time panic if multiple BeginTx() calls are made on the same handler
// - Runtime error if a transaction is already active in the context
//
// Use separate handlers for separate transactions instead of nesting.
func (h *Handler) BeginTx() *Handler {
	// Check for existing transaction begin step (fail fast at build time)
	for _, step := range h.steps {
		if meta := step.Meta(); meta.Kind == StepTxBegin {
			panic("multiple BeginTx() calls detected: nested transactions not supported")
		}
	}

	step := &txBeginStep{}
	h.steps = append(h.steps, step)
	return h
}

// CommitOrRollback adds a transaction end step to the handler pipeline
func (h *Handler) CommitOrRollback() *Handler {
	step := &txEndStep{}
	h.steps = append(h.steps, step)
	return h
}

// Rollback adds a transaction rollback step to the handler pipeline
func (h *Handler) Rollback() *Handler {
	step := &txEndStep{forceRollback: true}
	h.steps = append(h.steps, step)
	return h
}

// GetDbPool resolves database pool from defaultDbService via RegistrationContext
func (h *Handler) GetDbPool() (serviceapi.DbPool, error) {
	if h.defaultDbService == "" {
		return nil, fmt.Errorf("no default database service configured")
	}

	service, err := h.regCtx.GetService(h.defaultDbService)
	if err != nil {
		return nil, fmt.Errorf("failed to get database service '%s': %w", h.defaultDbService, err)
	}

	dbPool, ok := service.(serviceapi.DbPool)
	if !ok {
		return nil, fmt.Errorf("service '%s' does not implement DbPool interface", h.defaultDbService)
	}

	return dbPool, nil
}

// GetDbSchema returns the configured default database schema
func (h *Handler) GetDbSchema() string {
	return h.defaultDbSchema
}

// Do adds a custom function step to the handler pipeline.
// This allows you to execute custom logic within the flow.
func (h *Handler) Do(fn func(*Context) error) *Handler {
	step := &customStep{fn: fn, name: "custom.function"}
	h.steps = append(h.steps, step)
	return h
}

// DoNamed adds a custom function step with a specific name to the handler pipeline.
// This allows you to execute custom logic with a meaningful name for debugging/telemetry.
func (h *Handler) DoNamed(name string, fn func(*Context) error) *Handler {
	step := &customStep{fn: fn, name: name}
	h.steps = append(h.steps, step)
	return h
}
