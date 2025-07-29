package mapper

import (
	"errors"
	"fmt"

	"go.lsp.dev/protocol"
)

// InitalizeResultAppendServerCapabilitiesWorkspaceFileOperations appends ServerCapabilitiesWorkspaceFileOperations into an existing InitializeResult, adding to existing values if present or initializing an entry if not yet present.
func InitalizeResultAppendServerCapabilitiesWorkspaceFileOperations(initResult *protocol.InitializeResult, workspaceFileOperations *protocol.ServerCapabilitiesWorkspaceFileOperations) error {
	if initResult.Capabilities.Workspace == nil {
		initResult.Capabilities.Workspace = &protocol.ServerCapabilitiesWorkspace{}
	}

	if initResult.Capabilities.Workspace.FileOperations == nil {
		initResult.Capabilities.Workspace.FileOperations = &protocol.ServerCapabilitiesWorkspaceFileOperations{}
	}

	initResult.Capabilities.Workspace.FileOperations = AppendServerCapabilitiesWorkspaceFileOperations(initResult.Capabilities.Workspace.FileOperations, workspaceFileOperations)
	return nil
}

// InitializeResultAppendCodeActionProvider appends a CodeActionProvider into an existing InitializeResult, adding to existing values if present or initializing an entry if not yet present.
func InitializeResultAppendCodeActionProvider(initResult *protocol.InitializeResult, newOptions *protocol.CodeActionOptions) error {
	if initResult.Capabilities.CodeActionProvider == nil {
		initResult.Capabilities.CodeActionProvider = newOptions
		return nil
	}

	currentCodeActionOptions, ok := initResult.Capabilities.CodeActionProvider.(*protocol.CodeActionOptions)
	if !ok {
		return errors.New("CodeActionProvider does not match expected type of *protocol.CodeActionOptions")
	}

	if newOptions.CodeActionKinds != nil {
		if currentCodeActionOptions.CodeActionKinds == nil {
			// If the current CodeActionKinds is nil, just set it to the new value.
			currentCodeActionOptions.CodeActionKinds = newOptions.CodeActionKinds
		} else {
			// Otherwise, add values that are not already present in the current CodeActionKinds.
			seen := map[protocol.CodeActionKind]interface{}{}
			combined := []protocol.CodeActionKind{}
			for _, action := range currentCodeActionOptions.CodeActionKinds {
				seen[action] = struct{}{}
				combined = append(combined, action)
			}
			for _, action := range newOptions.CodeActionKinds {
				if _, ok := seen[action]; !ok {
					combined = append(combined, action)
				}
			}
			currentCodeActionOptions.CodeActionKinds = combined
		}
	}

	if newOptions.ResolveProvider == true {
		currentCodeActionOptions.ResolveProvider = true
	}

	initResult.Capabilities.CodeActionProvider = currentCodeActionOptions
	return nil
}

// InitializeResultEnsureCodeLensProvider ensures that a CodeLensProvider is present in an existing InitializeResult, initializing an entry if not yet present.
// If enableResolveProvider is true in at least one call across all plugins, then the final result for ResolveProvider will be true.
func InitializeResultEnsureCodeLensProvider(initResult *protocol.InitializeResult, enableResolveProvider bool) error {
	if initResult.Capabilities.CodeLensProvider == nil {
		initResult.Capabilities.CodeLensProvider = &protocol.CodeLensOptions{}
	}

	if enableResolveProvider == true {
		initResult.Capabilities.CodeLensProvider.ResolveProvider = true
	}
	return nil
}

