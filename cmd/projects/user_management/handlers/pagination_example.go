package handlers

import (
	"github.com/primadi/lokstra/core/flow"
	"github.com/primadi/lokstra/core/request"
	"github.com/primadi/lokstra/serviceapi/auth"
)

// Example: Simple pagination handler using the new helper
func CreateSimplePaginationExample() request.HandlerFunc {
	return flow.NewFlow[ListUserRequestDTO]("SimplePaginationExample").
		AddPaginationQueryAction(). // <- Helper yang baru!
		AddAction("get_data", simplePaginationAction).AsHandler()
}

func simplePaginationAction(fctx *flow.Context[ListUserRequestDTO]) error {
	// Pagination sudah otomatis di-bind dan di-validate oleh AddPaginationQueryAction
	_, exists := fctx.GetPagination()
	if !exists {
		return fctx.ErrorInternal("Pagination context not found")
	}

	// Simulasi data (di real app, ini dari database)
	users := []*auth.User{
		{ID: "1", Username: "user1", Email: "user1@example.com"},
		{ID: "2", Username: "user2", Email: "user2@example.com"},
	}
	total := 100 // Total records

	// Return response dengan format yang konsisten
	return fctx.PaginatedOk(users, total)
	// Output format:
	// {
	//   "data": [...],
	//   "pagination": {
	//     "page": 1,
	//     "page_size": 10,
	//     "total": 100,
	//     "total_pages": 10,
	//     "has_next": true,
	//     "has_prev": false
	//   },
	//   "filters": {
	//     "applied": {...}
	//   }
	// }
}

// Comparison: Before vs After

// BEFORE - Manual pagination binding (verbose):
func OldWayExample() request.HandlerFunc {
	return flow.NewFlow[ListUserRequestDTO]("OldWay").
		AddAction("bind_and_validate_pagination", func(fctx *flow.Context[ListUserRequestDTO]) error {
			pagination, err := fctx.BindPaginationQuery()
			if err != nil {
				return fctx.ErrorBadRequest("Invalid pagination parameters: " + err.Error())
			}

			// Manual validation
			if pagination.Page <= 0 {
				pagination.Page = 1
			}
			if pagination.PageSize <= 0 {
				pagination.PageSize = 10
			}
			if pagination.PageSize > 100 {
				pagination.PageSize = 100
			}

			// Manual response building
			users := []*auth.User{} // get data
			total := 100

			totalPages := (total + pagination.PageSize - 1) / pagination.PageSize
			response := map[string]interface{}{
				"data": users,
				"pagination": map[string]interface{}{
					"page":        pagination.Page,
					"page_size":   pagination.PageSize,
					"total":       total,
					"total_pages": totalPages,
					"has_next":    pagination.Page < totalPages,
					"has_prev":    pagination.Page > 1,
				},
				"filters": map[string]interface{}{
					"applied": pagination.Filter,
				},
			}

			return fctx.Ok(response)
		}).AsHandler()
}

// AFTER - Using helper (concise):
func NewWayExample() request.HandlerFunc {
	return flow.NewFlow[ListUserRequestDTO]("NewWay").
		AddPaginationQueryAction(). // <- Otomatis bind + validate
		AddAction("get_data", func(fctx *flow.Context[ListUserRequestDTO]) error {
			users := []*auth.User{} // get data
			total := 100
			return fctx.PaginatedOk(users, total) // <- Otomatis format response
		}).AsHandler()
}
