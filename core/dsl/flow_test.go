package dsl

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi"
)

// Mock implementations for testing
type mockDbPool struct{}

func (m *mockDbPool) Acquire(ctx context.Context, schema string) (serviceapi.DbConn, error) {
	return &mockDbConn{}, nil
}

type mockDbConn struct{}

func (m *mockDbConn) Begin(ctx context.Context) (serviceapi.DbTx, error) {
	return &mockDbTx{}, nil
}

func (m *mockDbConn) Transaction(ctx context.Context, fn func(tx serviceapi.DbExecutor) error) error {
	return fn(m)
}

func (m *mockDbConn) Release() error {
	return nil
}

func (m *mockDbConn) Exec(ctx context.Context, query string, args ...any) (serviceapi.CommandResult, error) {
	return serviceapi.NewCommandResult(func() int64 { return 1 }), nil
}

func (m *mockDbConn) Query(ctx context.Context, query string, args ...any) (serviceapi.Rows, error) {
	return &mockRows{}, nil
}

func (m *mockDbConn) QueryRow(ctx context.Context, query string, args ...any) serviceapi.Row {
	return &mockRow{}
}

func (m *mockDbConn) SelectOne(ctx context.Context, query string, args []any, dest ...any) error {
	return nil
}

func (m *mockDbConn) SelectMustOne(ctx context.Context, query string, args []any, dest ...any) error {
	return nil
}

func (m *mockDbConn) SelectOneRowMap(ctx context.Context, query string, args ...any) (serviceapi.RowMap, error) {
	return serviceapi.RowMap{"id": 1, "name": "test"}, nil
}

func (m *mockDbConn) SelectManyRowMap(ctx context.Context, query string, args ...any) ([]serviceapi.RowMap, error) {
	return []serviceapi.RowMap{
		{"id": 1, "name": "test1"},
		{"id": 2, "name": "test2"},
	}, nil
}

func (m *mockDbConn) SelectManyWithMapper(ctx context.Context, fnScan func(serviceapi.Row) (any, error), query string, args ...any) (any, error) {
	return nil, nil
}

func (m *mockDbConn) IsExists(ctx context.Context, query string, args ...any) (bool, error) {
	return true, nil
}

func (m *mockDbConn) IsErrorNoRows(err error) bool {
	return false
}

type mockDbTx struct {
	mockDbConn
}

func (m *mockDbTx) Commit(ctx context.Context) error {
	return nil
}

func (m *mockDbTx) Rollback(ctx context.Context) error {
	return nil
}

type mockRows struct {
	count int
}

func (m *mockRows) Next() bool {
	m.count++
	return m.count <= 2
}

func (m *mockRows) Scan(dest ...any) error {
	return nil
}

func (m *mockRows) Close() error {
	return nil
}

func (m *mockRows) Err() error {
	return nil
}

type mockRow struct{}

func (m *mockRow) Scan(dest ...any) error {
	return nil
}

type mockLogger struct{}

func (m *mockLogger) Debugf(msg string, v ...any)                              {}
func (m *mockLogger) Infof(msg string, v ...any)                               {}
func (m *mockLogger) Warnf(msg string, v ...any)                               {}
func (m *mockLogger) Errorf(msg string, v ...any)                              {}
func (m *mockLogger) Fatalf(msg string, v ...any)                              {}
func (m *mockLogger) GetLogLevel() serviceapi.LogLevel                         { return serviceapi.LogLevelInfo }
func (m *mockLogger) SetLogLevel(level serviceapi.LogLevel)                    {}
func (m *mockLogger) WithField(key string, value any) serviceapi.Logger        { return m }
func (m *mockLogger) WithFields(fields serviceapi.LogFields) serviceapi.Logger { return m }
func (m *mockLogger) SetFormat(format string)                                  {}
func (m *mockLogger) SetOutput(output string)                                  {}

type mockMetrics struct {
	counters   map[string]int
	histograms map[string][]float64
	gauges     map[string]float64
}

func newMockMetrics() *mockMetrics {
	return &mockMetrics{
		counters:   make(map[string]int),
		histograms: make(map[string][]float64),
		gauges:     make(map[string]float64),
	}
}

func (m *mockMetrics) IncCounter(name string, labels serviceapi.Labels) {
	key := name
	for k, v := range labels {
		key += "_" + k + ":" + v
	}
	m.counters[key]++
}

func (m *mockMetrics) ObserveHistogram(name string, value float64, labels serviceapi.Labels) {
	key := name
	for k, v := range labels {
		key += "_" + k + ":" + v
	}
	m.histograms[key] = append(m.histograms[key], value)
}

func (m *mockMetrics) SetGauge(name string, value float64, labels serviceapi.Labels) {
	key := name
	for k, v := range labels {
		key += "_" + k + ":" + v
	}
	m.gauges[key] = value
}

type mockI18n struct{}

func (m *mockI18n) T(lang, code string, params map[string]any) string {
	switch code {
	case "validation.failed":
		return fmt.Sprintf("Validation failed for field %s", params["field"])
	case "database.operation_failed":
		return fmt.Sprintf("Database operation %s failed: %s", params["operation"], params["error"])
	case "validation.required_field":
		return fmt.Sprintf("Field %s is required", params["field"])
	default:
		return code
	}
}

