package api_formatter

// ApiResponse standardizes API response structure
type ApiResponse struct {
	Status    string `json:"status"`               // "success" | "error"
	Message   string `json:"message,omitempty"`    // Human readable message
	Data      any    `json:"data,omitempty"`       // Response data
	Error     *Error `json:"error,omitempty"`      // Error details if status = "error"
	Meta      *Meta  `json:"meta,omitempty"`       // Metadata for lists/pagination
	RequestID string `json:"request_id,omitempty"` // For tracing
}

// Error represents detailed error information
type Error struct {
	Code    string         `json:"code"`              // Error code (e.g. "VALIDATION_ERROR")
	Message string         `json:"message"`           // Error message
	Details map[string]any `json:"details,omitempty"` // Additional error details
	Fields  []FieldError   `json:"fields,omitempty"`  // Validation field errors
}

// FieldError represents validation error for specific field
type FieldError struct {
	Field   string `json:"field"`           // Field name
	Code    string `json:"code"`            // Error code (e.g. "REQUIRED")
	Message string `json:"message"`         // Error message
	Value   any    `json:"value,omitempty"` // Invalid value provided
}

// Meta contains pagination and other metadata
type Meta struct {
	*ListMeta     `json:",omitempty"`
	*RequestMeta  `json:",omitempty"`
	*ResponseMeta `json:",omitempty"`
}

// ListMeta contains pagination information
type ListMeta struct {
	Page       int  `json:"page"`        // Current page
	PageSize   int  `json:"page_size"`   // Items per page
	Total      int  `json:"total"`       // Total items
	TotalPages int  `json:"total_pages"` // Total pages
	HasNext    bool `json:"has_next"`    // Has next page
	HasPrev    bool `json:"has_prev"`    // Has previous page
}

// RequestMeta contains request-related metadata
type RequestMeta struct {
	Filters  map[string]string `json:"filters,omitempty"`   // Applied filters
	OrderBy  []string          `json:"order_by,omitempty"`  // Applied ordering
	Fields   []string          `json:"fields,omitempty"`    // Selected fields
	Search   string            `json:"search,omitempty"`    // Search query
	DataType string            `json:"data_type,omitempty"` // Response format type
}

// ResponseMeta contains response-related metadata
type ResponseMeta struct {
	ProcessingTime string            `json:"processing_time,omitempty"` // e.g. "15ms"
	CacheStatus    string            `json:"cache_status,omitempty"`    // "hit" | "miss" | "bypass"
	Headers        map[string]string `json:"headers,omitempty"`         // Additional headers set
}

// CalculateListMeta calculates pagination metadata
func CalculateListMeta(page, pageSize, total int) *ListMeta {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	return &ListMeta{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
