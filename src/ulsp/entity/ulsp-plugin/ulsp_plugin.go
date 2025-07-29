package ulspplugin

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	"go.lsp.dev/protocol"
)

const (
	_errorUnrecognizedMethod = "%q included in priority config, but is not a recognized method. Method name must be a valid LSP method. If method is new to uLSP, ensure that ulspplugin.Validate is updated."
	_errorMissingMethod      = "%q is included in the priority configuration, but is nil in Methods"
	_errorMissingField       = "missing %q field for this plugin"

	// MethodEndSession is an additional method outside of LSP protocol, which is called when the JSON-RPC connection has been closed.
	// This should be used to ensure cleanup of resources even if the client exits before calling 'shutdown' and 'exit'.
	MethodEndSession = "end_session"
)

// RuntimePrioritizedMethods represents ordered list of modules to run for a given method.
type RuntimePrioritizedMethods map[string]MethodLists

// MethodLists maintains ordered list of modules to run, segmented by sync and async.
type MethodLists struct {
	Sync  []*Methods
	Async []*Methods
}

// Priority represents the ranked priority in which a plugin method will be run for a given method.
type Priority int64

const (
	// PriorityHigh for plugin methods that should be run in the highest priority group.
	PriorityHigh Priority = iota
	// PriorityRegular for plugins methods that should be run with regular priority.
	PriorityRegular
	// PriorityAsync for plugin methods should be run asynchronously and won't be included in the response.
	PriorityAsync
)

// Plugin defines a plugin which contributes a portion of language server functionality.
type Plugin interface {
	StartupInfo(ctx context.Context) (PluginInfo, error)
}