func TestFlowBasicExecution(t *testing.T) {
	metrics := newMockMetrics()
	sv := NewServiceVar(
		&mockDbPool{},
		"test_schema",
		&mockLogger{},
		metrics,
		&mockI18n{},
		&struct{}{},
		make(map[string]any),
	)

	flow := NewFlow("test-flow", sv)
	flow.BeginTx().
		ExecSql("INSERT INTO test (name) VALUES (?)", "test").
		QuerySaveAs("SELECT * FROM test", "results").
		CommitOrRollback()

	reqCtx := &request.Context{
		Context: context.Background(),
	}

	err := flow.Run(reqCtx)
	if err != nil {
		t.Errorf("Flow execution failed: %v", err)
	}

	// Check if variable was saved
	if results, ok := sv.Vars["results"]; !ok {
		t.Error("Results variable not saved")
	} else {
		resultSlice, ok := results.([]serviceapi.RowMap)
		if !ok {
			t.Error("Results not of expected type")
		} else if len(resultSlice) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resultSlice))
		}
	}

	// Verify metrics were collected
	if metrics.counters["dsl_flow_started_flow_name:test-flow"] != 1 {
		t.Error("Flow start metric not recorded")
	}

	if metrics.counters["dsl_flow_succeeded_flow_name:test-flow"] != 1 {
		t.Error("Flow success metric not recorded")
	}
}

func TestFlowConditional(t *testing.T) {
	sv := NewServiceVar(
		&mockDbPool{},
		"test_schema",
		&mockLogger{},
		nil,
		nil,
		&struct{}{},
		make(map[string]any),
	)

	flow := NewFlow("conditional-flow", sv)
	flow.Do(func(ctx *FlowContext[struct{}]) error {
		ctx.SetVar("condition", true)
		return nil
	}).
		If(func(ctx *FlowContext[struct{}]) bool {
			val, _ := ctx.GetVar("condition")
			return val == true
		}, newStepCustom(func(ctx *FlowContext[struct{}]) error {
			ctx.SetVar("executed", true)
			return nil
		}))

	reqCtx := &request.Context{
		Context: context.Background(),
	}

	err := flow.Run(reqCtx)
	if err != nil {
		t.Errorf("Flow execution failed: %v", err)
	}

	// Check if conditional step was executed
	if executed, ok := sv.Vars["executed"]; !ok || executed != true {
		t.Error("Conditional step was not executed")
	}
}

func TestFlowErrorHandling(t *testing.T) {
	sv := NewServiceVar(
		&mockDbPool{},
		"test_schema",
		&mockLogger{},
		nil,
		nil,
		&struct{}{},
		make(map[string]any),
	)

	flow := NewFlow("error-flow", sv)
	flow.Do(func(ctx *FlowContext[struct{}]) error {
		return errors.New("test error")
	})

	reqCtx := &request.Context{
		Context: context.Background(),
	}

	err := flow.Run(reqCtx)
	if err == nil {
		t.Error("Expected error but got none")
	}
	if err.Error() != "test error" {
		t.Errorf("Expected 'test error', got '%v'", err)
	}
}

func TestFlowRetry(t *testing.T) {
	sv := NewServiceVar(
		&mockDbPool{},
		"test_schema",
		&mockLogger{},
		nil,
		nil,
		&struct{}{},
		make(map[string]any),
	)

	attempts := 0
	retryStep := newStepCustom(func(ctx *FlowContext[struct{}]) error {
		attempts++
		if attempts < 3 {
			return errors.New("retry me")
		}
		ctx.SetVar("success", true)
		return nil
	})

	flow := NewFlow("retry-flow", sv)
	flow.Retry(retryStep, 3)

	reqCtx := &request.Context{
		Context: context.Background(),
	}

	err := flow.Run(reqCtx)
	if err != nil {
		t.Errorf("Flow execution failed: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if success, ok := sv.Vars["success"]; !ok || success != true {
		t.Error("Retry step did not succeed")
	}
}

func TestFlowI18nAndValidation(t *testing.T) {
	sv := NewServiceVar(
		&mockDbPool{},
		"test_schema",
		&mockLogger{},
		newMockMetrics(),
		&mockI18n{},
		&struct{}{},
		make(map[string]any),
	)

	flow := NewFlow("validation-flow", sv)
	flow.Validate(func(ctx *FlowContext[struct{}]) error {
		// Test localized error
		return ErrRequiredField("username")
	})

	reqCtx := &request.Context{
		Context: context.Background(),
		Request: &http.Request{
			Header: make(http.Header),
		},
	}

	err := flow.Run(reqCtx)
	if err == nil {
		t.Error("Expected validation error but got none")
	}

	// Test that the error can be localized
	flowCtx := NewFlowContext(reqCtx, sv)
	localizedMsg := LocalizeErrorString(flowCtx, err)
	expectedMsg := "Field username is required"
	if localizedMsg != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, localizedMsg)
	}
}

func TestFlowMetricsCollection(t *testing.T) {
	metrics := newMockMetrics()
	sv := NewServiceVar(
		&mockDbPool{},
		"test_schema",
		&mockLogger{},
		metrics,
		&mockI18n{},
		&struct{}{},
		make(map[string]any),
	)

	flow := NewFlow("metrics-test", sv)
	flow.QuerySaveAs("SELECT COUNT(*) FROM users", "count")

	reqCtx := &request.Context{
		Context: context.Background(),
	}

	err := flow.Run(reqCtx)
	if err != nil {
		t.Errorf("Flow execution failed: %v", err)
	}

	// Verify specific metrics
	expectedMetrics := []string{
		"dsl_flow_started_flow_name:metrics-test",
		"dsl_flow_succeeded_flow_name:metrics-test",
		"dsl_db_operation_started_operation:select_many",
		"dsl_db_operation_succeeded_operation:select_many",
	}

	for _, metric := range expectedMetrics {
		if count := metrics.counters[metric]; count != 1 {
			t.Errorf("Metric %s expected count 1, got %d", metric, count)
		}
	}

	// Verify histograms
	if len(metrics.histograms) == 0 {
		t.Error("No histogram metrics collected")
	}
}
