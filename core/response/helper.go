package response

import "net/http"

// =========================
// ✅ Success Responses
// =========================

func (r *Response) Ok(data any) error {
	r.StatusCode = http.StatusOK
	r.ResponseCode = CodeOK
	r.Success = true
	r.Data = data
	return nil
}

func (r *Response) OkCreated(data any) error {
	r.StatusCode = http.StatusCreated
	r.ResponseCode = CodeCreated
	r.Success = true
	r.Data = data
	return nil
}

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

func (r *Response) OkList(data any, meta any) error {
	r.StatusCode = http.StatusOK
	r.ResponseCode = CodeOK
	r.Success = true
	r.Data = data
	r.Meta = meta
	return nil
}

// =========================
// ❌ Error Responses
// =========================

func (r *Response) ErrorNotFound(msg string) error {
	r.StatusCode = http.StatusNotFound
	r.ResponseCode = CodeNotFound
	r.Success = false
	r.Message = msg
	return nil
}

func (r *Response) ErrorDuplicate(msg string) error {
	r.StatusCode = http.StatusConflict
	r.ResponseCode = CodeDuplicate
	r.Success = false
	r.Message = msg
	return nil
}

func (r *Response) ErrorBadRequest(msg string) error {
	r.StatusCode = http.StatusBadRequest
	r.ResponseCode = CodeBadRequest
	r.Success = false
	r.Message = msg
	return nil
}

func (r *Response) ErrorValidation(globalMsg string, fieldErrors map[string]string) error {
	r.StatusCode = http.StatusBadRequest
	r.ResponseCode = CodeBadRequest
	r.Success = false
	r.Message = globalMsg
	r.FieldErrors = fieldErrors
	return nil
}

func (r *Response) ErrorInternal(msg string) error {
	r.StatusCode = http.StatusInternalServerError
	r.ResponseCode = CodeInternal
	r.Success = false
	r.Message = msg
	return nil
}

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
func (r *Response) HTML(status int, html string) error {
	if r.Headers == nil {
		r.Headers = make(http.Header)
	}
	r.Headers.Set("Content-Type", "text/html; charset=utf-8")
	r.StatusCode = status
	r.Success = status < 400
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
