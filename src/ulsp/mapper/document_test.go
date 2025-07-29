package mapper

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

var _regexpInterfaceDef = regexp.MustCompile(`(?m)^type\s+(\w+)\s+interface\s*\{`)

func TestFindAllStringMatches(t *testing.T) {
	tests := []struct {
		name     string
		regExp   *regexp.Regexp
		text     string
		expected []RegexMatchResult
	}{
		{
			name:   "no matches",
			regExp: regexp.MustCompile(`nonexistent$`),
			text:   `package sample\n\ntype myInterface interface {\nmyValue string\n}`,
		},
		{
			name:   "one match",
			regExp: _regexpInterfaceDef,
			text: `
package sample

type myInterface interface {
	GetData() string
}
`,
			expected: []RegexMatchResult{
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      3,
							Character: 0,
						},
						End: protocol.Position{
							Line:      3,
							Character: 28,
						},
					},
					TextMatch:       "type myInterface interface {",
					CapturingGroups: []string{"myInterface"},
				},
			},
		},
		{
			name:   "multiple matches",
			regExp: _regexpInterfaceDef,
			text: `
package sample

type myInterface interface {
	GetData() string
}

type anotherInterface interface {
	GetData() string
}
`,
			expected: []RegexMatchResult{
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      3,
							Character: 0,
						},
						End: protocol.Position{
							Line:      3,
							Character: 28,
						},
					},
					TextMatch:       "type myInterface interface {",
					CapturingGroups: []string{"myInterface"},
				},
				{
					Range: protocol.Range{
						Start: protocol.Position{
							Line:      7,
							Character: 0,
						},
						End: protocol.Position{
							Line:      7,
							Character: 33,
						},
					},
					TextMatch:       "type anotherInterface interface {",
					CapturingGroups: []string{"anotherInterface"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := FindAllStringMatches(tt.regExp, tt.text)
			assert.ElementsMatch(t, tt.expected, results)
		})
	}

}

func TestSingleEditToApplyWorkspaceEditParams(t *testing.T) {
	doc := protocol.TextDocumentIdentifier{
		URI: "file:///sample/file.go",
	}
	start := protocol.Position{
		Line:      1,
		Character: 3,
	}
	end := protocol.Position{
		Line:      1,
		Character: 5,
	}
	newText := "new text"

	expected := &protocol.ApplyWorkspaceEditParams{
		Label: "label",
		Edit: protocol.WorkspaceEdit{
			DocumentChanges: []protocol.TextDocumentEdit{
				{
					TextDocument: protocol.OptionalVersionedTextDocumentIdentifier{TextDocumentIdentifier: doc},
					Edits: []protocol.TextEdit{
						{Range: protocol.Range{Start: start, End: end}, NewText: newText},
					},
				},
			},
		},
	}
	editRange := protocol.Range{Start: start, End: end}
	actual := SingleEditToApplyWorkspaceEditParams("label", doc, editRange, newText)
	assert.Equal(t, expected, actual)
}
