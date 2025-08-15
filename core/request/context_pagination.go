package request

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/primadi/lokstra/common/utils"
)

type PaginationQuery struct {
	Page     int
	PageSize int
	Sort     []SortParam
	Filter   map[string]string
	Fields   []string
}

type SortParam struct {
	Field     string
	Ascending bool
}

var validFieldNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// query parameter format:
// page=1&pageSize=10&sort[name]=asc&filter[status]=active&fields=id,name
func (ctx *Context) BindPaginationQuery() (*PaginationQuery, error) {
	query := ctx.Request.URL.Query()

	page := utils.ParseInt(query.Get("page"), 1)
	pageSize := utils.ParseInt(query.Get("pageSize"), 10)

	sort, err := ctx.sortFieldsAndOrders()
	if err != nil {
		return nil, err
	}

	filter, err := ctx.filterMap()
	if err != nil {
		return nil, err
	}

	fields, err := ctx.selectedFields()
	if err != nil {
		return nil, err
	}

	return &PaginationQuery{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Filter:   filter,
		Fields:   fields,
	}, nil
}

// sort[<fieldname>]=asc|desc
func (ctx *Context) sortFieldsAndOrders() ([]SortParam, error) {
	var result []SortParam

	for pair := range strings.SplitSeq(ctx.Request.URL.RawQuery, "&") {
		if !strings.HasPrefix(pair, "sort[") {
			continue
		}

		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0] // sort[name]
		value := kv[1]

		field := key[5 : len(key)-1] // ambil isi dalam [ ]
		order := strings.ToLower(value)

		if !validFieldNameRegex.MatchString(field) {
			return nil, fmt.Errorf("invalid sort field name: %s", field)
		}

		switch order {
		case "asc":
			result = append(result, SortParam{Field: field, Ascending: true})
		case "desc":
			result = append(result, SortParam{Field: field, Ascending: false})
		default:
			return nil, fmt.Errorf("sort order of field %s must be 'asc' or 'desc'", field)
		}
	}
	return result, nil
}

func (ctx *Context) filterMap() (map[string]string, error) {
	result := make(map[string]string)
	query := ctx.Request.URL.Query()

	for key, values := range query {
		if strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
			field := key[7 : len(key)-1]
			if field == "" {
				return nil, errors.New("invalid filter key format")
			}
			if !validFieldNameRegex.MatchString(field) {
				return nil, errors.New("invalid field name: " + field)
			}
			if len(values) == 0 {
				return nil, errors.New("missing filter value for field " + field)
			}
			result[field] = values[0]
		}
	}

	return result, nil
}

func (ctx *Context) selectedFields() ([]string, error) {
	fieldsParam := ctx.Request.URL.Query().Get("fields")
	if fieldsParam == "" {
		return nil, nil // tidak memilih field = select semua
	}

	fields := strings.Split(fieldsParam, ",")
	cleanFields := make([]string, 0, len(fields))

	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			return nil, errors.New("invalid field name in fields parameter")
		}
		if !validFieldNameRegex.MatchString(f) {
			return nil, errors.New("invalid selected field name: " + f)
		}
		cleanFields = append(cleanFields, f)
	}

	return cleanFields, nil
}
