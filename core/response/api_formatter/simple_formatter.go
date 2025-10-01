package api_formatter

// SimpleResponseFormatter implements simple JSON response format
type SimpleResponseFormatter struct{}

func NewSimpleResponseFormatter() ResponseFormatter {
	return &SimpleResponseFormatter{}
}

func (f *SimpleResponseFormatter) Success(data any, message ...string) any {
	return data
}

func (f *SimpleResponseFormatter) Created(data any, message ...string) any {
	if len(message) > 0 {
		return map[string]any{
			"data":    data,
			"message": message[0],
		}
	}
	return data
}

func (f *SimpleResponseFormatter) Error(code string, message string, details ...map[string]any) any {
	result := map[string]any{
		"error": message,
		"code":  code,
	}
	if len(details) > 0 {
		result["details"] = details[0]
	}
	return result
}

func (f *SimpleResponseFormatter) ValidationError(message string, fields []FieldError) any {
	return map[string]any{
		"error":  message,
		"fields": fields,
	}
}

func (f *SimpleResponseFormatter) NotFound(message string) any {
	return map[string]any{
		"error": message,
	}
}

func (f *SimpleResponseFormatter) List(data any, meta *ListMeta) any {
	if meta != nil {
		return map[string]any{
			"data": data,
			"meta": meta,
		}
	}
	return data
}
