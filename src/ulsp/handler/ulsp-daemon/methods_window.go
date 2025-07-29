package ulspdaemon

import (
	"context"

	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
)

func (r *jsonRPCRouter) WorkDoneProgressCancel(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToWorkDoneProgressCancelParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.WorkDoneProgressCancel(ctx, params)
	return reply(ctx, nil, err)
}
