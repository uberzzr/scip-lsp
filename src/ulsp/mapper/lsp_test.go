package mapper

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"go.lsp.dev/protocol"
)

func TestRequestToInitalizeParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.InitializeParams{
			Locale:    "exampleLocale",
			ProcessID: 5555,
		}
		validReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, params)
		result, err := RequestToInitializeParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.Locale, result.Locale)
		assert.Equal(t, params.ProcessID, result.ProcessID)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest("sampleMethodName", struct {
			Locale int
		}{
			Locale: 5,
		})
		_, err := RequestToInitializeParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToInitalizedParams(t *testing.T) {
	params := protocol.InitializedParams{}
	validReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, params)
	_, err := RequestToInitializedParams(validReq)
	assert.NoError(t, err)
}

func TestRequestToCreateFilesParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.CreateFilesParams{
			Files: []protocol.FileCreate{
				{
					URI: "file://path/file1.go",
				},
				{
					URI: "file://path/file2.go",
				},
			},
		}
		validReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, params)
		result, err := RequestToCreateFilesParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, len(params.Files), len(result.Files))
		for i := range result.Files {
			assert.Equal(t, params.Files[i], result.Files[i])
		}
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest("sampleMethodName", struct {
			Files int
		}{
			Files: 5,
		})
		_, err := RequestToCreateFilesParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToRenameFilesParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.RenameFilesParams{
			Files: []protocol.FileRename{
				{
					OldURI: "file://path/file1.go",
					NewURI: "file://path/file1_renamed.go",
				},
				{
					OldURI: "file://path/file2.go",
					NewURI: "file://path/file2_renamed.go",
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToRenameFilesParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, len(params.Files), len(result.Files))
		for i := range result.Files {
			assert.Equal(t, params.Files[i].NewURI, result.Files[i].NewURI)
			assert.Equal(t, params.Files[i].OldURI, result.Files[i].OldURI)
		}
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			Files int
		}{
			Files: 5,
		})
		_, err := RequestToRenameFilesParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDeleteFilesParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DeleteFilesParams{
			Files: []protocol.FileDelete{
				{
					URI: "file://path/file1.go",
				},
				{
					URI: "file://path/file2.go",
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDeleteFilesParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, len(params.Files), len(result.Files))
		for i := range result.Files {
			assert.Equal(t, params.Files[i], result.Files[i])
		}
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			Files int
		}{
			Files: 5,
		})
		_, err := RequestToDeleteFilesParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDidChangeTextDocumentParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DidChangeTextDocumentParams{
			TextDocument: protocol.VersionedTextDocumentIdentifier{
				TextDocumentIdentifier: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Version: 100,
			},
			ContentChanges: []protocol.TextDocumentContentChangeEvent{
				{
					Text: "sample1",
				},
				{
					Text: "sample2",
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDidChangeTextDocumentParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument.URI, result.TextDocument.URI)
		assert.Equal(t, len(params.ContentChanges), len(result.ContentChanges))
		for i := range result.ContentChanges {
			assert.Equal(t, params.ContentChanges[i].Text, result.ContentChanges[i].Text)
		}
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDidChangeTextDocumentParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDidCloseTextDocumentParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DidCloseTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file://path/file1.go",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDidCloseTextDocumentParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument.URI, result.TextDocument.URI)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDidCloseTextDocumentParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDidOpenTextDocumentParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DidOpenTextDocumentParams{
			TextDocument: protocol.TextDocumentItem{
				URI: "file://path/file1.go",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDidOpenTextDocumentParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument.URI, result.TextDocument.URI)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDidOpenTextDocumentParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDidSaveTextDocumentParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DidSaveTextDocumentParams{
			Text: "sampleText",
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file://path/file1.go",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDidSaveTextDocumentParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument.URI, result.TextDocument.URI)
		assert.Equal(t, params.Text, result.Text)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDidSaveTextDocumentParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToWillSaveTextDocumentParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.WillSaveTextDocumentParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file://path/file1.go",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToWillSaveTextDocumentParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument.URI, result.TextDocument.URI)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToWillSaveTextDocumentParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDidChangeWatchedFilesParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DidChangeWatchedFilesParams{
			Changes: []*protocol.FileEvent{
				{
					URI: "file://path/file1.go",
				},
				{
					URI: "file://path/file2.go",
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDidChangeWatchedFilesParams(validReq)
		assert.NoError(t, err)
		for i := range result.Changes {
			assert.Equal(t, params.Changes[i].URI, result.Changes[i].URI)
		}
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			Changes int
		}{
			Changes: 5,
		})
		_, err := RequestToDidChangeWatchedFilesParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToCodeActionParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.CodeActionParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file://path/file1.go",
			},
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      1,
					Character: 1,
				},
				End: protocol.Position{
					Line:      2,
					Character: 2,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToCodeActionParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Range, result.Range)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToCodeActionParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToCodeLensParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.CodeLensParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file://path/file1.go",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToCodeLensParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToCodeLensParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToCodeLens(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.CodeLens{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      1,
					Character: 1,
				},
				End: protocol.Position{
					Line:      2,
					Character: 2,
				},
			},
			Command: &protocol.Command{
				Command: "sampleCommand",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToCodeLens(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.Range, result.Range)
		assert.Equal(t, *params.Command, *result.Command)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			Range int
		}{
			Range: 5,
		})
		_, err := RequestToCodeLens(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDeclaration(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DeclarationParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Position: protocol.Position{
					Line:      3,
					Character: 14,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDeclarationParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Position, result.Position)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDeclarationParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDefinition(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Position: protocol.Position{
					Line:      3,
					Character: 14,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDefinitionParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Position, result.Position)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDefinitionParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToExecuteCommandParams(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.ExecuteCommandParams{
			Command: "sample",
			Arguments: []interface{}{
				struct {
					SampleArg1 string `json:"sampleArg1"`
				}{SampleArg1: "sampleVal1"},
				struct {
					SampleArg2 string `json:"sampleArg2"`
				}{SampleArg2: "sampleVal2"},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToExecuteCommandParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.Command, result.Command)
		for i := range result.Arguments {
			rawArg, _ := json.Marshal(params.Arguments[i])
			assert.Equal(t, rawArg, result.Arguments[i])
		}
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			Command int
		}{
			Command: 5,
		})
		_, err := RequestToExecuteCommandParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToWorkDoneProgressCancelParams(t *testing.T) {
	//  Below code should have tested the function but due to bug
	// https://github.com/go-language-server/protocol/issues/30
	// we cannot test this function

	// t.Run("valid params", func(t *testing.T) {
	// 	sampleToken := protocol.NewProgressToken("Sample-token")
	// 	params := protocol.WorkDoneProgressCancelParams{
	// 		Token: *sampleToken,
	// 	}
	// 	validReq := factory.JSONRPCRequest("sampleMethodName", params)
	// 	result, err := RequestToWorkDoneProgressCancelParams(validReq)
	// 	assert.NoError(t, err)
	// 	assert.Equal(t, params.Token, result.Token)

	// })
	// t.Run("invalid params", func(t *testing.T) {
	// 	invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
	// 		TextDocument int
	// 	}{
	// 		TextDocument: 5,
	// 	})
	// 	_, err := RequestToWorkDoneProgressCancelParams(invalidReq)
	// 	assert.Error(t, err)
	// })
}

func TestRequestToTypeDefinition(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.TypeDefinitionParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Position: protocol.Position{
					Line:      3,
					Character: 14,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToTypeDefinitionParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Position, result.Position)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToTypeDefinitionParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToImplementation(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.ImplementationParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Position: protocol.Position{
					Line:      3,
					Character: 14,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToImplementationParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Position, result.Position)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToImplementationParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToReferences(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.ReferenceParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Position: protocol.Position{
					Line:      3,
					Character: 14,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToReferencesParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Position, result.Position)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToReferencesParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToDocumentSymbol(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.DocumentSymbolParams{
			TextDocument: protocol.TextDocumentIdentifier{
				URI: "file://path/file1.go",
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToDocumentSymbolParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToDocumentSymbolParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestRequestToHover(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		params := protocol.HoverParams{
			TextDocumentPositionParams: protocol.TextDocumentPositionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					URI: "file://path/file1.go",
				},
				Position: protocol.Position{
					Line:      3,
					Character: 14,
				},
			},
		}
		validReq := factory.JSONRPCRequest("sampleMethodName", params)
		result, err := RequestToHoverParams(validReq)
		assert.NoError(t, err)
		assert.Equal(t, params.TextDocument, result.TextDocument)
		assert.Equal(t, params.Position, result.Position)
	})

	t.Run("invalid params", func(t *testing.T) {
		invalidReq := factory.JSONRPCRequest(protocol.MethodDidCreateFiles, struct {
			TextDocument int
		}{
			TextDocument: 5,
		})
		_, err := RequestToHoverParams(invalidReq)
		assert.Error(t, err)
	})
}

func TestNewCodeLens(t *testing.T) {
	t.Run("valid params", func(t *testing.T) {
		sampleTitle := "Sample Title"
		sampleRange := protocol.Range{
			Start: protocol.Position{
				Line:      1,
				Character: 1,
			},
			End: protocol.Position{
				Line:      2,
				Character: 2,
			},
		}
		sampleCommand := "sampleCommand"
		sampleArgs := "sample args"
		result := NewCodeLens(sampleRange, sampleTitle, sampleCommand, sampleArgs)
		assert.Equal(t, sampleRange, result.Range)
		assert.Equal(t, sampleTitle, result.Command.Title)
		assert.Equal(t, sampleCommand, result.Command.Command)
		assert.Equal(t, sampleArgs, result.Command.Arguments[0])
	})
}

func TestNewCodeAction(t *testing.T) {
	sampleTitle := "Sample Title"
	sampleCommand := "sampleCommand"
	sampleArgs := "sample args"
	result := NewCodeAction(sampleTitle, sampleCommand, protocol.Refactor, sampleArgs)
	assert.Equal(t, sampleTitle, result.Title)
	assert.Equal(t, sampleCommand, result.Command.Command)
	assert.Equal(t, sampleArgs, result.Command.Arguments[0])
}

func TestNewCodeActionWithRange(t *testing.T) {
	sampleTitle := "Sample Title"
	sampleRange := protocol.Range{
		Start: protocol.Position{
			Line:      1,
			Character: 1,
		},
		End: protocol.Position{
			Line:      2,
			Character: 2,
		},
	}
	sampleCommand := "sampleCommand"
	sampleArgs := "sample args"
	result := NewCodeActionWithRange(sampleRange, sampleTitle, sampleCommand, protocol.Refactor, sampleArgs)
	assert.Equal(t, sampleRange, result.Range)
	assert.Equal(t, sampleTitle, result.CodeAction.Title)
	assert.Equal(t, sampleCommand, result.CodeAction.Command.Command)
	assert.Equal(t, sampleTitle, result.CodeAction.Command.Title)
	assert.Equal(t, sampleArgs, result.CodeAction.Command.Arguments[0])
}

func TestApplyContentChanges(t *testing.T) {

	tests := []struct {
		name        string
		initialText string
		changes     []protocol.TextDocumentContentChangeEvent
		expected    string
		wantErr     bool
	}{
		{
			name:        "added at end",
			initialText: "sample\ncontent\n",
			changes: []protocol.TextDocumentContentChangeEvent{
				{
					Range: &protocol.Range{
						Start: protocol.Position{
							Line:      3,
							Character: 0,
						},
						End: protocol.Position{
							Line:      3,
							Character: 0,
						},
					},
					Text: "addedText",
				},
			},
			expected: "sample\ncontent\naddedText",
			wantErr:  false,
		},
		{
			name:        "added at beginning",
			initialText: "sample\ncontent\n",
			changes: []protocol.TextDocumentContentChangeEvent{
				{
					Range: &protocol.Range{
						Start: protocol.Position{
							Line:      0,
							Character: 0,
						},
						End: protocol.Position{
							Line:      0,
							Character: 0,
						},
					},
					Text: "addedText\n",
				},
			},
			expected: "addedText\nsample\ncontent\n",
			wantErr:  false,
		},
		{
			name:        "replace full line",
			initialText: "sample\ncontent\n",
			changes: []protocol.TextDocumentContentChangeEvent{
				{
					Range: &protocol.Range{
						Start: protocol.Position{
							Line:      0,
							Character: 0,
						},
						End: protocol.Position{
							Line:      1,
							Character: 0,
						},
					},
					Text: "addedText\n",
				},
			},
			expected: "addedText\ncontent\n",
			wantErr:  false,
		},
		{
			name:        "replace mid line",
			initialText: "sample\ncontent\n",
			changes: []protocol.TextDocumentContentChangeEvent{
				{
					Range: &protocol.Range{
						Start: protocol.Position{
							Line:      1,
							Character: 3,
						},
						End: protocol.Position{
							Line:      1,
							Character: 5,
						},
					},
					Text: "--",
				},
			},
			expected: "sample\ncon--nt\n",
			wantErr:  false,
		},
		{
			name:        "bad start",
			initialText: "sample\ncontent\n",
			changes: []protocol.TextDocumentContentChangeEvent{
				{
					Range: &protocol.Range{
						Start: protocol.Position{
							Line:      10,
							Character: 5,
						},
						End: protocol.Position{
							Line:      1,
							Character: 5,
						},
					},
					Text: "--",
				},
			},
			wantErr: true,
		},
		{
			name:        "bad end",
			initialText: "sample\ncontent\n",
			changes: []protocol.TextDocumentContentChangeEvent{
				{
					Range: &protocol.Range{
						Start: protocol.Position{
							Line:      1,
							Character: 5,
						},
						End: protocol.Position{
							Line:      10,
							Character: 5,
						},
					},
					Text: "--",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ApplyContentChanges(tt.initialText, tt.changes)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, result)
				assert.NoError(t, err)
			}

		})
	}

}

var textEditTestCases = []struct {
	name              string
	initialText       string
	updatedText       string
	editOffsets       []EditOffset
	expectedTextEdits []protocol.TextEdit
}{
	{
		name:        "added at end",
		initialText: "sample\ncontent\n",
		updatedText: "sample\ncontent\naddedText",
		editOffsets: []EditOffset{
			{
				start: 15,
				end:   15,
				text:  "addedText",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      2,
						Character: 0,
					},
					End: protocol.Position{
						Line:      2,
						Character: 0,
					},
				},
				NewText: "addedText",
			},
		},
	},
	{
		name:        "added at beginning",
		initialText: "sample\ncontent\n",
		updatedText: "addedTextsample\ncontent\n",
		editOffsets: []EditOffset{
			{
				start: 0,
				end:   0,
				text:  "addedText",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 0,
					},
					End: protocol.Position{
						Line:      0,
						Character: 0,
					},
				},
				NewText: "addedText",
			},
		},
	},
	{
		name:        "added at middle",
		initialText: "sample\ncontent\n",
		updatedText: "sample\naddedText\ncontent\n",
		editOffsets: []EditOffset{
			{
				start: 7,
				end:   7,
				text:  "addedText\n",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 0,
					},
					End: protocol.Position{
						Line:      1,
						Character: 0,
					},
				},
				NewText: "addedText\n",
			},
		},
	},
	{
		name:        "replace at middle",
		initialText: "sample\nco-ent\n",
		updatedText: "sample\naddedText\ncontent\n",
		editOffsets: []EditOffset{
			{
				start: 7,
				end:   7,
				text:  "addedText\n",
			},
			{
				start: 9,
				end:   10,
				text:  "",
			},
			{
				start: 10,
				end:   10,
				text:  "nt",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 0,
					},
					End: protocol.Position{
						Line:      1,
						Character: 0,
					},
				},
				NewText: "addedText\n",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 2,
					},
					End: protocol.Position{
						Line:      1,
						Character: 3,
					},
				},
				NewText: "",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 3,
					},
					End: protocol.Position{
						Line:      1,
						Character: 3,
					},
				},
				NewText: "nt",
			},
		},
	},
	{
		name:        "delete at middle",
		initialText: "sample\ncontent\n",
		updatedText: "samplecontent\n",
		editOffsets: []EditOffset{
			{
				start: 6,
				end:   7,
				text:  "",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 6,
					},
					End: protocol.Position{
						Line:      1,
						Character: 0,
					},
				},
				NewText: "",
			},
		},
	},
	{
		name:        "delete at beginning",
		initialText: "sample\ncontent\n",
		updatedText: "\ncontent\n",
		editOffsets: []EditOffset{
			{
				start: 0,
				end:   6,
				text:  "",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      0,
						Character: 0,
					},
					End: protocol.Position{
						Line:      0,
						Character: 6,
					},
				},
				NewText: "",
			},
		},
	},
	{
		name:        "delete at end",
		initialText: "sample\ncontent\n",
		updatedText: "sample\ncontent",
		editOffsets: []EditOffset{
			{
				start: 14,
				end:   15,
				text:  "",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      1,
						Character: 7,
					},
					End: protocol.Position{
						Line:      2,
						Character: 0,
					},
				},
				NewText: "",
			},
		},
	},
	{
		name: "code snippet with several edits",
		initialText: `
			package main
			import (
				"fmt"
			)

			func main() {
				fmt.Println("Hello, world")

			}
		`,
		updatedText: `
			package main

			import "fmt"

			func main() {
				fmt.Println("Hello, sample")
			}
		`,
		editOffsets: []EditOffset{
			{
				start: 17,
				end:   17,
				text:  "\n",
			},
			{
				start: 27,
				end:   33,
				text:  "",
			},
			{
				start: 38,
				end:   43,
				text:  "",
			},
			{
				start: 86,
				end:   89,
				text:  "",
			},
			{
				start: 89,
				end:   89,
				text:  "samp",
			},
			{
				start: 90,
				end:   91,
				text:  "",
			},
			{
				start: 91,
				end:   91,
				text:  "e",
			},
			{
				start: 93,
				end:   94,
				text:  "",
			},
		},
		expectedTextEdits: []protocol.TextEdit{
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      2,
						Character: 0,
					},
					End: protocol.Position{
						Line:      2,
						Character: 0,
					},
				},
				NewText: "\n",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      2,
						Character: 10,
					},
					End: protocol.Position{
						Line:      3,
						Character: 4,
					},
				},
				NewText: "",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      3,
						Character: 9,
					},
					End: protocol.Position{
						Line:      4,
						Character: 4,
					},
				},
				NewText: "",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      7,
						Character: 24,
					},
					End: protocol.Position{
						Line:      7,
						Character: 27,
					},
				},
				NewText: "",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      7,
						Character: 27,
					},
					End: protocol.Position{
						Line:      7,
						Character: 27,
					},
				},
				NewText: "samp",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      7,
						Character: 28,
					},
					End: protocol.Position{
						Line:      7,
						Character: 29,
					},
				},
				NewText: "",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      7,
						Character: 29,
					},
					End: protocol.Position{
						Line:      7,
						Character: 29,
					},
				},
				NewText: "e",
			},
			{
				Range: protocol.Range{
					Start: protocol.Position{
						Line:      7,
						Character: 31,
					},
					End: protocol.Position{
						Line:      8,
						Character: 0,
					},
				},
				NewText: "",
			},
		},
	},
}

