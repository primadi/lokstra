package flow

import (
	"testing"
)

func TestStepNaming(t *testing.T) {
	// Test default SQL step names
	t.Run("Default SQL Step Names", func(t *testing.T) {
		h := NewHandler(nil, "test").
			ExecSql("INSERT INTO test VALUES (?)", 1).
			Done().
			QueryRowSql("SELECT COUNT(*) FROM test").
			SaveAs("count").
			QuerySql("SELECT * FROM test").
			SaveAs("all_rows")

		expectedNames := []string{"sql.exec", "sql.query_row", "sql.query"}
		for i, expected := range expectedNames {
			if i >= len(h.steps) {
				t.Errorf("Missing step %d", i)
				continue
			}

			meta := h.steps[i].Meta()
			if meta.Name != expected {
				t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
			}
		}
	})

	// Test custom SQL step names
	t.Run("Custom SQL Step Names", func(t *testing.T) {
		h := NewHandler(nil, "test").
			ExecSql("INSERT INTO users (name) VALUES (?)", "John").
			WithName("user.create").
			Done().
			QueryRowSql("SELECT * FROM users WHERE name = ?", "John").
			WithName("user.find_by_name").
			SaveAs("user_data")

		expectedNames := []string{"user.create", "user.find_by_name"}
		for i, expected := range expectedNames {
			if i >= len(h.steps) {
				t.Errorf("Missing step %d", i)
				continue
			}

			meta := h.steps[i].Meta()
			if meta.Name != expected {
				t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
			}
		}
	})

	// Test default custom step names
	t.Run("Default Custom Step Names", func(t *testing.T) {
		h := NewHandler(nil, "test").
			Do(func(ctx *Context) error {
				ctx.Set("test", "value")
				return nil
			})

		if len(h.steps) != 1 {
			t.Fatalf("Expected 1 step, got %d", len(h.steps))
		}

		meta := h.steps[0].Meta()
		if meta.Name != "custom.function" {
			t.Errorf("Expected default name 'custom.function', got '%s'", meta.Name)
		}
	})

	// Test named custom steps
	t.Run("Named Custom Steps", func(t *testing.T) {
		h := NewHandler(nil, "test").
			DoNamed("user.validate", func(ctx *Context) error {
				// Validation logic here
				return nil
			}).
			DoNamed("payment.process", func(ctx *Context) error {
				// Payment processing logic
				return nil
			})

		expectedNames := []string{"user.validate", "payment.process"}
		for i, expected := range expectedNames {
			if i >= len(h.steps) {
				t.Errorf("Missing step %d", i)
				continue
			}

			meta := h.steps[i].Meta()
			if meta.Name != expected {
				t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
			}
		}
	})

	// Test transaction step names (should remain unchanged)
	t.Run("Transaction Step Names", func(t *testing.T) {
		h := NewHandler(nil, "test").
			BeginTx().
			CommitOrRollback()

		expectedNames := []string{"tx.begin", "tx.end"}
		for i, expected := range expectedNames {
			if i >= len(h.steps) {
				t.Errorf("Missing step %d", i)
				continue
			}

			meta := h.steps[i].Meta()
			if meta.Name != expected {
				t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
			}
		}
	})
}

func TestStepNamingChaining(t *testing.T) {
	// Test that WithName doesn't break fluent API
	handler := NewHandler(nil, "test").
		ExecSql("INSERT INTO test VALUES (?)", 1).
		WithName("test.insert").
		AffectOne().
		Done().
		QueryRowSql("SELECT COUNT(*) FROM test").
		WithName("test.count").
		SaveAs("count")

	if len(handler.steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(handler.steps))
	}

	// Verify step names
	expectedNames := []string{"test.insert", "test.count"}
	for i, expected := range expectedNames {
		meta := handler.steps[i].Meta()
		if meta.Name != expected {
			t.Errorf("Step %d: expected name '%s', got '%s'", i, expected, meta.Name)
		}
	}
}

func TestEmptyCustomName(t *testing.T) {
	// Test that empty custom name falls back to default
	step := &customStep{
		fn:   func(ctx *Context) error { return nil },
		name: "", // empty name
	}

	meta := step.Meta()
	if meta.Name != "custom.function" {
		t.Errorf("Expected fallback to 'custom.function', got '%s'", meta.Name)
	}

	// Test with actual empty name
	step2 := &customStep{
		fn: func(ctx *Context) error { return nil },
		// name field not set
	}

	meta2 := step2.Meta()
	if meta2.Name != "custom.function" {
		t.Errorf("Expected fallback to 'custom.function', got '%s'", meta2.Name)
	}
}
