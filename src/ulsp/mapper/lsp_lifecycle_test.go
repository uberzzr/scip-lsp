package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

func TestInitalizeResultAppendServerCapabilitiesWorkspaceFileOperations(t *testing.T) {
	tests := []struct {
		name         string
		dataToAppend *protocol.ServerCapabilitiesWorkspaceFileOperations
		expected     *protocol.ServerCapabilitiesWorkspaceFileOperations
	}{
		{
			name:         "no additional data",
			dataToAppend: nil,
			// Preserve the original file operations.
			expected: getSampleServerCapabilitiesWorkspaceFileOperation(true, true, false, false, false, false, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
		},
		{
			name:         "additional non-overlapping data",
			dataToAppend: getSampleServerCapabilitiesWorkspaceFileOperation(false, false, true, true, true, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
			// Preserve the original file operations.
			expected: getSampleServerCapabilitiesWorkspaceFileOperation(true, true, true, true, true, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
		},
		{
			name:         "additional overlapping data",
			dataToAppend: getSampleServerCapabilitiesWorkspaceFileOperation(true, true, false, false, false, false, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")}),
			expected:     getSampleServerCapabilitiesWorkspaceFileOperation(true, true, false, false, false, false, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar"), getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := getSampleInitResult()
			InitalizeResultAppendServerCapabilitiesWorkspaceFileOperations(initResult, tt.dataToAppend)
			assert.Equal(t, tt.expected.DidCreate, initResult.Capabilities.Workspace.FileOperations.DidCreate)
			assert.Equal(t, tt.expected.WillCreate, initResult.Capabilities.Workspace.FileOperations.WillCreate)
			assert.Equal(t, tt.expected.DidRename, initResult.Capabilities.Workspace.FileOperations.DidRename)
			assert.Equal(t, tt.expected.WillRename, initResult.Capabilities.Workspace.FileOperations.WillRename)
			assert.Equal(t, tt.expected.DidDelete, initResult.Capabilities.Workspace.FileOperations.DidDelete)
			assert.Equal(t, tt.expected.WillDelete, initResult.Capabilities.Workspace.FileOperations.WillDelete)
		})
	}

}

func TestInitializeResultAppendCodeActionProvider(t *testing.T) {
	tests := []struct {
		name         string
		initial      interface{}
		dataToAppend *protocol.CodeActionOptions
		expected     *protocol.CodeActionOptions
		error        bool
	}{
		{
			name: "added action kinds",
			initial: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.Refactor,
				},
			},
			dataToAppend: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
				ResolveProvider: true,
			},
			expected: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
					protocol.Refactor,
				},
				ResolveProvider: true,
			},
		},
		{
			name: "nil existing data",
			dataToAppend: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
					protocol.Refactor,
				},
			},
			expected: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
					protocol.Refactor,
				},
				ResolveProvider: false,
			},
		},
		{
			name: "overlapping data",
			initial: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.Refactor,
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
			},
			dataToAppend: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
				ResolveProvider: true,
			},
			expected: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
					protocol.Refactor,
				},
				ResolveProvider: true,
			},
		},
		{
			name:    "empty fields",
			initial: &protocol.CodeActionOptions{},
			dataToAppend: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
				ResolveProvider: true,
			},
			expected: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
				ResolveProvider: true,
			},
		},
		{
			name:    "type assertion failure",
			initial: 5,
			dataToAppend: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
				ResolveProvider: true,
			},
			expected: &protocol.CodeActionOptions{
				CodeActionKinds: []protocol.CodeActionKind{
					protocol.SourceOrganizeImports,
					protocol.Source,
				},
				ResolveProvider: true,
			},
			error: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					CodeActionProvider: tt.initial,
				},
			}
			err := InitializeResultAppendCodeActionProvider(initResult, tt.dataToAppend)

			result, ok := initResult.Capabilities.CodeActionProvider.(*protocol.CodeActionOptions)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.True(t, ok)
				assert.Equal(t, tt.expected.ResolveProvider, result.ResolveProvider)
				assert.ElementsMatch(t, tt.expected.CodeActionKinds, result.CodeActionKinds)
			}
		})
	}
}