func TestDiffsToEditOffsets(t *testing.T) {
	for _, tt := range textEditTestCases {
		t.Run(tt.name, func(t *testing.T) {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(tt.initialText, tt.updatedText, false)
			initialText, editOffsets := DiffsToEditOffsets(diffs)
			assert.Equal(t, tt.initialText, initialText.String())

			editOffsetIndex := 0
			offset := 0
			for i := range diffs {
				start := offset
				switch diffs[i].Type {
				case diffmatchpatch.DiffEqual:
					offset += len(diffs[i].Text)
				case diffmatchpatch.DiffDelete:
					offset += len(diffs[i].Text)
					assert.Equal(t, "", editOffsets[editOffsetIndex].text)
					assert.Equal(t, start, editOffsets[editOffsetIndex].start)
					assert.Equal(t, offset, editOffsets[editOffsetIndex].end)
					editOffsetIndex++
				case diffmatchpatch.DiffInsert:
					assert.Equal(t, diffs[i].Text, editOffsets[editOffsetIndex].text)
					assert.Equal(t, start, editOffsets[editOffsetIndex].start)
					assert.Equal(t, start, editOffsets[editOffsetIndex].end)
					editOffsetIndex++
				}
			}

			assert.Equal(t, len(tt.editOffsets), len(editOffsets))
			for i := range editOffsets {
				assert.Equal(t, tt.editOffsets[i], editOffsets[i])
			}
		})
	}
}

