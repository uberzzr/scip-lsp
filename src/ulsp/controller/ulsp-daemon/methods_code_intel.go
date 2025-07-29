package ulspdaemon

import (
	"context"
	"fmt"

	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"go.lsp.dev/protocol"
)

func (c *controller) CodeAction(ctx context.Context, params *protocol.CodeActionParams) ([]protocol.CodeAction, error) {
	result := []protocol.CodeAction{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeAction(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeAction(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentCodeAction, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) CodeLens(ctx context.Context, params *protocol.CodeLensParams) ([]protocol.CodeLens, error) {
	result := []protocol.CodeLens{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeLens(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeLens(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentCodeLens, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) CodeLensRefresh(ctx context.Context) (err error) {
	call := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeLensRefresh(ctx); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	return c.executePluginMethods(ctx, protocol.MethodCodeLensRefresh, call, call)
}

func (c *controller) CodeLensResolve(ctx context.Context, params *protocol.CodeLens) (*protocol.CodeLens, error) {
	result := &protocol.CodeLens{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeLensResolve(ctx, params, result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.CodeLensResolve(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodCodeLensResolve, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) GotoDeclaration(ctx context.Context, params *protocol.DeclarationParams) ([]protocol.LocationLink, error) {
	result := make([]protocol.LocationLink, 0)

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoDeclaration(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoDeclaration(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentDeclaration, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) GotoDefinition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.LocationLink, error) {
	result := make([]protocol.LocationLink, 0)

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoDefinition(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoDefinition(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentDefinition, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) GotoTypeDefinition(ctx context.Context, params *protocol.TypeDefinitionParams) ([]protocol.LocationLink, error) {
	result := make([]protocol.LocationLink, 0)

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoTypeDefinition(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoTypeDefinition(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentTypeDefinition, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) GotoImplementation(ctx context.Context, params *protocol.ImplementationParams) ([]protocol.LocationLink, error) {
	result := make([]protocol.LocationLink, 0)

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoImplementation(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.GotoImplementation(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentImplementation, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) References(ctx context.Context, params *protocol.ReferenceParams) ([]protocol.Location, error) {
	result := make([]protocol.Location, 0)

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.References(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.References(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentReferences, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}

func (c *controller) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
	result := &protocol.Hover{}

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Hover(ctx, params, result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.Hover(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentHover, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	// Send nil back for empty hovers to avoid an error message or empty popup
	if result.Contents.Value == "" {
		result = nil
	}

	return result, nil
}

func (c *controller) DocumentSymbol(ctx context.Context, params *protocol.DocumentSymbolParams) ([]protocol.DocumentSymbol, error) {
	result := make([]protocol.DocumentSymbol, 0)

	callSync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DocumentSymbol(ctx, params, &result); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}
	callAsync := func(ctx context.Context, m *ulspplugin.Methods) {
		if err := m.DocumentSymbol(ctx, params, nil); err != nil {
			c.logger.Errorf(_errPluginReturnedError, m.PluginNameKey, err)
		}
	}

	if err := c.executePluginMethods(ctx, protocol.MethodTextDocumentDocumentSymbol, callSync, callAsync); err != nil {
		return nil, fmt.Errorf(_errBadPluginCall, err)
	}

	return result, nil
}
