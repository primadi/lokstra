package core

import (
	"context"
	"lokstra/common/response"
	"lokstra/core/request"
	"net/http"
)

func NewRequestContext(w http.ResponseWriter, r *http.Request) (*request.Context, func()) {
	ctx, cancel := context.WithCancel(r.Context())
	req := r.WithContext(ctx)
	resp := response.NewResponse()

	return globalRuntime.requestContextHelper(
		&request.Context{
			Context:  ctx,
			Response: resp,
			W:        w,
			R:        req,
		}), cancel
}