func TestEditOffsetsToTextEdits(t *testing.T) {
	for _, tt := range textEditTestCases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EditOffsetsToTextEdits(*bytes.NewBuffer([]byte(tt.initialText)), tt.editOffsets)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTextEdits, result)
		})
	}

	t.Run("invalid range", func(t *testing.T) {
		_, err := EditOffsetsToTextEdits(*bytes.NewBuffer([]byte("sample\ncontent\n")), []EditOffset{
			{
				start: 3,
				end:   -1,
				text:  "addedText",
			},
		})
		assert.Error(t, err)
	})

	t.Run("add beyond end", func(t *testing.T) {
		_, err := EditOffsetsToTextEdits(*bytes.NewBuffer([]byte("sample\ncontent\n")), []EditOffset{
			{
				start: 30,
				end:   30,
				text:  "addedText",
			},
		})
		assert.Error(t, err)
	})
}

func TestDiffsToTextEdits(t *testing.T) {
	for _, tt := range textEditTestCases {
		t.Run(tt.name, func(t *testing.T) {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(tt.initialText, tt.updatedText, false)
			result, err := DiffsToTextEdits(diffs)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTextEdits, result)
		})
	}
}

func TestPositionsToRange(t *testing.T) {
	start := protocol.Position{
		Line:      1,
		Character: 2,
	}
	end := protocol.Position{
		Line:      3,
		Character: 4,
	}
	result := PositionsToRange(start, end)
	assert.Equal(t, start, result.Start)
	assert.Equal(t, end, result.End)
}
