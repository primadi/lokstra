package response

import "net/http"

// =========================
// ✅ Success Responses
// =========================

// Ok sends a structured response indicating success
func (r *Response) Ok(data any) error {
	r.StatusCode = http.StatusOK
	r.ResponseCode = CodeOK
	r.Success = true
	r.Data = data
	return nil
}

// OkCreated sends a structured response indicating successful creation
func (r *Response) OkCreated(data any) error {
	r.StatusCode = http.StatusCreated
	r.ResponseCode = CodeCreated
	r.Success = true
	r.Data = data
	return nil
}

// OkUpdated sends a structured response indicating successful update
func (r *Response) OkUpdated(data any) error {
	r.StatusCode = http.StatusOK
	r.ResponseCode = CodeUpdated
	r.Success = true
	if r.Message == "" {
		r.Message = "Updated successfully"
	}
	r.Data = data
	return nil
}

// OkList sends a structured list response with optional metadata
func (r *Response) OkList(data any, meta any) error {
	r.StatusCode = http.StatusOK
	r.ResponseCode = CodeOK
	r.Success = true
	r.Data = data
	r.Meta = meta
	return nil
}

// PageData is a structured response for HTMX page data
type PageData struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	Data map[string]any `json:"data,omitempty"`
}

// HtmxPageData sends a structured page data response for HTMX
func (r *Response) HtmxPageData(title string, description string,
	data map[string]any) error {
	r.StatusCode = http.StatusOK
	r.ResponseCode = CodeOK
	r.Success = true
	r.Data = PageData{
		Title:       title,
		Description: description,
		Data:        data,
	}
	return nil
}

// =========================
// ❌ Error Responses
// =========================

// ErrorNotFound sends a structured 404 not found response
func (r *Response) ErrorNotFound(msg string) error {
	r.StatusCode = http.StatusNotFound
	r.ResponseCode = CodeNotFound
	r.Success = false
	r.Message = msg
	return nil
}

// ErrorDuplicate sends a structured 409 conflict response
func (r *Response) ErrorDuplicate(msg string) error {
	r.StatusCode = http.StatusConflict
	r.ResponseCode = CodeDuplicate
	r.Success = false
	r.Message = msg
	return nil
}

// ErrorBadRequest sends a structured 400 bad request response
func (r *Response) ErrorBadRequest(msg string) error {
	r.StatusCode = http.StatusBadRequest
	r.ResponseCode = CodeBadRequest
	r.Success = false
	r.Message = msg
	return nil
}

// ErrorValidation sends a structured 400 bad request response with field errors
func (r *Response) ErrorValidation(globalMsg string, fieldErrors map[string]string) error {
	r.StatusCode = http.StatusBadRequest
	r.ResponseCode = CodeBadRequest
	r.Success = false
	r.Message = globalMsg
	r.FieldErrors = fieldErrors
	return nil
}

// ErrorInternal sends a structured 500 internal server error response
func (r *Response) ErrorInternal(msg string) error {
	r.StatusCode = http.StatusInternalServerError
	r.ResponseCode = CodeInternal
	r.Success = false
	r.Message = msg
	return nil
}

// WriteRaw for custom raw responses
func (r *Response) WriteRaw(contentType string, status int, data []byte) error {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set("Content-Type", contentType)
	r.StatusCode = status
	r.Success = true
	r.Data = data
	r.RawData = data
	return nil
}

// HTML renders HTML content with 200 status
func (r *Response) HTML(html string) error {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	r.StatusCode = http.StatusOK
	r.Success = true
	r.RawData = []byte(html)
	return nil
}

// ErrorHTML renders HTML content with error status and message
func (r *Response) ErrorHTML(status int, html string) error {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	r.StatusCode = status
	r.Success = false
	r.Message = html
	r.RawData = []byte(html)
	return nil
}
