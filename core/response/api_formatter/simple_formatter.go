package api_formatter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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

func (f *SimpleResponseFormatter) ParseClientResponse(resp *http.Response, cr *ClientResponse) error {
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Repository raw body and status code
	cr.RawBody = body
	cr.StatusCode = resp.StatusCode

	// Parse headers (optional)
	cr.Headers = make(map[string]any)
	for key, values := range resp.Header {
		if len(values) == 1 {
			cr.Headers[key] = values[0]
		} else {
			cr.Headers[key] = values
		}
	}

	// Try to parse as JSON
	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		// If not valid JSON, treat as plain text
		cr.Status = "success"
		cr.Data = string(body)
		return nil
	}

	// Check for error field (simple error format)
	if errorMsg, hasError := result["error"]; hasError {
		cr.Status = "error"
		cr.Error = &Error{
			Code:    "ERROR",
			Message: fmt.Sprint(errorMsg),
		}
		if code, hasCode := result["code"]; hasCode {
			cr.Error.Code = fmt.Sprint(code)
		}
		if details, hasDetails := result["details"]; hasDetails {
			if detailsMap, ok := details.(map[string]any); ok {
				cr.Error.Details = detailsMap
			}
		}
		if fields, hasFields := result["fields"]; hasFields {
			if fieldsSlice, ok := fields.([]any); ok {
				cr.Error.Fields = make([]FieldError, 0, len(fieldsSlice))
				for _, field := range fieldsSlice {
					if fieldMap, ok := field.(map[string]any); ok {
						fe := FieldError{}
						if f, ok := fieldMap["field"].(string); ok {
							fe.Field = f
						}
						if c, ok := fieldMap["code"].(string); ok {
							fe.Code = c
						}
						if m, ok := fieldMap["message"].(string); ok {
							fe.Message = m
						}
						fe.Value = fieldMap["value"]
						cr.Error.Fields = append(cr.Error.Fields, fe)
					}
				}
			}
		}
	} else {
		// Success response
		cr.Status = "success"

		// Check if there's a "data" field
		if data, hasData := result["data"]; hasData {
			cr.Data = data
			// Check for message
			if msg, hasMsg := result["message"]; hasMsg {
				cr.Message = fmt.Sprint(msg)
			}
			// Check for meta
			if meta, hasMeta := result["meta"]; hasMeta {
				if metaMap, ok := meta.(map[string]any); ok {
					metaBytes, _ := json.Marshal(metaMap)
					var parsedMeta Meta
					if json.Unmarshal(metaBytes, &parsedMeta) == nil {
						cr.Meta = &parsedMeta
					}
				}
			}
		} else {
			// Entire result is the data
			cr.Data = result
		}
	}

	return nil
}
