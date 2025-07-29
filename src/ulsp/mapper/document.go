package mapper

import (
	"bytes"
	"regexp"

	"go.lsp.dev/protocol"
)

// RegexMatchResult contains the result of a regular expression match.
type RegexMatchResult struct {
	// Full range of the text match
	Range protocol.Range
	// Full text match
	TextMatch string
	// Ordered contents of capturing groups
	CapturingGroups []string
}

// FindAllStringMatches determines the protocol.Range and matching text value for all matches of a regular expression in a string.
func FindAllStringMatches(regExp *regexp.Regexp, text string) []RegexMatchResult {
	textSubmatches := regExp.FindAllStringSubmatch(text, -1)
	indexMatches := regExp.FindAllStringIndex(text, -1)
	if indexMatches == nil {
		return nil
	}

	editOffsets := make([]EditOffset, len(indexMatches))
	for i := range indexMatches {
		editOffsets[i] = EditOffset{
			start: indexMatches[i][0],
			end:   indexMatches[i][1],
		}
	}

	ranges, err := EditOffsetsToTextEdits(*bytes.NewBufferString(text), editOffsets)
	if err != nil {
		return nil
	}

	result := make([]RegexMatchResult, len(ranges))
	for i := range ranges {
		result[i] = RegexMatchResult{
			Range:           ranges[i].Range,
			TextMatch:       textSubmatches[i][0],
			CapturingGroups: textSubmatches[i][1:],
		}
	}

	return result
}

// SingleEditToApplyWorkspaceEditParams creates a ApplyWorkspaceEditParams with a single edit to be applied based on the specified parameters.
func SingleEditToApplyWorkspaceEditParams(label string, doc protocol.TextDocumentIdentifier, editRange protocol.Range, newText string) *protocol.ApplyWorkspaceEditParams {
	return &protocol.ApplyWorkspaceEditParams{
		Label: label,
		Edit: protocol.WorkspaceEdit{
			DocumentChanges: []protocol.TextDocumentEdit{
				{
					TextDocument: protocol.OptionalVersionedTextDocumentIdentifier{TextDocumentIdentifier: doc},
					Edits: []protocol.TextEdit{
						{Range: editRange, NewText: newText},
					},
				},
			},
		},
	}
}
