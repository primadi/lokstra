package flow

import (
	"github.com/primadi/lokstra/core/request"
)

// AddPaginationQueryAction adds an action that automatically binds pagination query and stores it in context
func (f *Flow[T]) AddPaginationQueryAction() *Flow[T] {
	return f.AddAction("bind_pagination_query", func(fctx *Context[T]) error {
		pagination, err := fctx.BindPaginationQuery()
		if err != nil {
			return fctx.ErrorBadRequest("Invalid pagination parameters: " + err.Error())
		}

		// Validate pagination limits
		if pagination.Page <= 0 {
			pagination.Page = 1
		}
		if pagination.PageSize <= 0 {
			pagination.PageSize = 10
		}
		if pagination.PageSize > 100 {
			pagination.PageSize = 100
		}

		// Store pagination in context for later actions
		fctx.Set("pagination", pagination)
		return nil
	})
}

// GetPagination retrieves pagination from context set by AddPaginationQueryAction
func (fctx *Context[T]) GetPagination() (*request.PaginationQuery, bool) {
	if pagination, exists := fctx.Get("pagination"); exists {
		if p, ok := pagination.(*request.PaginationQuery); ok {
			return p, true
		}
	}
	return nil, false
}

// PaginatedOk returns a standardized paginated response using stored pagination context
func (fctx *Context[T]) PaginatedOk(data any, total int) error {
	pagination, exists := fctx.GetPagination()
	if !exists {
		return fctx.ErrorInternal("Pagination context not found. Did you call AddPaginationQueryAction()?")
	}

	// Create filters map
	filters := make(map[string]any)
	if pagination.Filter != nil {
		for k, v := range pagination.Filter {
			filters[k] = v
		}
	}

	// Calculate pagination info
	totalPages := (total + pagination.PageSize - 1) / pagination.PageSize
	if totalPages < 1 {
		totalPages = 1
	}

	hasNext := pagination.Page < totalPages
	hasPrev := pagination.Page > 1

	// Create standardized response
	response := map[string]any{
		"data": data,
		"pagination": map[string]any{
			"page":        pagination.Page,
			"page_size":   pagination.PageSize,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
		"filters": map[string]any{
			"applied": filters,
		},
	}

	return fctx.Ok(response)
}
