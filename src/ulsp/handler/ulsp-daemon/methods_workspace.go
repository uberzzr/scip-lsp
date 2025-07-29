package ulspdaemon

import (
	"context"

	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
)

func (r *jsonRPCRouter) ExecuteCommand(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToExecuteCommandParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.ExecuteCommand(ctx, params)
	return reply(ctx, result, err)
}
