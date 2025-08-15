package flow

import (
	"fmt"

	"github.com/primadi/lokstra/serviceapi"
)

type sqlKind int

const (
	kindExec sqlKind = iota
	kindQueryRow
	kindQuery
)

type sqlStep struct {
	kind   sqlKind
	query  string
	args   []any
	argsFn func(*Context) []any

	// snapshot
	saveAs  string // snapshot variable name in Ctx
	scanTo  func(serviceapi.Row) error
	collect func(serviceapi.Rows) error
	forEach func(serviceapi.Row) error // lightweight row-by-row mapping

	// guards
	guardExactOne   bool
	guardAtLeast    *int64
	guardAtMost     *int64
	ensureExists    error
	ensureNotExists error

	// custom name for debugging/telemetry
	customName string

	// RETURNING support for EXEC queries
	hasReturning bool // indicates if this EXEC query has RETURNING clause
} // Meta implements Step.
func (s *sqlStep) Meta() StepMeta {
	// Use custom name if provided
	if s.customName != "" {
		return StepMeta{
			Name: s.customName,
			Kind: StepNormal,
		}
	}

	// Otherwise use default naming convention
	var kind string
	switch s.kind {
	case kindExec:
		kind = "sql.exec"
	case kindQueryRow:
		kind = "sql.query_row"
	case kindQuery:
		kind = "sql.query"
	default:
		panic(fmt.Sprintf("unknown sql step kind: %d", s.kind))
	}

	return StepMeta{
		Name: kind,
		Kind: StepNormal,
	}
}

func (s *sqlStep) effectiveArgs(c *Context) []any {
	if s.argsFn != nil {
		return s.argsFn(c)
	}
	return s.args
}

