package response_iface

import (
	"io"
	"net/http"
)

type ResponseFormatter interface {
	ContentType() string
	WriteHttp(w http.ResponseWriter, r Response) error
	WriteBuffer(w io.Writer, r Response) error
	WriteStdout(r Response) error
}

type Response interface {
	GetStatusCode() int
	GetHeaders() http.Header
}
