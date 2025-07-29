package ulspdaemon

import (
	"context"

	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
)

func (r *jsonRPCRouter) CodeAction(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToCodeActionParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.CodeAction(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) CodeLens(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToCodeLensParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.CodeLens(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) CodeLensResolve(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToCodeLens(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.CodeLensResolve(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) CodeLensRefresh(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	err := r.ulspdaemon.CodeLensRefresh(ctx)
	return reply(ctx, nil, err)
}

func (r *jsonRPCRouter) GotoDeclaration(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDeclarationParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.GotoDeclaration(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) GotoDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDefinitionParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.GotoDefinition(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) GotoTypeDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToTypeDefinitionParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.GotoTypeDefinition(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) GotoImplementation(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToImplementationParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.GotoImplementation(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) References(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToReferencesParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.References(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) Hover(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToHoverParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.Hover(ctx, params)
	return reply(ctx, result, err)
}

func (r *jsonRPCRouter) DocumentSymbol(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToDocumentSymbolParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.DocumentSymbol(ctx, params)
	return reply(ctx, result, err)
}
