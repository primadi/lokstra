package response_iface

import "net/http"

type Response interface {
	GetStatusCode() int
	GetHeaders() http.Header
}
