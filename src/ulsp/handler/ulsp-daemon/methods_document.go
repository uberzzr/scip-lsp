package ulspdaemon

import (
	"context"

	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
)

func (r *jsonRPCRouter) DidChange(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDidChangeTextDocumentParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidChange(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) DidChangeWatchedFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDidChangeWatchedFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidChangeWatchedFiles(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) DidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDidOpenTextDocumentParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidOpen(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) DidClose(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDidCloseTextDocumentParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidClose(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) WillSave(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToWillSaveTextDocumentParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.WillSave(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) WillSaveWaitUntil(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToWillSaveTextDocumentParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.WillSaveWaitUntil(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) DidSave(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDidSaveTextDocumentParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidSave(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) WillRenameFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToRenameFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.WillRenameFiles(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) DidRenameFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToRenameFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidRenameFiles(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) WillCreateFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToCreateFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.WillCreateFiles(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) DidCreateFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToCreateFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidCreateFiles(ctx, params)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) WillDeleteFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDeleteFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.WillDeleteFiles(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) DidDeleteFiles(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDeleteFilesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.DidDeleteFiles(ctx, params)
	return reply(ctx, nil, err)
}