// InitializeResultEnsureDefinitionProvider ensures the definition provider capability is set
func InitializeResultEnsureDefinitionProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.DefinitionProvider == nil {
		initResult.Capabilities.DefinitionProvider = &protocol.DefinitionOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.DefinitionProvider.(*protocol.DefinitionOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.DefinitionProvider = &protocol.DefinitionOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultEnsureDeclarationProvider ensures the declaration provider capability is set
func InitializeResultEnsureDeclarationProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.DeclarationProvider == nil {
		initResult.Capabilities.DeclarationProvider = &protocol.DeclarationOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.DeclarationProvider.(*protocol.DeclarationOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.DeclarationProvider = &protocol.DeclarationOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultEnsureImplementationProvider ensures the implementation provider capability is set
func InitializeResultEnsureImplementationProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.ImplementationProvider == nil {
		initResult.Capabilities.ImplementationProvider = &protocol.ImplementationOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.ImplementationProvider.(*protocol.ImplementationOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.ImplementationProvider = &protocol.ImplementationOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultEnsureTypeDefinitionProvider ensures the type definition provider capability is set
func InitializeResultEnsureTypeDefinitionProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.TypeDefinitionProvider == nil {
		initResult.Capabilities.TypeDefinitionProvider = &protocol.TypeDefinitionOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.TypeDefinitionProvider.(*protocol.TypeDefinitionOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.TypeDefinitionProvider = &protocol.TypeDefinitionOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultEnsureReferencesProvider ensures the references provider capability is set
func InitializeResultEnsureReferencesProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.ReferencesProvider == nil {
		initResult.Capabilities.ReferencesProvider = &protocol.ReferencesOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.ReferencesProvider.(*protocol.ReferencesOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.ReferencesProvider = &protocol.ReferencesOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultEnsureHoverProvider ensures the hover provider capability is set
func InitializeResultEnsureHoverProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.HoverProvider == nil {
		initResult.Capabilities.HoverProvider = &protocol.HoverOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.HoverProvider.(*protocol.HoverOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.HoverProvider = &protocol.HoverOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultEnsureDocumentSymbolProvider ensures the document symbol provider capability is set
func InitializeResultEnsureDocumentSymbolProvider(initResult *protocol.InitializeResult, workDoneProgress bool) {
	if initResult == nil {
		return
	}
	if initResult.Capabilities.DocumentSymbolProvider == nil {
		initResult.Capabilities.DocumentSymbolProvider = &protocol.DocumentSymbolOptions{}
	}

	if workDoneProgress == true {
		defOpts, ok := initResult.Capabilities.DocumentSymbolProvider.(*protocol.DocumentSymbolOptions)
		if ok {
			defOpts.WorkDoneProgress = true
		} else {
			initResult.Capabilities.DocumentSymbolProvider = &protocol.DocumentSymbolOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: true,
				},
			}
		}
	}
}

// InitializeResultAppendExecuteCommandProvider appends ExecuteCommandOptions into an existing InitializeResult.
// Commands must be unique across all plugins, and this function will fail if a duplicate is found.
func InitializeResultAppendExecuteCommandProvider(initResult *protocol.InitializeResult, newOptions *protocol.ExecuteCommandOptions) error {
	if initResult.Capabilities.ExecuteCommandProvider == nil {
		initResult.Capabilities.ExecuteCommandProvider = newOptions
		return nil
	}

	if newOptions.Commands == nil {
		return nil
	}

	if initResult.Capabilities.ExecuteCommandProvider.Commands == nil {
		// If the current Commands is nil, just set it to the new value.
		initResult.Capabilities.ExecuteCommandProvider.Commands = newOptions.Commands
	} else {
		// Otherwise, combine with existing Commands and fail on duplicate.
		seen := map[string]interface{}{}
		combined := []string{}
		for _, cmd := range initResult.Capabilities.ExecuteCommandProvider.Commands {
			seen[cmd] = struct{}{}
			combined = append(combined, cmd)
		}
		for _, cmd := range newOptions.Commands {
			if _, ok := seen[cmd]; ok {
				return fmt.Errorf("command %q in ExecuteCommandOptions already exists and cannot be duplicated", cmd)
			}
			combined = append(combined, cmd)
		}
		initResult.Capabilities.ExecuteCommandProvider.Commands = combined
	}

	return nil
}

// AppendServerCapabilitiesWorkspaceFileOperations combines two ServerCapabilitiesWorkspaceFileOperations into a single new result.
func AppendServerCapabilitiesWorkspaceFileOperations(primary *protocol.ServerCapabilitiesWorkspaceFileOperations, secondary *protocol.ServerCapabilitiesWorkspaceFileOperations) *protocol.ServerCapabilitiesWorkspaceFileOperations {
	if primary == nil && secondary == nil {
		return &protocol.ServerCapabilitiesWorkspaceFileOperations{}
	} else if primary != nil && secondary == nil {
		return primary
	} else if primary == nil && secondary != nil {
		return secondary
	}

	result := &protocol.ServerCapabilitiesWorkspaceFileOperations{}

	// If both capabilities are blank, do not initialize a new entry.
	if primary.DidCreate != nil || secondary.DidCreate != nil {
		result.DidCreate = AppendFileOperationRegistrationOptions(primary.DidCreate, secondary.DidCreate)
	}

	if primary.WillCreate != nil || secondary.WillCreate != nil {
		result.WillCreate = AppendFileOperationRegistrationOptions(primary.WillCreate, secondary.WillCreate)
	}

	if primary.DidRename != nil || secondary.DidRename != nil {
		result.DidRename = AppendFileOperationRegistrationOptions(primary.DidRename, secondary.DidRename)
	}

	if primary.WillRename != nil || secondary.WillRename != nil {
		result.WillRename = AppendFileOperationRegistrationOptions(primary.WillRename, secondary.WillRename)
	}

	if primary.DidDelete != nil || secondary.DidDelete != nil {
		result.DidDelete = AppendFileOperationRegistrationOptions(primary.DidDelete, secondary.DidDelete)
	}

	if primary.WillDelete != nil || secondary.WillDelete != nil {
		result.WillDelete = AppendFileOperationRegistrationOptions(primary.WillDelete, secondary.WillDelete)
	}

	return result
}

// AppendFileOperationRegistrationOptions combines two FileOperationRegistrationOptions into a single new result.
func AppendFileOperationRegistrationOptions(primary *protocol.FileOperationRegistrationOptions, secondary *protocol.FileOperationRegistrationOptions) *protocol.FileOperationRegistrationOptions {
	result := &protocol.FileOperationRegistrationOptions{}
	if primary == nil && secondary == nil {
		return result
	} else if primary == nil && secondary != nil {
		result.Filters = secondary.Filters
	} else if primary != nil && secondary == nil {
		result.Filters = primary.Filters
	} else {
		resultFilters := make([]protocol.FileOperationFilter, 0, len(primary.Filters)+len(secondary.Filters))
		seen := make(map[protocol.FileOperationFilter]interface{})
		for _, filter := range primary.Filters {
			seen[filter] = nil
			resultFilters = append(resultFilters, filter)
		}

		for _, filter := range secondary.Filters {
			if _, ok := seen[filter]; !ok {
				resultFilters = append(resultFilters, filter)
			}
		}
		result.Filters = resultFilters
	}
	return result
}
