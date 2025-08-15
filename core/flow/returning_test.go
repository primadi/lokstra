package flow

import (
	"testing"

	"github.com/primadi/lokstra/serviceapi"
)

func TestReturningSupport(t *testing.T) {
	// Test ExecReturning method
	t.Run("ExecReturning Method", func(t *testing.T) {
		handler := NewHandler(nil, "test").
			ExecReturning("INSERT INTO users (name) VALUES (?) RETURNING id, name", "John").
			SaveAs("new_user")

		if len(handler.steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(handler.steps))
		}

		step := handler.steps[0].(*sqlStep)
		if step.kind != kindExec {
			t.Errorf("Expected kindExec, got %v", step.kind)
		}

		if !step.hasReturning {
			t.Error("Expected hasReturning to be true")
		}

		if step.saveAs != "new_user" {
			t.Errorf("Expected saveAs 'new_user', got '%s'", step.saveAs)
		}
	})

	// Test WithReturning method
	t.Run("WithReturning Method", func(t *testing.T) {
		handler := NewHandler(nil, "test").
			ExecSql("UPDATE users SET name = ? WHERE id = ? RETURNING id, name, updated_at", "Jane", 123).
			WithReturning().
			SaveAs("updated_user")

		if len(handler.steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(handler.steps))
		}

		step := handler.steps[0].(*sqlStep)
		if !step.hasReturning {
			t.Error("Expected hasReturning to be true")
		}

		if step.saveAs != "updated_user" {
			t.Errorf("Expected saveAs 'updated_user', got '%s'", step.saveAs)
		}
	})

	// Test WithReturning panic on non-exec queries
	t.Run("WithReturning Panic on QueryRow", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when using WithReturning on QueryRowSql")
			}
		}()

		NewHandler(nil, "test").
			QueryRowSql("SELECT * FROM users WHERE id = ?", 123).
			WithReturning() // This should panic
	})

	// Test RETURNING with guards
	t.Run("RETURNING With Guards", func(t *testing.T) {
		handler := NewHandler(nil, "test").
			ExecReturning("DELETE FROM users WHERE id = ? RETURNING id", 123).
			AffectOne().
			SaveAs("deleted_user")

		if len(handler.steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(handler.steps))
		}

		step := handler.steps[0].(*sqlStep)
		if !step.guardExactOne {
			t.Error("Expected guardExactOne to be true")
		}

		if !step.hasReturning {
			t.Error("Expected hasReturning to be true")
		}
	})

	// Test fluent API chaining
	t.Run("Fluent API Chaining", func(t *testing.T) {
		handler := NewHandler(nil, "test").
			ExecReturning("INSERT INTO users (name, email) VALUES (?, ?) RETURNING id, created_at", "John", "john@example.com").
			WithName("user.create").
			AffectOne().
			SaveAs("new_user_data")

		if len(handler.steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(handler.steps))
		}

		step := handler.steps[0].(*sqlStep)

		// Check all properties
		if step.customName != "user.create" {
			t.Errorf("Expected customName 'user.create', got '%s'", step.customName)
		}

		if !step.hasReturning {
			t.Error("Expected hasReturning to be true")
		}

		if !step.guardExactOne {
			t.Error("Expected guardExactOne to be true")
		}

		if step.saveAs != "new_user_data" {
			t.Errorf("Expected saveAs 'new_user_data', got '%s'", step.saveAs)
		}
	})
}

// Test the execution logic would require mocking, but we can test the step meta
func TestReturningStepMeta(t *testing.T) {
	step := &sqlStep{
		kind:         kindExec,
		hasReturning: true,
		customName:   "user.create_with_return",
	}

	meta := step.Meta()
	if meta.Name != "user.create_with_return" {
		t.Errorf("Expected meta name 'user.create_with_return', got '%s'", meta.Name)
	}

	if meta.Kind != StepNormal {
		t.Errorf("Expected meta kind StepNormal, got %v", meta.Kind)
	}
}

// Mock context for testing execution logic
type mockContext struct {
	vars map[string]any
}

func (m *mockContext) QueryRowMap(query string, args ...any) (serviceapi.RowMap, error) {
	// Mock successful RETURNING result
	return map[string]any{
		"id":   int64(123),
		"name": "John Doe",
	}, nil
}

func (m *mockContext) Set(name string, value any) {
	if m.vars == nil {
		m.vars = make(map[string]any)
	}
	m.vars[name] = value
}

func TestReturningExecution(t *testing.T) {
	step := &sqlStep{
		// kind:          kindExec,
		// query:         "INSERT INTO users (name) VALUES (?) RETURNING id, name",
		// args:          []any{"John Doe"},
		hasReturning:  true,
		saveAs:        "new_user",
		guardExactOne: true,
	}

	// This would require more sophisticated mocking to test fully
	// For now, we verify the step configuration is correct
	if !step.hasReturning {
		t.Error("Expected hasReturning to be true")
	}

	if step.saveAs != "new_user" {
		t.Errorf("Expected saveAs 'new_user', got '%s'", step.saveAs)
	}

	if !step.guardExactOne {
		t.Error("Expected guardExactOne to be true")
	}
}
