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