// Methods defines methods which can be optionally implemented by a module, based on the protocol.Server interface.
type Methods struct {
	// PluginNameKey identifies the name of the plugin that provides these method implementations.
	PluginNameKey string

	// Lifecycle related methods.
	Initialize  func(ctx context.Context, params *protocol.InitializeParams, result *protocol.InitializeResult) error
	Initialized func(ctx context.Context, params *protocol.InitializedParams) error
	Shutdown    func(ctx context.Context) error
	Exit        func(ctx context.Context) error

	// Document related methods.
	DidChange             func(ctx context.Context, params *protocol.DidChangeTextDocumentParams) error
	DidChangeWatchedFiles func(ctx context.Context, params *protocol.DidChangeWatchedFilesParams) error
	DidOpen               func(ctx context.Context, params *protocol.DidOpenTextDocumentParams) error
	DidClose              func(ctx context.Context, params *protocol.DidCloseTextDocumentParams) error
	WillSave              func(ctx context.Context, params *protocol.WillSaveTextDocumentParams) error
	WillSaveWaitUntil     func(ctx context.Context, params *protocol.WillSaveTextDocumentParams, result *[]protocol.TextEdit) error
	DidSave               func(ctx context.Context, params *protocol.DidSaveTextDocumentParams) error
	WillRenameFiles       func(ctx context.Context, params *protocol.RenameFilesParams, result *protocol.WorkspaceEdit) error
	DidRenameFiles        func(ctx context.Context, params *protocol.RenameFilesParams) error
	WillCreateFiles       func(ctx context.Context, params *protocol.CreateFilesParams, result *protocol.WorkspaceEdit) error
	DidCreateFiles        func(ctx context.Context, params *protocol.CreateFilesParams) error
	WillDeleteFiles       func(ctx context.Context, params *protocol.DeleteFilesParams, result *protocol.WorkspaceEdit) error
	DidDeleteFiles        func(ctx context.Context, params *protocol.DeleteFilesParams) error

	// Codeintel related methods.
	CodeAction      func(ctx context.Context, params *protocol.CodeActionParams, result *[]protocol.CodeAction) error
	CodeLens        func(ctx context.Context, params *protocol.CodeLensParams, result *[]protocol.CodeLens) error
	CodeLensRefresh func(ctx context.Context) error
	CodeLensResolve func(ctx context.Context, params *protocol.CodeLens, result *protocol.CodeLens) error

	GotoDeclaration    func(ctx context.Context, params *protocol.DeclarationParams, result *[]protocol.LocationLink) error
	GotoDefinition     func(ctx context.Context, params *protocol.DefinitionParams, result *[]protocol.LocationLink) error
	GotoTypeDefinition func(ctx context.Context, params *protocol.TypeDefinitionParams, result *[]protocol.LocationLink) error
	GotoImplementation func(ctx context.Context, params *protocol.ImplementationParams, result *[]protocol.LocationLink) error
	References         func(ctx context.Context, params *protocol.ReferenceParams, result *[]protocol.Location) error
	Hover              func(ctx context.Context, params *protocol.HoverParams, result *protocol.Hover) error
	DocumentSymbol     func(ctx context.Context, params *protocol.DocumentSymbolParams, result *[]protocol.DocumentSymbol) error

	// Workspace related methods.
	ExecuteCommand func(ctx context.Context, params *protocol.ExecuteCommandParams) error

	// Window related Features
	WorkDoneProgressCancel func(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error

	// Connection related methods outside of the LSP protocol.
	EndSession func(ctx context.Context, uuid uuid.UUID) error
}

// PluginInfo provides both prioritization for each method, as well as access to call each method implemented by this plugin.
type PluginInfo struct {
	Priorities map[string]Priority
	Methods    *Methods
	NameKey    string
	// Optional set of monorepos for which this plugin should be enabled.
	RelevantRepos map[entity.MonorepoName]struct{}
}

// Validate provides runtime validation that the a Plugin implementation returns valid PluginInfo.
func (m *PluginInfo) Validate() error {
	// Required fields.
	if len(m.Priorities) == 0 {
		return fmt.Errorf(_errorMissingField, "Priorities")
	} else if m.Methods == nil {
		return fmt.Errorf(_errorMissingField, "Methods")
	} else if m.NameKey == "" {
		return fmt.Errorf(_errorMissingField, "NameKey")
	} else if m.Methods.PluginNameKey != m.NameKey {
		return fmt.Errorf(_errorMissingField, "Methods.SourcePlugin")
	}

	// Each configuration key must have a matching entry in Methods.
	for key := range m.Priorities {
		switch key {
		// Lifecycle related methods.
		case protocol.MethodInitialize:
			if m.Methods.Initialize == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodInitialized:
			if m.Methods.Initialized == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodShutdown:
			if m.Methods.Shutdown == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodExit:
			if m.Methods.Exit == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		// Document related methods.
		case protocol.MethodTextDocumentDidChange:
			if m.Methods.DidChange == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodWorkspaceDidChangeWatchedFiles:
			if m.Methods.DidChangeWatchedFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentDidOpen:
			if m.Methods.DidOpen == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentDidClose:
			if m.Methods.DidClose == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentWillSave:
			if m.Methods.WillSave == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentWillSaveWaitUntil:
			if m.Methods.WillSaveWaitUntil == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentDidSave:
			if m.Methods.DidSave == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodWillRenameFiles:
			if m.Methods.WillRenameFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodDidRenameFiles:
			if m.Methods.DidRenameFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodWillCreateFiles:
			if m.Methods.WillCreateFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodDidCreateFiles:
			if m.Methods.DidCreateFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodWillDeleteFiles:
			if m.Methods.WillDeleteFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodDidDeleteFiles:
			if m.Methods.DidDeleteFiles == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentCodeAction:
			if m.Methods.CodeAction == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentCodeLens:
			if m.Methods.CodeLens == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodCodeLensRefresh:
			if m.Methods.CodeLensRefresh == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodCodeLensResolve:
			if m.Methods.CodeLensResolve == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentDeclaration:
			if m.Methods.GotoDeclaration == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentDefinition:
			if m.Methods.GotoDefinition == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentTypeDefinition:
			if m.Methods.GotoTypeDefinition == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentImplementation:
			if m.Methods.GotoImplementation == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentReferences:
			if m.Methods.References == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentHover:
			if m.Methods.Hover == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodTextDocumentDocumentSymbol:
			if m.Methods.DocumentSymbol == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}
		case protocol.MethodWorkspaceExecuteCommand:
			if m.Methods.ExecuteCommand == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}

		case protocol.MethodWorkDoneProgressCancel:
			if m.Methods.WorkDoneProgressCancel == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}

		case MethodEndSession:
			if m.Methods.EndSession == nil {
				return fmt.Errorf(_errorMissingMethod, key)
			}

		default:
			return fmt.Errorf(_errorUnrecognizedMethod, key)
		}
	}
	return nil
}