func TestInitializeResultAppendExecuteCommandProvider(t *testing.T) {
	tests := []struct {
		name         string
		initial      *protocol.ExecuteCommandOptions
		dataToAppend *protocol.ExecuteCommandOptions
		expected     *protocol.ExecuteCommandOptions
		error        bool
	}{
		{
			name: "existing commands",
			initial: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
			dataToAppend: &protocol.ExecuteCommandOptions{
				Commands: []string{"command3"},
			},
			expected: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2", "command3"},
			},
		},
		{
			name: "overlapping commands",
			initial: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
			dataToAppend: &protocol.ExecuteCommandOptions{
				Commands: []string{"command2"},
			},
			error: true,
		},
		{
			name: "no new commands",
			initial: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
			dataToAppend: &protocol.ExecuteCommandOptions{},
			expected: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
		},
		{
			name: "no existing options",
			dataToAppend: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
			expected: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
		},
		{
			name:    "no existing commands",
			initial: &protocol.ExecuteCommandOptions{},
			dataToAppend: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
			expected: &protocol.ExecuteCommandOptions{
				Commands: []string{"command1", "command2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					ExecuteCommandProvider: tt.initial,
				},
			}
			err := InitializeResultAppendExecuteCommandProvider(initResult, tt.dataToAppend)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.ElementsMatch(t, tt.expected.Commands, initResult.Capabilities.ExecuteCommandProvider.Commands)
			}
		})
	}
}

func TestInitializeResultEnsureCodeLensProvider(t *testing.T) {
	tests := []struct {
		name                  string
		initial               *protocol.CodeLensOptions
		enableResolveProvider bool
		expected              *protocol.CodeLensOptions
	}{
		{
			name: "existing options",
			initial: &protocol.CodeLensOptions{
				ResolveProvider: false,
			},
			enableResolveProvider: true,
			expected: &protocol.CodeLensOptions{
				ResolveProvider: true,
			},
		},
		{
			name:                  "no existing options",
			initial:               nil,
			enableResolveProvider: true,
			expected: &protocol.CodeLensOptions{
				ResolveProvider: true,
			},
		},
		{
			name: "no resolve provider",
			initial: &protocol.CodeLensOptions{
				ResolveProvider: false,
			},
			enableResolveProvider: false,
			expected: &protocol.CodeLensOptions{
				ResolveProvider: false,
			},
		},
		{
			name: "resolve provider already enabled",
			initial: &protocol.CodeLensOptions{
				ResolveProvider: true,
			},
			enableResolveProvider: false,
			expected: &protocol.CodeLensOptions{
				ResolveProvider: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					CodeLensProvider: tt.initial,
				},
			}
			InitializeResultEnsureCodeLensProvider(initResult, tt.enableResolveProvider)
			assert.Equal(t, tt.expected.ResolveProvider, initResult.Capabilities.CodeLensProvider.ResolveProvider)
		})
	}
}

