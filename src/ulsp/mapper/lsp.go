package mapper

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sergi/go-diff/diffmatchpatch"
	protocolmapper "github.com/uber/scip-lsp/src/ulsp/internal/protocol"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// EditOffset stores a string modification based on character offset in the string.
type EditOffset struct {
	start int
	end   int
	text  string
}

// RequestToInitializeParams maps the parameters from a jsconrpc2.Request into protocol.InitializeParams.
func RequestToInitializeParams(req jsonrpc2.Request) (*protocol.InitializeParams, error) {
	params := protocol.InitializeParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToInitializedParams maps the parameters from a jsconrpc2.Request into protocol.InitializedParams.
func RequestToInitializedParams(req jsonrpc2.Request) (*protocol.InitializedParams, error) {
	params := protocol.InitializedParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToCreateFilesParams maps the parameters from a jsconrpc2.Request into protocol.CreateFilesParams.
func RequestToCreateFilesParams(req jsonrpc2.Request) (*protocol.CreateFilesParams, error) {
	params := protocol.CreateFilesParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToRenameFilesParams maps the parameters from a jsconrpc2.Request into protocol.RenameFilesParams.
func RequestToRenameFilesParams(req jsonrpc2.Request) (*protocol.RenameFilesParams, error) {
	params := protocol.RenameFilesParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDeleteFilesParams maps the parameters from a jsconrpc2.Request into protocol.DeleteFilesParams.
func RequestToDeleteFilesParams(req jsonrpc2.Request) (*protocol.DeleteFilesParams, error) {
	params := protocol.DeleteFilesParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDidChangeTextDocumentParams maps the parameters from a jsconrpc2.Request into protocol.DidChangeTextDocumentParams.
func RequestToDidChangeTextDocumentParams(req jsonrpc2.Request) (*protocol.DidChangeTextDocumentParams, error) {
	params := protocol.DidChangeTextDocumentParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDidCloseTextDocumentParams maps the parameters from a jsconrpc2.Request into protocol.DidCloseTextDocumentParams.
func RequestToDidCloseTextDocumentParams(req jsonrpc2.Request) (*protocol.DidCloseTextDocumentParams, error) {
	params := protocol.DidCloseTextDocumentParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDidOpenTextDocumentParams maps the parameters from a jsconrpc2.Request into protocol.DidOpenTextDocumentParams.
func RequestToDidOpenTextDocumentParams(req jsonrpc2.Request) (*protocol.DidOpenTextDocumentParams, error) {
	params := protocol.DidOpenTextDocumentParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDidSaveTextDocumentParams maps the parameters from a jsconrpc2.Request into protocol.DidSaveTextDocumentParams.
func RequestToDidSaveTextDocumentParams(req jsonrpc2.Request) (*protocol.DidSaveTextDocumentParams, error) {
	params := protocol.DidSaveTextDocumentParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToWillSaveTextDocumentParams maps the parameters from a jsconrpc2.Request into protocol.WillSaveTextDocumentParams.
func RequestToWillSaveTextDocumentParams(req jsonrpc2.Request) (*protocol.WillSaveTextDocumentParams, error) {
	params := protocol.WillSaveTextDocumentParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDidChangeWatchedFilesParams maps the parameters from a jsconrpc2.Request into protocol.DidChangeWatchedFilesParams.
func RequestToDidChangeWatchedFilesParams(req jsonrpc2.Request) (*protocol.DidChangeWatchedFilesParams, error) {
	params := protocol.DidChangeWatchedFilesParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToCodeActionParams maps the parameters from a jsconrpc2.Request into protocol.CodeActionParams.
func RequestToCodeActionParams(req jsonrpc2.Request) (*protocol.CodeActionParams, error) {
	params := protocol.CodeActionParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToCodeLensParams maps the parameters from a jsconrpc2.Request into protocol.CodeLensParams.
func RequestToCodeLensParams(req jsonrpc2.Request) (*protocol.CodeLensParams, error) {
	params := protocol.CodeLensParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToCodeLens maps the parameters from a jsconrpc2.Request into protocol.CodeLens.
func RequestToCodeLens(req jsonrpc2.Request) (*protocol.CodeLens, error) {
	params := protocol.CodeLens{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDeclarationParams maps the parameters from a jsonrpc2.Request into protocol.DeclarationParams
func RequestToDeclarationParams(req jsonrpc2.Request) (*protocol.DeclarationParams, error) {
	params := protocol.DeclarationParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDefinitionParams maps the parameters from a jsonrpc2.Request into protocol.DefinitionParams
func RequestToDefinitionParams(req jsonrpc2.Request) (*protocol.DefinitionParams, error) {
	params := protocol.DefinitionParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToTypeDefinitionParams maps the parameters from a jsonrpc2.Request into protocol.TypeDefinitionParams
func RequestToTypeDefinitionParams(req jsonrpc2.Request) (*protocol.TypeDefinitionParams, error) {
	params := protocol.TypeDefinitionParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToImplementationParams maps the parameters from a jsonrpc2.Request into protocol.ImplementationParams
func RequestToImplementationParams(req jsonrpc2.Request) (*protocol.ImplementationParams, error) {
	params := protocol.ImplementationParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToReferencesParams maps the parameters from a jsonrpc2.Request into protocol.ReferencesParams
func RequestToReferencesParams(req jsonrpc2.Request) (*protocol.ReferenceParams, error) {
	params := protocol.ReferenceParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToHoverParams maps the parameters from a jsconrpc2.Request into protocol.HoverParams.
func RequestToHoverParams(req jsonrpc2.Request) (*protocol.HoverParams, error) {
	params := protocol.HoverParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToDocumentSymbolParams maps the parameters from a jsconrpc2.Request into protocol.DocumentSymbolParams
func RequestToDocumentSymbolParams(req jsonrpc2.Request) (*protocol.DocumentSymbolParams, error) {
	params := protocol.DocumentSymbolParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// RequestToExecuteCommandParams maps the parameters from a jsconrpc2.Request into protocol.ExecuteCommandParams.
func RequestToExecuteCommandParams(req jsonrpc2.Request) (*protocol.ExecuteCommandParams, error) {
	params := protocol.ExecuteCommandParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}

	// store params.Arguments as []byte instead of []map[string]interface{}
	// this will allow plugins to handle unmarshalling the arguments themselves
	rawArgs := []interface{}{}
	for _, arg := range params.Arguments {
		rawArg, err := json.Marshal(arg)
		if err != nil {
			return nil, wrapErrParse(err)
		}
		rawArgs = append(rawArgs, rawArg)
	}

	params.Arguments = rawArgs
	return &params, nil
}

// RequestToWorkDoneProgressCancelParams maps the parameters from a jsconrpc2.Request into protocol.WorkDoneProgressCancelParams.
func RequestToWorkDoneProgressCancelParams(req jsonrpc2.Request) (*protocol.WorkDoneProgressCancelParams, error) {
	params := protocol.WorkDoneProgressCancelParams{}
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil, wrapErrParse(err)
	}
	return &params, nil
}

// NewCodeLens creates a new CodeLens with the given values.
func NewCodeLens(placementRange protocol.Range, title string, command string, arguments interface{}) protocol.CodeLens {
	return protocol.CodeLens{
		Range: placementRange,
		Command: &protocol.Command{
			Title:     title,
			Command:   command,
			Arguments: []interface{}{arguments},
		},
	}
}

// NewCodeAction creates a new CodeAction with the given values.
func NewCodeAction(title string, command string, kind protocol.CodeActionKind, arguments interface{}) protocol.CodeAction {
	return protocol.CodeAction{
		Title: title,
		Kind:  kind,
		Command: &protocol.Command{
			Title:     title,
			Command:   command,
			Arguments: []interface{}{arguments},
		},
	}
}

// CodeActionWithRange is a CodeAction that includes the range for which it is applicable.
type CodeActionWithRange struct {
	CodeAction protocol.CodeAction
	Range      protocol.Range
}

// NewCodeActionWithRange creates a new CodeActionWithRange with the given values.
func NewCodeActionWithRange(placementRange protocol.Range, title string, command string, kind protocol.CodeActionKind, arguments interface{}) CodeActionWithRange {
	return CodeActionWithRange{
		CodeAction: NewCodeAction(title, command, kind, arguments),
		Range:      placementRange,
	}
}

// ApplyContentChanges applies the given content change events to a given text string.
func ApplyContentChanges(initialText string, changes []protocol.TextDocumentContentChangeEvent) (string, error) {
	content := []byte(initialText)
	m := protocolmapper.NewTextOffsetMapper(content)
	for _, change := range changes {
		start, err := m.PositionOffset(change.Range.Start)
		if err != nil {
			return "", fmt.Errorf("unable to apply changes: %w", err)
		}
		end, err := m.PositionOffset(change.Range.End)
		if err != nil {
			return "", fmt.Errorf("unable to apply changes: %w", err)
		}
		var buf bytes.Buffer
		buf.Write(content[:start])
		buf.Write([]byte(change.Text))
		buf.Write(content[end:])
		content = buf.Bytes()
	}

	return string(content), nil
}

// DiffsToEditOffsets converts diffs into a list of text edits based on offsets within the initial text.
func DiffsToEditOffsets(diffs []diffmatchpatch.Diff) (initialText bytes.Buffer, offsets []EditOffset) {
	edits := make([]EditOffset, 0, len(diffs))
	offset := 0
	for _, d := range diffs {
		start := offset
		switch d.Type {
		case diffmatchpatch.DiffDelete:
			initialText.Write([]byte(d.Text))
			offset += len(d.Text)
			edits = append(edits, EditOffset{start: start, end: offset, text: ""})
		case diffmatchpatch.DiffEqual:
			initialText.Write([]byte(d.Text))
			offset += len(d.Text)
		case diffmatchpatch.DiffInsert:
			edits = append(edits, EditOffset{start: start, end: start, text: d.Text})
		}
	}
	return initialText, edits
}

// EditOffsetsToTextEdits converts a list of offset based edits to TextEdits formatted for LSP protocol.
func EditOffsetsToTextEdits(initialText bytes.Buffer, edits []EditOffset) ([]protocol.TextEdit, error) {
	protocolTextEdits := make([]protocol.TextEdit, 0, len(edits))
	m := protocolmapper.NewTextOffsetMapper(initialText.Bytes())
	for _, edit := range edits {
		startPosition, err := m.OffsetPosition(edit.start)
		if err != nil {
			return nil, err
		}
		endPosition, err := m.OffsetPosition(edit.end)
		if err != nil {
			return nil, err
		}
		protocolTextEdits = append(protocolTextEdits, rangeToTextEdit(PositionsToRange(startPosition, endPosition), edit.text))
	}
	return protocolTextEdits, nil
}

// DiffsToTextEdits converts diffs into to a list of text edits that can be applied to a document.
func DiffsToTextEdits(diffs []diffmatchpatch.Diff) ([]protocol.TextEdit, error) {
	foundText, edits := DiffsToEditOffsets(diffs)
	return EditOffsetsToTextEdits(foundText, edits)
}

// PositionsToRange converts two positions into a range.
func PositionsToRange(start, end protocol.Position) protocol.Range {
	return protocol.Range{
		Start: start,
		End:   end,
	}
}

func rangeToTextEdit(r protocol.Range, text string) protocol.TextEdit {
	return protocol.TextEdit{
		Range:   r,
		NewText: text,
	}
}

func wrapErrParse(err error) error {
	return fmt.Errorf("%s: %w", jsonrpc2.ErrParse, err)
}
