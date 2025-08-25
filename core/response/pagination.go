package response

// PaginatedResponse provides a standardized format for paginated API responses
type PaginatedResponse[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
	Filters    map[string]any `json:"filters,omitempty"`
}

// PaginationInfo contains comprehensive pagination metadata
type PaginationInfo struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"page_size"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewPaginatedResponse creates a new paginated response with calculated pagination info
func NewPaginatedResponse[T any](data []T, page, pageSize, total int, filters map[string]any) *PaginatedResponse[T] {
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	hasNext := page < totalPages
	hasPrev := page > 1

	return &PaginatedResponse[T]{
		Data: data,
		Pagination: PaginationInfo{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
		Filters: filters,
	}
}

// EmptyPaginatedResponse creates an empty paginated response
func EmptyPaginatedResponse[T any](page, pageSize int, filters map[string]any) *PaginatedResponse[T] {
	return NewPaginatedResponse([]T{}, page, pageSize, 0, filters)
}
