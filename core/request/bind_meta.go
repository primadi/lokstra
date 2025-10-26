package request

import (
	"reflect"
	"sync"
)

type bindFieldMeta struct {
	Field           reflect.StructField
	Index           []int
	Name            string // param name
	Tag             string // path/query/header/json
	IsSlice         bool
	IsUnmarshalJSON bool

	IsIndexedKeyValue bool
	IndexKey          []int
	IndexValue        []int
	IsMap             bool
	IsWildcard        bool // true if json:"*" - captures all body as map
}

type bindMeta struct {
	Type   reflect.Type
	Fields []bindFieldMeta
}

var bindMetaCache sync.Map // map[reflect.Type]*bindMeta

func getOrBuildBindMeta(t reflect.Type) *bindMeta {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return &bindMeta{Type: t, Fields: nil}
	}

	if bm, ok := bindMetaCache.Load(t); ok {
		return bm.(*bindMeta)
	}

	bm := &bindMeta{
		Type: t,
	}

	numField := t.NumField()
	for i := range numField {
		field := t.Field(i)

		// If this field is an anonymous (embedded) struct, iterate its inner fields
		// and create combined index entries so embedded struct fields (like
		// request.PagingRequest) are discoverable for binding.
		if field.Anonymous {
			ft := field.Type
			if ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				// iterate inner struct fields
				innerNum := ft.NumField()
				for j := range innerNum {
					inner := ft.Field(j)
					tagType, paramName, isWildcard := parseBindingTag(inner)
					if tagType == "" {
						continue
					}

					// build combined index (outer field index + inner field index)
					combinedIndex := make([]int, 0, len(field.Index)+len(inner.Index))
					combinedIndex = append(combinedIndex, field.Index...)
					combinedIndex = append(combinedIndex, inner.Index...)

					// Determine indexed key/value for slice-of-struct inner fields
					isIndexedKeyValue := false
					var indexKey, indexValue []int
					if inner.Type.Kind() == reflect.Slice && inner.Type.Elem().Kind() == reflect.Struct {
						structType := inner.Type.Elem()
						for k := 0; k < structType.NumField(); k++ {
							f := structType.Field(k)
							switch f.Name {
							case "Key", "Field":
								indexKey = f.Index
							case "Value":
								indexValue = f.Index
							}
						}
						if len(indexKey) > 0 && len(indexValue) > 0 {
							isIndexedKeyValue = true
						}
					}

					isMap := false
					if inner.Type.Kind() == reflect.Map && inner.Type.Key().Kind() == reflect.String {
						isMap = true
					}

					fieldMeta := bindFieldMeta{
						Field:             inner,
						Index:             combinedIndex,
						Name:              paramName,
						Tag:               tagType,
						IsSlice:           inner.Type.Kind() == reflect.Slice,
						IsUnmarshalJSON:   implementsUnmarshalJSON(inner.Type),
						IsIndexedKeyValue: isIndexedKeyValue,
						IndexKey:          indexKey,
						IndexValue:        indexValue,
						IsMap:             isMap,
						IsWildcard:        isWildcard,
					}
					bm.Fields = append(bm.Fields, fieldMeta)
				}
				// continue to next top-level field
				continue
			}
		}

		tagType, paramName, isWildcard := parseBindingTag(field)
		if tagType == "" {
			continue
		}

		isIndexedKeyValue := false
		var indexKey, indexValue []int
		// Check if it's a slice of struct with Key and Value fields
		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
			structType := field.Type.Elem()
			for k := 0; k < structType.NumField(); k++ {
				f := structType.Field(k)
				switch f.Name {
				case "Key", "Field":
					indexKey = f.Index
				case "Value":
					indexValue = f.Index
				}
			}

			if len(indexKey) > 0 && len(indexValue) > 0 {
				isIndexedKeyValue = true
			}
		}

		// Check if it's a map[string]string or map[string]any
		isMap := false
		if field.Type.Kind() == reflect.Map && field.Type.Key().Kind() == reflect.String {
			isMap = true
		}

		fieldMeta := bindFieldMeta{
			Field:             field,
			Index:             field.Index,
			Name:              paramName,
			Tag:               tagType,
			IsSlice:           field.Type.Kind() == reflect.Slice,
			IsUnmarshalJSON:   implementsUnmarshalJSON(field.Type),
			IsIndexedKeyValue: isIndexedKeyValue,
			IndexKey:          indexKey,
			IndexValue:        indexValue,
			IsMap:             isMap,
			IsWildcard:        isWildcard,
		}

		bm.Fields = append(bm.Fields, fieldMeta)
	}

	actual, loaded := bindMetaCache.LoadOrStore(t, bm)
	if loaded {
		return actual.(*bindMeta)
	}
	return bm
}

func parseBindingTag(field reflect.StructField) (tagType, paramName string, isWildcard bool) {
	// Check for path, query, header tags
	for _, key := range []string{"path", "query", "header"} {
		if val, ok := field.Tag.Lookup(key); ok && val != "" {
			return key, val, false
		}
	}

	// Check for json tag (for body binding)
	if val, ok := field.Tag.Lookup("json"); ok && val != "" {
		// Check for wildcard: json:"*"
		if val == "*" {
			return "json", "", true
		}
		return "json", val, false
	}

	return "", "", false
} // unmarshalJSONType represents the interface type for json.Unmarshaler
var unmarshalJSONType = reflect.TypeOf((*interface {
	UnmarshalJSON([]byte) error
})(nil)).Elem()

func implementsUnmarshalJSON(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Check if the type or its pointer implements UnmarshalJSON
	return t.Implements(unmarshalJSONType) || reflect.PointerTo(t).Implements(unmarshalJSONType)
}
