package ulspdaemon

import (
	"context"

	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"go.lsp.dev/protocol"
)

func (c *controller) WorkDoneProgressCancel(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WorkDoneProgressCancel(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodWorkDoneProgressCancel, call, call)
}
