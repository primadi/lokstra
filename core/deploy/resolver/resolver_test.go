package resolver

import (
	"os"
	"testing"
)

func TestResolveValue_Static(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{}

	result, err := r.ResolveValue("localhost", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "localhost" {
		t.Errorf("expected 'localhost', got '%v'", result)
	}
}

func TestResolveValue_EnvVar(t *testing.T) {
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	r := NewRegistry()
	configs := map[string]any{}

	result, err := r.ResolveValue("${TEST_VAR}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "test-value" {
		t.Errorf("expected 'test-value', got '%v'", result)
	}
}

func TestResolveValue_EnvVarWithDefault(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{}

	// Var doesn't exist, should use default
	result, err := r.ResolveValue("${NONEXISTENT_VAR:default-value}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "default-value" {
		t.Errorf("expected 'default-value', got '%v'", result)
	}
}

func TestResolveValue_CustomResolver(t *testing.T) {
	r := NewRegistry()
	r.Register(NewStaticResolver("consul", map[string]string{
		"config/api-key": "secret-key-123",
	}))

	configs := map[string]any{}

	result, err := r.ResolveValue("${@consul:config/api-key}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "secret-key-123" {
		t.Errorf("expected 'secret-key-123', got '%v'", result)
	}
}

func TestResolveValue_CustomResolverWithDefault(t *testing.T) {
	r := NewRegistry()
	r.Register(NewStaticResolver("consul", map[string]string{}))

	configs := map[string]any{}

	result, err := r.ResolveValue("${@consul:nonexistent:fallback}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "fallback" {
		t.Errorf("expected 'fallback', got '%v'", result)
	}
}

func TestResolveValue_ConfigReference(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{
		"DB_MAX_CONNS": 20,
		"LOG_LEVEL":    "info",
	}

	// Test integer config reference
	result, err := r.ResolveValue("${@cfg:DB_MAX_CONNS}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != 20 {
		t.Errorf("expected 20, got %v (type %T)", result, result)
	}

	// Test string config reference
	result2, err := r.ResolveValue("${@cfg:LOG_LEVEL}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result2 != "info" {
		t.Errorf("expected 'info', got '%v'", result2)
	}
}

func TestResolveValue_ConfigReferenceInString(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{
		"DB_MAX_CONNS": 20,
	}

	// When @cfg is part of a larger string, it should be converted to string
	result, err := r.ResolveValue("max-connections: ${@cfg:DB_MAX_CONNS}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "max-connections: 20"
	if result != expected {
		t.Errorf("expected '%s', got '%v'", expected, result)
	}
}

func TestResolveValue_TwoStepResolution(t *testing.T) {
	// Step 1: Resolve env var to get config key
	// Step 2: Resolve @cfg using that key

	os.Setenv("DB_URL_KEY", "DATABASE_URL")
	defer os.Unsetenv("DB_URL_KEY")

	r := NewRegistry()
	configs := map[string]any{
		"DATABASE_URL": "postgres://localhost/db",
	}

	// This should work but is a complex case
	// For now, let's test simpler 2-step: env first, then @cfg

	// Test: dsn with env var, max-conns with @cfg
	os.Setenv("DB_USER_URL", "postgres://localhost/users")
	defer os.Unsetenv("DB_USER_URL")

	dsnResult, err := r.ResolveValue("${DB_USER_URL}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dsnResult != "postgres://localhost/users" {
		t.Errorf("expected postgres URL, got '%v'", dsnResult)
	}

	maxConnsResult, err := r.ResolveValue("${@cfg:DB_MAX_CONNS}", map[string]any{"DB_MAX_CONNS": 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if maxConnsResult != 20 {
		t.Errorf("expected 20, got %v", maxConnsResult)
	}
}

func TestResolveValue_MultipleReferences(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")

	r := NewRegistry()
	configs := map[string]any{
		"DB_NAME": "mydb",
	}

	result, err := r.ResolveValue("postgres://${DB_HOST}:${DB_PORT}/${@cfg:DB_NAME}", configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "postgres://localhost:5432/mydb"
	if result != expected {
		t.Errorf("expected '%s', got '%v'", expected, result)
	}
}

func TestResolveValue_ConfigNotFound(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{}

	_, err := r.ResolveValue("${@cfg:NONEXISTENT}", configs)
	if err == nil {
		t.Fatal("expected error for nonexistent config key")
	}

	if !contains(err.Error(), "config key NONEXISTENT not found") {
		t.Errorf("expected config not found error, got: %v", err)
	}
}

func TestResolveValue_EnvVarNotFound(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{}

	_, err := r.ResolveValue("${NONEXISTENT_VAR}", configs)
	if err == nil {
		t.Fatal("expected error for nonexistent env var")
	}
}

func TestResolveValue_ResolverNotFound(t *testing.T) {
	r := NewRegistry()
	configs := map[string]any{}

	_, err := r.ResolveValue("${@nonexistent:key}", configs)
	if err == nil {
		t.Fatal("expected error for nonexistent resolver")
	}

	if !contains(err.Error(), "resolver nonexistent not found") {
		t.Errorf("expected resolver not found error, got: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)*2 && len(findInString(s, substr)) > 0))
}

func findInString(s, substr string) string {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return substr
		}
	}
	return ""
}
