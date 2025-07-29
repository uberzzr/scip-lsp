package ulspdaemon

import (
	"context"

	"github.com/uber/scip-lsp/src/ulsp/mapper"
	"go.lsp.dev/jsonrpc2"
)

// Initialize extracts protocol.InitializeParams from the request and calls initialization logic for a new IDE connection.
func (r *jsonRPCRouter) Initialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToInitializeParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	result, err := r.ulspdaemon.Initialize(ctx, params)
	if err != nil {
		return reply(ctx, nil, err)
	}

	return reply(ctx, result, nil)
}

// Initialized is sent after the client received the result of the initialize request but before the client sends any other request or notification.
func (r *jsonRPCRouter) Initialized(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	params, err := mapper.RequestToInitializedParams(req)
	if err != nil {
		return reply(ctx, nil, err)
	}

	err = r.ulspdaemon.Initialized(ctx, params)
	return reply(ctx, nil, err)
}

// Shutdown asks the server to shut down, but to not exit.
// RequestFullShutdown must be sent first if full shutdown is needed, otherwise it will be used only to clean up from that specific client.
func (r *jsonRPCRouter) Shutdown(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	err := r.ulspdaemon.Shutdown(ctx)
	return reply(ctx, nil, err)
}

// Exit asks the server to exit its process.
// Because the server is shared between multiple IDE processes, the process will only exit when RequestFullShutdown is sent first.
func (r *jsonRPCRouter) Exit(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	// Reply first to ensure that a reply is sent before the controller initiates the shutdown.
	reply(ctx, nil, nil)
	err := r.ulspdaemon.Exit(ctx)
	return err
}

// RequestFullShutdown will indicate that the next Shutdown and Exit requests should perform a full shutdown and exit of the server.
func (r *jsonRPCRouter) RequestFullShutdown(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	err := r.ulspdaemon.RequestFullShutdown(ctx)
	return reply(ctx, nil, err)
}
