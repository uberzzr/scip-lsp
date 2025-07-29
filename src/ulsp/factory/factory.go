package factory

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// UUID is a user-defined factory for a random uuid.UUID.
func UUID() uuid.UUID {
	return uuid.Must(uuid.NewV4())
}

// JSONRPCRequest is a user-defined factory for a JSON-RPC request containing the specified method and parameters.
func JSONRPCRequest(method string, params interface{}) jsonrpc2.Request {
	req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), method, params)
	return req
}

// PluginInfoValid is a factory for PluginInfo that passes validation.
func PluginInfoValid(id int) ulspplugin.PluginInfo {
	sampleDidOpenFunc := func(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
		return nil
	}
	return ulspplugin.PluginInfo{
		Priorities: map[string]ulspplugin.Priority{
			protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityHigh,
		},
		Methods: &ulspplugin.Methods{
			PluginNameKey: fmt.Sprintf("test-plugin-%v", id),

			DidOpen: sampleDidOpenFunc,
		},
		NameKey: fmt.Sprintf("test-plugin-%v", id),
	}
}

// PluginInfoInvalid is a factory for PluginInfo that fails validation.
func PluginInfoInvalid(id int) ulspplugin.PluginInfo {
	return ulspplugin.PluginInfo{
		Priorities: map[string]ulspplugin.Priority{
			protocol.MethodTextDocumentDidOpen: ulspplugin.PriorityHigh,
		},
		Methods: &ulspplugin.Methods{},
		NameKey: fmt.Sprintf("test-plugin-%v", id),
	}
}
