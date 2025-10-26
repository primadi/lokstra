package request

import "strings"

// PagingRequest standardizes pagination and data formatting for list APIs
type PagingRequest struct {
	Page       int      `query:"page"`        // default: 1
	PageSize   int      `query:"page_size"`   // default: 20, max: 100
	OrderBy    []string `query:"order_by"`    // e.g. order_by=id,-name
	QueryAll   bool     `query:"all"`         // true → ignore paging
	Fields     []string `query:"fields"`      // e.g. fields=id,name,email
	Search     string   `query:"search"`      // global keyword search
	Filters    []string `query:"filter"`      // e.g. filter=status:active&filter=role:admin
	DataType   string   `query:"data_type"`   // "list" | "table", default "list"
	DataFormat string   `query:"data_format"` // "json", "json_download", "csv", "xlsx"
	Download   bool     `query:"download"`    // true = force download, false = inline
}

// SetDefaults applies default values for PagingRequest
func (p *PagingRequest) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	if p.DataType == "" {
		p.DataType = "list"
	}
	if p.DataFormat == "" {
		p.DataFormat = "json"
	}
}

// GetOffset calculates offset for database queries
func (p *PagingRequest) GetOffset() int {
	if p.QueryAll {
		return 0
	}
	return (p.Page - 1) * p.PageSize
}

// GetLimit returns the limit for database queries
func (p *PagingRequest) GetLimit() int {
	if p.QueryAll {
		return 0 // No limit when QueryAll is true
	}
	return p.PageSize
}

// IsTableFormat returns true if data should be returned as table (2D array)
func (p *PagingRequest) IsTableFormat() bool {
	return p.DataType == "table"
}

// IsDownloadFormat returns true if response should be served as attachment
func (p *PagingRequest) IsDownloadFormat() bool {
	return p.Download || p.DataFormat == "csv" || p.DataFormat == "xlsx" || p.DataFormat == "json_download"
}

// ParseFilters converts filter strings to map for easier processing
// Example: ["status:active", "role:admin"] → {"status": "active", "role": "admin"}
func (p *PagingRequest) ParseFilters() map[string]string {
	filters := make(map[string]string)
	for _, filter := range p.Filters {
		if parts := strings.Split(filter, ":"); len(parts) == 2 {
			filters[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return filters
}

// ParseOrderBy converts order_by strings to structured format
// Example: ["id", "-name"] → [{"field": "id", "desc": false}, {"field": "name", "desc": true}]
func (p *PagingRequest) ParseOrderBy() []OrderField {
	var orders []OrderField
	for _, order := range p.OrderBy {
		if order == "" {
			continue
		}

		field := OrderField{}
		if strings.HasPrefix(order, "-") {
			field.Field = strings.TrimPrefix(order, "-")
			field.Desc = true
		} else {
			field.Field = order
			field.Desc = false
		}
		orders = append(orders, field)
	}
	return orders
}

// OrderField represents a single order by field
type OrderField struct {
	Field string
	Desc  bool
}

// ToSQL converts OrderField to SQL ORDER BY clause
func (o OrderField) ToSQL() string {
	if o.Desc {
		return o.Field + " DESC"
	}
	return o.Field + " ASC"
}
