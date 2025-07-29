package ulspdaemon

import (
	"context"

	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"go.lsp.dev/protocol"
)

func (c *controller) ExecuteCommand(ctx context.Context, params *protocol.ExecuteCommandParams) (interface{}, error) {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.ExecuteCommand(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return nil, c.executePluginMethods(ctx, protocol.MethodWorkspaceExecuteCommand, call, call)
}