// Run implements Step.
func (s *sqlStep) Run(c *Context) error {
	switch s.kind {
	case kindExec:
		args := s.effectiveArgs(c)

		// Handle RETURNING clause differently
		if s.hasReturning {
			if s.saveAs == "" {
				return fmt.Errorf("RETURNING query requires SaveAs() to capture returned data")
			}

			// For RETURNING, we use QueryRowMap to get the returned data
			result, err := c.QueryRowMap(s.query, args...)
			if err != nil {
				// Check if it's "no rows" error
				conn, connErr := c.CurrentExecutor()
				if connErr == nil && conn.IsErrorNoRows(err) {
					// No rows returned - this might be expected for some operations
					c.Set(s.saveAs, make(map[string]any))
				} else {
					return fmt.Errorf("failed to execute RETURNING query: %w", err)
				}
			} else {
				c.Set(s.saveAs, result)
			}

			// For RETURNING queries, we can't easily get RowsAffected,
			// but we assume one row was affected if we got data back
			rowsAffected := int64(0)
			if len(result) > 0 {
				rowsAffected = 1
			}

			// Apply guards based on assumed rows affected
			if s.guardExactOne && rowsAffected != 1 {
				return fmt.Errorf("expected exactly 1 row affected, got %d", rowsAffected)
			}

			if s.guardAtLeast != nil && rowsAffected < *s.guardAtLeast {
				return fmt.Errorf("expected at least %d rows affected, got %d", *s.guardAtLeast, rowsAffected)
			}

			if s.guardAtMost != nil && rowsAffected > *s.guardAtMost {
				return fmt.Errorf("expected at most %d rows affected, got %d", *s.guardAtMost, rowsAffected)
			}

		} else {
			// Regular EXEC without RETURNING
			res, err := c.Exec(s.query, args...)
			if err != nil {
				return err
			}

			// guards
			if s.guardExactOne || s.guardAtLeast != nil || s.guardAtMost != nil {
				ra := res.RowsAffected()
				if s.guardExactOne && ra != 1 {
					return fmt.Errorf("expected exactly 1 row affected, got %d", ra)
				}

				if s.guardAtLeast != nil && s.guardAtMost != nil {
					if ra < *s.guardAtLeast || ra > *s.guardAtMost {
						return fmt.Errorf("expected rows affected between [%d,%d], got %d",
							*s.guardAtLeast, *s.guardAtMost, ra)
					}
				} else {
					if s.guardAtLeast != nil && ra < *s.guardAtLeast {
						return fmt.Errorf("expected at least %d rows affected, got %d", *s.guardAtLeast, ra)
					}
					if s.guardAtMost != nil && ra > *s.guardAtMost {
						return fmt.Errorf("expected at most %d rows affected, got %d", *s.guardAtMost, ra)
					}
				}
			}

			if s.saveAs != "" {
				c.Set(s.saveAs, serviceapi.NewCommandResult(res.RowsAffected))
			}
		}
	case kindQueryRow:
		args := s.effectiveArgs(c)
		var rowExists bool

		if s.scanTo != nil {
			row, err := c.QueryRow(s.query, args...)
			if err != nil {
				return fmt.Errorf("failed to execute query row: %w", err)
			}

			// Execute scanTo
			if err = s.scanTo(row); err != nil {
				// Check if error is "no rows" using serviceapi method
				conn, connErr := c.CurrentExecutor()
				if connErr != nil {
					return fmt.Errorf("failed to get executor: %w", connErr)
				}
				if conn.IsErrorNoRows(err) {
					rowExists = false
				} else {
					return fmt.Errorf("failed to scan row: %w", err)
				}
			} else {
				rowExists = true
			}
		} else if s.forEach != nil {
			rowCount, err := c.ForEach(s.forEach, s.query, args...)
			if err != nil {
				return fmt.Errorf("failed to process rows: %w", err)
			}
			rowExists = rowCount > 0
		} else if s.saveAs != "" {
			result, err := c.QueryRowMap(s.query, args...)
			if err != nil {
				// Check if it's "no rows" error using serviceapi method
				conn, connErr := c.CurrentExecutor()
				if connErr != nil {
					return fmt.Errorf("failed to get executor: %w", connErr)
				}
				if conn.IsErrorNoRows(err) {
					rowExists = false
					// For saveAs, we still want to set empty result
					c.Set(s.saveAs, make(map[string]any))
				} else {
					return fmt.Errorf("failed to execute query row for saveAs: %w", err)
				}
			} else {
				// Save the result
				c.Set(s.saveAs, result)
				rowExists = len(result) > 0
			}
		} else {
			// No operation specified, just check existence
			row, err := c.QueryRow(s.query, args...)
			if err != nil {
				rowExists = false
			} else {
				var dummy any
				err := row.Scan(&dummy)
				if err != nil {
					// Check if it's a "no rows" error using serviceapi method
					conn, connErr := c.CurrentExecutor()
					if connErr == nil && conn.IsErrorNoRows(err) {
						rowExists = false
					} else {
						rowExists = false // Any scan error means no valid row
					}
				} else {
					rowExists = true
				}
			}
		}

		// Apply the guards
		if s.ensureExists != nil && !rowExists {
			return s.ensureExists
		}
		if s.ensureNotExists != nil && rowExists {
			return s.ensureNotExists
		}
	case kindQuery:
		args := s.effectiveArgs(c)
		var rowsExist bool

		if s.collect != nil {
			rows, err := c.Query(s.query, args...)
			if err != nil {
				return fmt.Errorf("failed to execute query: %w", err)
			}
			defer rows.Close()

			// Check if there are any rows first
			rowsExist = rows.Next()

			// If there are rows, we need to re-execute the query for collect
			// since we already consumed the first row
			if rowsExist {
				rows.Close()
				rows, err = c.Query(s.query, args...)
				if err != nil {
					return fmt.Errorf("failed to re-execute query for collect: %w", err)
				}
			}

			// Execute collect function
			if err = s.collect(rows); err != nil {
				return fmt.Errorf("failed to collect rows: %w", err)
			}
		} else if s.forEach != nil {
			rowCount, err := c.ForEach(s.forEach, s.query, args...)
			if err != nil {
				return fmt.Errorf("failed to process rows: %w", err)
			}
			rowsExist = rowCount > 0
		} else if s.saveAs != "" {
			results, err := c.QueryManyRowMap(s.query, args...)
			if err != nil {
				return fmt.Errorf("failed to execute query for saveAs: %w", err)
			}
			// Save the results
			c.Set(s.saveAs, results)
			rowsExist = len(results) > 0
		} else {
			// No operation specified, just check existence
			rows, err := c.Query(s.query, args...)
			if err != nil {
				rowsExist = false
			} else {
				defer rows.Close()
				rowsExist = rows.Next()
			}
		}

		// Apply the guards
		if s.ensureExists != nil && !rowsExist {
			return s.ensureExists
		}
		if s.ensureNotExists != nil && rowsExist {
			return s.ensureNotExists
		}
	}

	return nil
}

var _ Step = (*sqlStep)(nil)
