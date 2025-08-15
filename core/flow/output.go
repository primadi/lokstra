package flow

import "github.com/primadi/lokstra/serviceapi"

// FlowOutput stores the results of flow execution
type FlowOutput struct {
	// Variables set by steps using SaveAs()
	Variables map[string]any

	// Execution metadata
	StepsExecuted int
	LastStepMeta  *StepMeta

	// Error information (if any)
	Error error
}

// NewFlowOutput creates new FlowOutput
func NewFlowOutput() *FlowOutput {
	return &FlowOutput{
		Variables: make(map[string]any),
	}
}

// GetVariable retrieves a variable from output
func (fo *FlowOutput) GetVariable(key string) (any, bool) {
	value, exists := fo.Variables[key]
	return value, exists
}

// GetString retrieves a string variable
func (fo *FlowOutput) GetString(key string) string {
	if value, exists := fo.Variables[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt64 retrieves an int64 variable
func (fo *FlowOutput) GetInt64(key string) int64 {
	if value, exists := fo.Variables[key]; exists {
		if i, ok := value.(int64); ok {
			return i
		}
	}
	return 0
}

// GetRowMap retrieves a RowMap variable (from SaveAs on QueryRow)
func (fo *FlowOutput) GetRowMap(key string) serviceapi.RowMap {
	if value, exists := fo.Variables[key]; exists {
		if rowMap, ok := value.(serviceapi.RowMap); ok {
			return rowMap
		}
	}
	return nil
}

// GetRowMaps retrieves multiple RowMaps (from SaveAs on Query)
func (fo *FlowOutput) GetRowMaps(key string) []serviceapi.RowMap {
	if value, exists := fo.Variables[key]; exists {
		if rowMaps, ok := value.([]serviceapi.RowMap); ok {
			return rowMaps
		}
	}
	return nil
}
