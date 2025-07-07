package router

import (
	"context"
	"lokstra/core/request"
	"lokstra/core/response"
	"net/http"
)

func NewContext(w http.ResponseWriter, r *http.Request) (*request.Context, func()) {
	ctx, cancel := context.WithCancel(r.Context())
	req := r.WithContext(ctx)
	resp := response.NewResponse()

	return &request.Context{
		Context:  ctx,
		Response: resp,
		W:        w,
		R:        req,
	}, cancel
}
