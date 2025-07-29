package ulspdaemon

import (
	"context"
	"fmt"

	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"go.lsp.dev/protocol"
)

func (c *controller) DidChange(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidChange(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodTextDocumentDidChange, call, call)
}

func (c *controller) DidChangeWatchedFiles(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidChangeWatchedFiles(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodWorkspaceDidChangeWatchedFiles, call, call)
}

func (c *controller) DidOpen(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidOpen(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodTextDocumentDidOpen, call, call)
}

func (c *controller) DidClose(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidClose(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodTextDocumentDidClose, call, call)
}

func (c *controller) WillSave(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillSave(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodTextDocumentWillSave, call, call)
}

func (c *controller) WillSaveWaitUntil(ctx context.Context, params *protocol.WillSaveTextDocumentParams) ([]protocol.TextEdit, error) {
	result := []protocol.TextEdit{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillSaveWaitUntil(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillSaveWaitUntil(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentWillSaveWaitUntil, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) DidSave(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidSave(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodTextDocumentDidSave, call, call)
}

func (c *controller) WillRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) (*protocol.WorkspaceEdit, error) {
	result := &protocol.WorkspaceEdit{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillRenameFiles(ctx, params, result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillRenameFiles(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodWillRenameFiles, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) DidRenameFiles(ctx context.Context, params *protocol.RenameFilesParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidRenameFiles(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodDidRenameFiles, call, call)
}

func (c *controller) WillCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) (*protocol.WorkspaceEdit, error) {
	result := &protocol.WorkspaceEdit{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillCreateFiles(ctx, params, result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillCreateFiles(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodWillCreateFiles, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) DidCreateFiles(ctx context.Context, params *protocol.CreateFilesParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidCreateFiles(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodDidCreateFiles, call, call)
}

func (c *controller) WillDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) (*protocol.WorkspaceEdit, error) {
	result := &protocol.WorkspaceEdit{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillDeleteFiles(ctx, params, result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.WillDeleteFiles(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodWillDeleteFiles, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) DidDeleteFiles(ctx context.Context, params *protocol.DeleteFilesParams) error {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DidDeleteFiles(ctx, params); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodDidDeleteFiles, call, call)
}
