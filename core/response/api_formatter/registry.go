package api_formatter

import "net/http"

// ResponseFormatter defines interface for different response formatting strategies
type ResponseFormatter interface {
	// Formats successful response with data and optional message
	Success(data any, message ...string) any

	// Formats resource creation response (HTTP 201) with data and optional message
	Created(data any, message ...string) any

	// Formats error response with code, message, and optional details
	Error(code string, message string, details ...map[string]any) any

	// Formats validation error response
	ValidationError(message string, fields []FieldError) any

	// Formats not found error response
	NotFound(message string) any

	// Formats paginated list response
	List(data any, meta *ListMeta) any

	// Parses HTTP response into ClientResponse according to formatter's expected format
	ParseClientResponse(resp *http.Response, cr *ClientResponse) error
}

// Registry for response formatters
var formatterRegistry = make(map[string]func() ResponseFormatter)

// RegisterFormatter registers a new ResponseFormatter with a name
func RegisterFormatter(name string, constructor func() ResponseFormatter) {
	formatterRegistry[name] = constructor
}

// CreateFormatter creates a ResponseFormatter based on the formatter type
func CreateFormatter(formatterType string) ResponseFormatter {
	if constructor, exists := formatterRegistry[formatterType]; exists {
		return constructor()
	}
	// Default to structured API formatter
	return NewApiResponseFormatter()
}

// Global response formatter - can be replaced at startup
var globalFormatter ResponseFormatter = NewApiResponseFormatter()

// SetGlobalFormatter sets the global response formatter
func SetGlobalFormatter(formatter ResponseFormatter) {
	globalFormatter = formatter
}

// SetGlobalFormatterByName sets the global response formatter by registered name
func SetGlobalFormatterByName(name string) {
	globalFormatter = CreateFormatter(name)
}

// GetGlobalFormatter returns the current global response formatter
func GetGlobalFormatter() ResponseFormatter {
	return globalFormatter
}

func init() {
	// Register built-in formatters
	RegisterFormatter("default", NewApiResponseFormatter)
	RegisterFormatter("simple", NewSimpleResponseFormatter)
}
