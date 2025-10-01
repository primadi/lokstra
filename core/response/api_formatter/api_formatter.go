package api_formatter

// ApiResponseFormatter implements structured API response format
type ApiResponseFormatter struct{}

func NewApiResponseFormatter() ResponseFormatter {
	return &ApiResponseFormatter{}
}

func (f *ApiResponseFormatter) Success(data any, message ...string) any {
	resp := &ApiResponse{
		Status: "success",
		Data:   data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	}
	return resp
}

func (f *ApiResponseFormatter) Created(data any, message ...string) any {
	resp := &ApiResponse{
		Status: "success",
		Data:   data,
	}
	if len(message) > 0 {
		resp.Message = message[0]
	} else {
		resp.Message = "Resource created successfully"
	}
	return resp
}

func (f *ApiResponseFormatter) Error(code string, message string, details ...map[string]any) any {
	errorObj := &Error{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		errorObj.Details = details[0]
	}
	return &ApiResponse{
		Status: "error",
		Error:  errorObj,
	}
}

func (f *ApiResponseFormatter) ValidationError(message string, fields []FieldError) any {
	return &ApiResponse{
		Status: "error",
		Error: &Error{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Fields:  fields,
		},
	}
}

func (f *ApiResponseFormatter) NotFound(message string) any {
	return f.Error("NOT_FOUND", message)
}

func (f *ApiResponseFormatter) List(data any, meta *ListMeta) any {
	resp := &ApiResponse{
		Status: "success",
		Data:   data,
	}
	if meta != nil {
		resp.Meta = &Meta{
			ListMeta: meta,
		}
	}
	return resp
}

var _ ResponseFormatter = (*ApiResponseFormatter)(nil)
