package response

import (
	"errors"
	"net/http"
)

// return JSON response with status code 200
func (r *Response) Ok(data any) error {
	r.StatusCode = http.StatusOK
	return r.Json(data)
}

// return JSON response with status code 201
func (r *Response) OkCreated(data any) error {
	r.StatusCode = http.StatusCreated
	return r.Json(data)
}

// return no content with status code 204
func (r *Response) OkNoContent() error {
	r.Data = nil
	r.StatusCode = http.StatusNoContent
	return nil
}

// return JSON error response with status code 400
func (r *Response) ErrorBadRequest(err error) error {
	if jerr := r.Json(map[string]string{"error": err.Error()}); jerr != nil {
		return jerr
	}
	r.StatusCode = http.StatusBadRequest
	return err
}

// return JSON error response with status code 422
func (r *Response) ErrorUnauthorized(err error) error {
	if jerr := r.Json(map[string]string{"error": err.Error()}); jerr != nil {
		return jerr
	}
	r.StatusCode = http.StatusUnauthorized
	return err
}

// return JSON error response with status code 403
func (r *Response) ErrorForbidden(err error) error {
	if jerr := r.Json(map[string]string{"error": err.Error()}); jerr != nil {
		return jerr
	}
	r.StatusCode = http.StatusForbidden
	return err
}

// return JSON error response with status code 404
func (r *Response) ErrorNotFound(err error) error {
	if jerr := r.Json(map[string]string{"error": err.Error()}); jerr != nil {
		return jerr
	}
	r.StatusCode = http.StatusNotFound
	return err
}

// return JSON error response with status code 409
func (r *Response) ErrorConflict(err error) error {
	if jerr := r.Json(map[string]string{"error": err.Error()}); jerr != nil {
		return jerr
	}
	r.StatusCode = http.StatusConflict
	return err
}

// return JSON error response with status code 500
func (r *Response) ErrorInternal(err error) error {
	if err == nil {
		err = errors.New("internal server error")
	}
	if jerr := r.Json(map[string]string{"error": err.Error()}); jerr != nil {
		return jerr
	}
	r.StatusCode = http.StatusInternalServerError
	return err
}
