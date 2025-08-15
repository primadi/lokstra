package flow

import (
	"errors"
	"testing"
)

func TestDoMethod(t *testing.T) {
	handler := NewHandler(nil, "test-handler")

	// Test Do adds a custom step
	handler.Do(func(ctx *Context) error {
		ctx.Set("test_value", "hello")
		return nil
	})

	if len(handler.steps) != 1 {
		t.Errorf("Expected 1 step after Do, got %d", len(handler.steps))
	}

	// Verify it's a customStep
	if step, ok := handler.steps[0].(*customStep); ok {
		meta := step.Meta()
		if meta.Name != "custom.function" {
			t.Errorf("Expected step name 'custom.function', got '%s'", meta.Name)
		}
		if meta.Kind != StepNormal {
			t.Errorf("Expected step kind StepNormal, got %v", meta.Kind)
		}
	} else {
		t.Error("Expected step to be customStep")
	}
}

func TestDoExecution(t *testing.T) {
	// Create a simple context for testing
	ctx := &Context{
		vars: make(map[string]any),
	}

	// Create custom step
	executed := false
	step := &customStep{
		fn: func(ctx *Context) error {
			executed = true
			ctx.Set("custom_result", "success")
			return nil
		},
	}

	// Execute the step
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Custom function was not executed")
	}

	// Check if variable was set
	if value, exists := ctx.Get("custom_result"); !exists || value != "success" {
		t.Errorf("Expected custom_result to be 'success', got %v (exists: %v)", value, exists)
	}
}

func TestDoErrorHandling(t *testing.T) {
	ctx := &Context{
		vars: make(map[string]any),
	}

	// Create custom step that returns error
	expectedError := errors.New("custom function error")
	step := &customStep{
		fn: func(ctx *Context) error {
			return expectedError
		},
	}

	// Execute the step
	err := step.Run(ctx)
	if err == nil {
		t.Error("Expected error from custom function")
	}

	if err != expectedError {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestDoChaining(t *testing.T) {
	handler := NewHandler(nil, "test-handler")

	// Test fluent API chaining with Do
	result := handler.
		ExecSql("INSERT INTO test VALUES (?)", 1).
		Done().
		Do(func(ctx *Context) error {
			ctx.Set("step_completed", true)
			return nil
		}).
		QueryRowSql("SELECT COUNT(*) FROM test").
		SaveAs("count")

	if result != handler {
		t.Error("Expected chaining to return same handler")
	}

	if len(handler.steps) != 3 {
		t.Errorf("Expected 3 steps in chain, got %d", len(handler.steps))
	}

	// Verify step types
	expectedTypes := []string{"sql.exec", "custom.function", "sql.query_row"}
	for i, expected := range expectedTypes {
		if i >= len(handler.steps) {
			t.Errorf("Missing step %d", i)
			continue
		}

		meta := handler.steps[i].Meta()
		if meta.Name != expected {
			t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
		}
	}
}

func TestDoWithVariableAccess(t *testing.T) {
	ctx := &Context{
		vars: make(map[string]any),
	}

	// Set initial variables
	ctx.Set("input_value", 42)
	ctx.Set("multiplier", 2)

	// Create custom step that processes variables
	step := &customStep{
		fn: func(ctx *Context) error {
			input, exists1 := ctx.Get("input_value")
			multiplier, exists2 := ctx.Get("multiplier")

			if !exists1 || !exists2 {
				return errors.New("required variables not found")
			}

			result := input.(int) * multiplier.(int)
			ctx.Set("calculated_result", result)
			return nil
		},
	}

	// Execute the step
	err := step.Run(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check calculated result
	if value, exists := ctx.Get("calculated_result"); !exists || value != 84 {
		t.Errorf("Expected calculated_result to be 84, got %v (exists: %v)", value, exists)
	}
}
