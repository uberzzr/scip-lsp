package ulspplugin

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		priorities []map[string]Priority
		methods    *Methods
		nameKey    string
		wantErr    bool
	}{
		{
			name: "valid info",
			priorities: []map[string]Priority{
				{
					protocol.MethodInitialize:                     PriorityHigh,
					protocol.MethodInitialized:                    PriorityHigh,
					protocol.MethodShutdown:                       PriorityHigh,
					protocol.MethodExit:                           PriorityHigh,
					protocol.MethodTextDocumentDidOpen:            PriorityHigh,
					protocol.MethodTextDocumentDidChange:          PriorityHigh,
					protocol.MethodWorkspaceDidChangeWatchedFiles: PriorityHigh,
					protocol.MethodTextDocumentDidClose:           PriorityHigh,
					protocol.MethodTextDocumentWillSave:           PriorityHigh,
					protocol.MethodTextDocumentWillSaveWaitUntil:  PriorityHigh,
					protocol.MethodTextDocumentDidSave:            PriorityHigh,
					protocol.MethodWillRenameFiles:                PriorityHigh,
					protocol.MethodDidRenameFiles:                 PriorityHigh,
					protocol.MethodWillCreateFiles:                PriorityHigh,
					protocol.MethodDidCreateFiles:                 PriorityHigh,
					protocol.MethodWillDeleteFiles:                PriorityHigh,
					protocol.MethodDidDeleteFiles:                 PriorityHigh,
					protocol.MethodTextDocumentCodeAction:         PriorityHigh,
					protocol.MethodTextDocumentCodeLens:           PriorityHigh,
					protocol.MethodCodeLensRefresh:                PriorityHigh,
					protocol.MethodCodeLensResolve:                PriorityHigh,
					protocol.MethodWorkspaceExecuteCommand:        PriorityHigh,
					protocol.MethodTextDocumentDefinition:         PriorityHigh,
					protocol.MethodTextDocumentDeclaration:        PriorityHigh,
					protocol.MethodTextDocumentTypeDefinition:     PriorityHigh,
					protocol.MethodTextDocumentImplementation:     PriorityHigh,
					protocol.MethodTextDocumentReferences:         PriorityHigh,
					protocol.MethodTextDocumentHover:              PriorityHigh,
					protocol.MethodTextDocumentDocumentSymbol:     PriorityHigh,
					MethodEndSession:                              PriorityHigh,
				},
			},
			methods: &Methods{
				PluginNameKey: "sample-plugin",
				Initialize: func(ctx context.Context, params *protocol.InitializeParams, result *protocol.InitializeResult) error {
					return nil
				},
				Initialized:           func(ctx context.Context, params *protocol.InitializedParams) error { return nil },
				Shutdown:              func(ctx context.Context) error { return nil },
				Exit:                  func(ctx context.Context) error { return nil },
				DidChange:             func(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error { return nil },
				DidChangeWatchedFiles: func(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error { return nil },
				DidOpen:               func(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error { return nil },
				DidClose:              func(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error { return nil },
				WillSave:              func(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error { return nil },
				WillSaveWaitUntil: func(ctx context.Context, params *protocol.WillSaveTextDocumentParams, result *[]protocol.TextEdit) error {
					return nil
				},
				DidSave: func(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error { return nil },
				WillRenameFiles: func(ctx context.Context, params *protocol.RenameFilesParams, result *protocol.WorkspaceEdit) error {
					return nil
				},
				DidRenameFiles: func(ctx context.Context, params *protocol.RenameFilesParams) error { return nil },
				WillCreateFiles: func(ctx context.Context, params *protocol.CreateFilesParams, result *protocol.WorkspaceEdit) error {
					return nil
				},
				DidCreateFiles: func(ctx context.Context, params *protocol.CreateFilesParams) error { return nil },
				WillDeleteFiles: func(ctx context.Context, params *protocol.DeleteFilesParams, result *protocol.WorkspaceEdit) error {
					return nil
				},
				DidDeleteFiles: func(ctx context.Context, params *protocol.DeleteFilesParams) error { return nil },
				EndSession:     func(ctx context.Context, uuid uuid.UUID) error { return nil },
				CodeAction: func(ctx context.Context, params *protocol.CodeActionParams, result *[]protocol.CodeAction) error {
					return nil
				},
				CodeLens: func(ctx context.Context, params *protocol.CodeLensParams, result *[]protocol.CodeLens) error {
					return nil
				},
				CodeLensRefresh: func(ctx context.Context) error { return nil },
				CodeLensResolve: func(ctx context.Context, params *protocol.CodeLens, result *protocol.CodeLens) error { return nil },
				ExecuteCommand:  func(ctx context.Context, params *protocol.ExecuteCommandParams) error { return nil },
				GotoDeclaration: func(ctx context.Context, params *protocol.DeclarationParams, result *[]protocol.LocationLink) error {
					return nil
				},
				GotoDefinition: func(ctx context.Context, params *protocol.DefinitionParams, result *[]protocol.LocationLink) error {
					return nil
				},
				GotoTypeDefinition: func(ctx context.Context, params *protocol.TypeDefinitionParams, result *[]protocol.LocationLink) error {
					return nil
				},
				GotoImplementation: func(ctx context.Context, params *protocol.ImplementationParams, result *[]protocol.LocationLink) error {
					return nil
				},
				References: func(ctx context.Context, params *protocol.ReferenceParams, result *[]protocol.Location) error {
					return nil
				},
				Hover: func(ctx context.Context, params *protocol.HoverParams, result *protocol.Hover) error {
					return nil
				},
				DocumentSymbol: func(ctx context.Context, params *protocol.DocumentSymbolParams, result *[]protocol.DocumentSymbol) error {
					return nil
				},
			},
			nameKey: "sample-plugin",
			wantErr: false,
		},
		{
			name: "nil method",
			priorities: []map[string]Priority{
				{
					protocol.MethodInitialize: PriorityHigh,
				},
				{
					protocol.MethodInitialized: PriorityHigh,
				},
				{
					protocol.MethodShutdown: PriorityHigh,
				},
				{
					protocol.MethodExit: PriorityHigh,
				},
				{
					protocol.MethodShutdown: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDidChange: PriorityHigh,
				},
				{
					protocol.MethodWorkspaceDidChangeWatchedFiles: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDidOpen: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDidClose: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentWillSave: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentWillSaveWaitUntil: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDidSave: PriorityHigh,
				},
				{
					protocol.MethodWillRenameFiles: PriorityHigh,
				},
				{
					protocol.MethodDidRenameFiles: PriorityHigh,
				},
				{
					protocol.MethodWillCreateFiles: PriorityHigh,
				},
				{
					protocol.MethodDidCreateFiles: PriorityHigh,
				},
				{
					protocol.MethodWillDeleteFiles: PriorityHigh,
				},
				{
					protocol.MethodDidDeleteFiles: PriorityHigh,
				},
				{
					protocol.MethodDidDeleteFiles: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentCodeAction: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentCodeLens: PriorityHigh,
				},
				{
					protocol.MethodCodeLensRefresh: PriorityHigh,
				},
				{
					protocol.MethodCodeLensResolve: PriorityHigh,
				},
				{
					protocol.MethodWorkspaceExecuteCommand: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDefinition: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDeclaration: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDeclaration: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentTypeDefinition: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentImplementation: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentReferences: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentHover: PriorityHigh,
				},
				{
					protocol.MethodTextDocumentDocumentSymbol: PriorityHigh,
				},
				{
					MethodEndSession: PriorityHigh,
				},
			},
			methods: &Methods{PluginNameKey: "sample-plugin"},
			nameKey: "sample-plugin",
			wantErr: true,
		},
		{
			name: "empty priorities",
			priorities: []map[string]Priority{
				{},
			},
			methods: &Methods{PluginNameKey: "sample-plugin"},
			nameKey: "sample-plugin",
			wantErr: true,
		},
		{
			name: "missing all methods",
			priorities: []map[string]Priority{
				{
					protocol.MethodTextDocumentDidOpen: PriorityHigh,
				},
			},
			nameKey: "sample-plugin",
			wantErr: true,
		},
		{
			name: "unknown method name",
			priorities: []map[string]Priority{
				{
					"invalidMethodName": PriorityHigh,
				},
			},
			nameKey: "sample-plugin",
			methods: &Methods{PluginNameKey: "sample-plugin"},
			wantErr: true,
		},
		{
			name: "missing name key",
			priorities: []map[string]Priority{
				{
					protocol.MethodTextDocumentDidOpen: PriorityHigh,
				},
			},
			methods: &Methods{PluginNameKey: "sample-plugin"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, p := range tt.priorities {
				info := PluginInfo{
					Priorities: p,
					Methods:    tt.methods,
					NameKey:    tt.nameKey,
				}

				err := info.Validate()
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