func TestAppendServerCapabilitiesWorkspaceFileOperations(t *testing.T) {
	tests := []struct {
		name      string
		primary   *protocol.ServerCapabilitiesWorkspaceFileOperations
		secondary *protocol.ServerCapabilitiesWorkspaceFileOperations
		expected  *protocol.ServerCapabilitiesWorkspaceFileOperations
	}{
		{
			name:     "primary only",
			primary:  getSampleServerCapabilitiesWorkspaceFileOperation(true, true, true, true, true, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
			expected: getSampleServerCapabilitiesWorkspaceFileOperation(true, true, true, true, true, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
		},
		{
			name:      "secondary only",
			secondary: getSampleServerCapabilitiesWorkspaceFileOperation(true, true, true, true, true, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
			expected:  getSampleServerCapabilitiesWorkspaceFileOperation(true, true, true, true, true, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
		},
		{
			name:     "both nil",
			expected: &protocol.ServerCapabilitiesWorkspaceFileOperations{},
		},
		{
			name:      "distinct methods",
			primary:   getSampleServerCapabilitiesWorkspaceFileOperation(false, true, false, true, false, true, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
			secondary: getSampleServerCapabilitiesWorkspaceFileOperation(true, false, true, false, true, false, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")}),
			expected: &protocol.ServerCapabilitiesWorkspaceFileOperations{
				DidCreate: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")},
				},
				WillCreate: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
				},
				DidRename: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")},
				},
				WillRename: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
				},
				DidDelete: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")},
				},
				WillDelete: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
				},
			},
		},
		{
			name:      "overlapping methods",
			primary:   getSampleServerCapabilitiesWorkspaceFileOperation(false, true, false, true, false, false, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def")}),
			secondary: getSampleServerCapabilitiesWorkspaceFileOperation(false, true, false, true, false, false, []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")}),
			expected: &protocol.ServerCapabilitiesWorkspaceFileOperations{
				WillCreate: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def"), getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
				},
				WillRename: &protocol.FileOperationRegistrationOptions{
					Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.abc"), getExampleFileOperationFilter("file", "**/*.def"), getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AppendServerCapabilitiesWorkspaceFileOperations(tt.primary, tt.secondary)
			assert.Equal(t, tt.expected.DidCreate, result.DidCreate)
			assert.Equal(t, tt.expected.WillCreate, result.WillCreate)
			assert.Equal(t, tt.expected.DidRename, result.DidRename)
			assert.Equal(t, tt.expected.WillRename, result.WillRename)
			assert.Equal(t, tt.expected.DidDelete, result.DidDelete)
			assert.Equal(t, tt.expected.WillDelete, result.WillDelete)
		})
	}
}

func TestAppendFileOperationRegistrationOptions(t *testing.T) {
	tests := []struct {
		name      string
		primary   *protocol.FileOperationRegistrationOptions
		secondary *protocol.FileOperationRegistrationOptions
		expected  *protocol.FileOperationRegistrationOptions
	}{
		{
			name: "primary only",
			primary: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
			},
			expected: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
			},
		},
		{
			name: "secondary only",
			secondary: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
			},
			expected: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
			},
		},
		{
			name: "no overlap",
			primary: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{
					getExampleFileOperationFilter("file", "**/*.abc"),
					getExampleFileOperationFilter("file", "**/*.def"),
					getExampleFileOperationFilter("file", "**/*.ghi"),
				},
			},
			secondary: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{getExampleFileOperationFilter("file", "**/*.foo"), getExampleFileOperationFilter("file", "**/*.bar")},
			},
			expected: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{
					getExampleFileOperationFilter("file", "**/*.abc"),
					getExampleFileOperationFilter("file", "**/*.def"),
					getExampleFileOperationFilter("file", "**/*.ghi"),
					getExampleFileOperationFilter("file", "**/*.foo"),
					getExampleFileOperationFilter("file", "**/*.bar"),
				},
			},
		},
		{
			name: "overlapping values",
			primary: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{
					getExampleFileOperationFilter("file", "**/*.abc"),
					getExampleFileOperationFilter("file", "**/*.def"),
					getExampleFileOperationFilter("file", "**/*.ghi"),
				},
			},
			secondary: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{
					getExampleFileOperationFilter("file", "**/*.foo"),
					getExampleFileOperationFilter("file", "**/*.bar"),
					getExampleFileOperationFilter("file", "**/*.def"),
					getExampleFileOperationFilter("file", "**/*.ghi"),
				},
			},
			expected: &protocol.FileOperationRegistrationOptions{
				Filters: []protocol.FileOperationFilter{
					getExampleFileOperationFilter("file", "**/*.abc"),
					getExampleFileOperationFilter("file", "**/*.def"),
					getExampleFileOperationFilter("file", "**/*.ghi"),
					getExampleFileOperationFilter("file", "**/*.foo"),
					getExampleFileOperationFilter("file", "**/*.bar"),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := AppendFileOperationRegistrationOptions(test.primary, test.secondary)
			assert.ElementsMatch(t, test.expected.Filters, result.Filters)
		})
	}
}

func TestInitializeResultEnsureDefinitionProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.DefinitionOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.DefinitionOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					DefinitionProvider: tt.initial,
				},
			}
			InitializeResultEnsureDefinitionProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.DefinitionOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.DefinitionProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureDefinitionProvider(nil, false) })
	})
}

func TestInitializeResultEnsureTypeDefinitionProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.TypeDefinitionOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.TypeDefinitionOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					TypeDefinitionProvider: tt.initial,
				},
			}
			InitializeResultEnsureTypeDefinitionProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.TypeDefinitionOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.TypeDefinitionProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureTypeDefinitionProvider(nil, false) })
	})
}

func TestInitializeResultEnsureDeclarationProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.DeclarationOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.DeclarationOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					DeclarationProvider: tt.initial,
				},
			}
			InitializeResultEnsureDeclarationProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.DeclarationOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.DeclarationProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureDeclarationProvider(nil, false) })
	})
}

func TestInitializeResultEnsureReferencesProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.ReferencesOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.ReferencesOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					ReferencesProvider: tt.initial,
				},
			}
			InitializeResultEnsureReferencesProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.ReferencesOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.ReferencesProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureReferencesProvider(nil, false) })
	})
}

func TestInitializeResultEnsureImplementationProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.ImplementationOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.ImplementationOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					ImplementationProvider: tt.initial,
				},
			}
			InitializeResultEnsureImplementationProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.ImplementationOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.ImplementationProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureImplementationProvider(nil, false) })
	})
}

func TestInitializeResultEnsureDocumentSymbolProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.DocumentSymbolOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.DocumentSymbolOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					DocumentSymbolProvider: tt.initial,
				},
			}
			InitializeResultEnsureDocumentSymbolProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.DocumentSymbolOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.DocumentSymbolProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureDocumentSymbolProvider(nil, false) })
	})

}

func TestInitializeResultEnsureHoverProvider(t *testing.T) {
	tests := []struct {
		name     string
		initial  interface{}
		workDone bool
		expected bool
	}{
		{
			name:     "existing options",
			initial:  &protocol.HoverOptions{},
			expected: false,
		},
		{
			name: "existing workDone",
			initial: &protocol.HoverOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{
					WorkDoneProgress: false,
				},
			},
			workDone: true,
			expected: true,
		},
		{
			name:     "no existing options",
			initial:  nil,
			expected: false,
		},
		{
			name:     "type assertion failure",
			initial:  5,
			expected: true,
			workDone: true,
		},
		{
			name:     "workDone enabled",
			initial:  nil,
			expected: true,
			workDone: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initResult := &protocol.InitializeResult{
				Capabilities: protocol.ServerCapabilities{
					HoverProvider: tt.initial,
				},
			}
			InitializeResultEnsureHoverProvider(initResult, tt.workDone)
			assert.Equal(t, &protocol.HoverOptions{
				WorkDoneProgressOptions: protocol.WorkDoneProgressOptions{WorkDoneProgress: tt.expected},
			}, initResult.Capabilities.HoverProvider)
		})
	}

	t.Run("no init", func(t *testing.T) {
		assert.NotPanics(t, func() { InitializeResultEnsureHoverProvider(nil, false) })
	})
}

func getExampleFileOperationFilter(scheme, glob string) protocol.FileOperationFilter {
	return protocol.FileOperationFilter{
		Scheme: scheme,
		Pattern: protocol.FileOperationPattern{
			Glob: glob,
		},
	}
}

func getSampleServerCapabilitiesWorkspaceFileOperation(didCreate, willCreate, didRename, willRename, didDelete, willDelete bool, filters []protocol.FileOperationFilter) *protocol.ServerCapabilitiesWorkspaceFileOperations {
	result := &protocol.ServerCapabilitiesWorkspaceFileOperations{}

	if didCreate {
		result.DidCreate = &protocol.FileOperationRegistrationOptions{
			Filters: filters,
		}
	}

	if willCreate {
		result.WillCreate = &protocol.FileOperationRegistrationOptions{
			Filters: filters,
		}
	}

	if didRename {
		result.DidRename = &protocol.FileOperationRegistrationOptions{
			Filters: filters,
		}
	}

	if willRename {
		result.WillRename = &protocol.FileOperationRegistrationOptions{
			Filters: filters,
		}
	}

	if didDelete {
		result.DidDelete = &protocol.FileOperationRegistrationOptions{
			Filters: filters,
		}
	}

	if willDelete {
		result.WillDelete = &protocol.FileOperationRegistrationOptions{
			Filters: filters,
		}
	}

	return result
}

func getSampleInitResult() *protocol.InitializeResult {
	return &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			Workspace: &protocol.ServerCapabilitiesWorkspace{
				FileOperations: &protocol.ServerCapabilitiesWorkspaceFileOperations{
					DidCreate: &protocol.FileOperationRegistrationOptions{
						Filters: []protocol.FileOperationFilter{
							{
								Scheme: "file",
								Pattern: protocol.FileOperationPattern{
									Glob: "**/*.foo",
								},
							},
							{
								Scheme: "file",
								Pattern: protocol.FileOperationPattern{
									Glob: "**/*.bar",
								},
							},
						},
					},
					WillCreate: &protocol.FileOperationRegistrationOptions{
						Filters: []protocol.FileOperationFilter{
							{
								Scheme: "file",
								Pattern: protocol.FileOperationPattern{
									Glob: "**/*.foo",
								},
							},
							{
								Scheme: "file",
								Pattern: protocol.FileOperationPattern{
									Glob: "**/*.bar",
								},
							},
						},
					},
				},
			},
		},
	}
}
