package quickactions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"go.lsp.dev/protocol"
)

func TestNewActionRangeStore(t *testing.T) {
	assert.NotPanics(t, func() { newActionRangeStore() })
}

func TestAddCodeAction(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleRanges := []protocol.Range{}
	for i := 0; i < 5; i++ {
		sampleRanges = append(sampleRanges, factory.Range())
	}
	sampleCodeAction := protocol.CodeAction{Title: "sampleCodeAction"}
	store := newActionRangeStore()

	for _, r := range sampleRanges {
		store.AddCodeAction(sampleUUID, sampleDocument, r, sampleCodeAction)
	}

	for _, r := range sampleRanges {
		assert.Contains(t, store.actions[sampleUUID][sampleDocument][r], sampleCodeAction)
	}
}

func TestAddCodeLens(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleCodeLens := protocol.CodeLens{Command: &protocol.Command{Command: "sampleCommand"}}
	store := newActionRangeStore()

	store.AddCodeLens(sampleUUID, sampleDocument, sampleCodeLens)

	assert.Contains(t, store.codeLenses[sampleUUID][sampleDocument], sampleCodeLens)
}

func TestDeleteExistingDocumentRanges(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleRanges := []protocol.Range{}
	for i := 0; i < 5; i++ {
		sampleRanges = append(sampleRanges, factory.Range())
	}
	sampleCodeAction := protocol.CodeAction{Title: "sampleCodeAction"}
	sampleCodeLens := protocol.CodeLens{Command: &protocol.Command{Command: "sampleCommand"}}
	store := newActionRangeStore()

	for _, r := range sampleRanges {
		store.AddCodeAction(sampleUUID, sampleDocument, r, sampleCodeAction)
		store.AddCodeLens(sampleUUID, sampleDocument, sampleCodeLens)
	}

	store.DeleteExistingDocumentRanges(sampleUUID, sampleDocument)

	for _, r := range sampleRanges {
		assert.NotContains(t, store.actions[sampleUUID][sampleDocument][r], sampleCodeAction)
		assert.NotContains(t, store.codeLenses[sampleUUID][sampleDocument], sampleCodeLens)
	}
}

func TestClearDocument(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleRanges := []protocol.Range{}
	for i := 0; i < 5; i++ {
		sampleRanges = append(sampleRanges, factory.Range())
	}
	sampleCodeAction := protocol.CodeAction{Title: "sampleCodeAction"}
	sampleCodeLens := protocol.CodeLens{Command: &protocol.Command{Command: "sampleCommand"}}
	store := newActionRangeStore()
	store.SetVersion(sampleUUID, sampleDocument, 1)
	assert.NotNil(t, store.versions[sampleUUID][sampleDocument])

	for _, r := range sampleRanges {
		store.AddCodeAction(sampleUUID, sampleDocument, r, sampleCodeAction)
		store.AddCodeLens(sampleUUID, sampleDocument, sampleCodeLens)
	}

	store.ClearDocument(sampleUUID, sampleDocument)

	assert.Empty(t, store.actions[sampleUUID][sampleDocument])
	assert.Empty(t, store.versions[sampleUUID][sampleDocument])
	assert.Empty(t, store.codeLenses[sampleUUID][sampleDocument])
}

func TestDeleteSession(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleRanges := []protocol.Range{}
	for i := 0; i < 5; i++ {
		sampleRanges = append(sampleRanges, factory.Range())
	}
	sampleCodeAction := protocol.CodeAction{Title: "sampleCodeAction"}
	sampleCodeLens := protocol.CodeLens{Command: &protocol.Command{Command: "sampleCommand"}}
	store := newActionRangeStore()

	for _, r := range sampleRanges {
		store.AddCodeAction(sampleUUID, sampleDocument, r, sampleCodeAction)
		store.AddCodeLens(sampleUUID, sampleDocument, sampleCodeLens)
	}

	store.DeleteSession(sampleUUID)
	assert.Nil(t, store.actions[sampleUUID])
}

func TestGetMatchingCodeActions(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleRanges := []protocol.Range{}
	for i := 0; i < 5; i++ {
		r := factory.Range()

		// Ensure single line result.
		r.End.Line = r.Start.Line
		if r.Start.Character > r.End.Character {
			r.Start.Character, r.End.Character = r.End.Character, r.Start.Character
		}

		sampleRanges = append(sampleRanges, r)
	}

	sampleActions := []protocol.CodeAction{
		{
			Title: "sampleCodeAction0",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		},
		{
			Title: "sampleCodeActionC",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		},
		{
			Title: "sampleCodeActionA",
			Command: &protocol.Command{
				Command: "sampleCommand2",
			},
		},
		{
			Title: "sampleCodeActionB",
		},
		{
			Title: "sampleCodeActionD",
			Command: &protocol.Command{
				Command: "sampleCommand2",
			},
		},
	}

	store := newActionRangeStore()

	// Populate each sample range with each sample action.
	for _, r := range sampleRanges {
		for _, a := range sampleActions {
			store.AddCodeAction(sampleUUID, sampleDocument, r, a)
		}
	}

	t.Run("valid matches", func(t *testing.T) {
		for _, r := range sampleRanges {
			result, err := store.GetMatchingCodeActions(sampleUUID, sampleDocument, r)
			for _, a := range sampleActions {
				assert.Contains(t, result, a)
				assert.NoError(t, err)
			}
		}
	})

	t.Run("invalid session", func(t *testing.T) {
		result, _ := store.GetMatchingCodeActions(factory.UUID(), sampleDocument, sampleRanges[0])
		assert.Len(t, result, 0)
	})
}

func TestGetCodeLenses(t *testing.T) {
	sampleUUID := factory.UUID()
	sampleDocument := protocol.TextDocumentIdentifier{URI: "file:///test.go"}
	sampleCodeLenses := []protocol.CodeLens{
		{
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		},
		{
			Command: &protocol.Command{
				Command: "sampleCommand2",
			},
		},
		{
			Command: &protocol.Command{
				Command: "sampleCommand3",
			},
		},
	}

	store := newActionRangeStore()

	for _, c := range sampleCodeLenses {
		store.AddCodeLens(sampleUUID, sampleDocument, c)
	}

	t.Run("valid matches", func(t *testing.T) {
		result, err := store.GetCodeLenses(sampleUUID, sampleDocument)
		assert.NoError(t, err)
		assert.Equal(t, sampleCodeLenses, result)
	})

	t.Run("invalid session", func(t *testing.T) {
		result, _ := store.GetCodeLenses(factory.UUID(), sampleDocument)
		assert.Len(t, result, 0)
	})
}

func TestCompareCodeActions(t *testing.T) {
	t.Run("same command", func(t *testing.T) {
		a := protocol.CodeAction{
			Title: "sampleCodeAction0",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		}
		assert.True(t, compareCodeActions(a, a))
	})

	t.Run("different command", func(t *testing.T) {
		a := protocol.CodeAction{
			Title: "sampleCodeAction0",
			Command: &protocol.Command{
				Command: "sampleCommand2",
			},
		}

		b := protocol.CodeAction{
			Title: "sampleCodeAction0",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		}
		assert.False(t, compareCodeActions(a, b))
	})

	t.Run("nil command", func(t *testing.T) {
		a := protocol.CodeAction{
			Title: "sampleCodeAction0",
		}

		b := protocol.CodeAction{
			Title: "sampleCodeAction0",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		}
		assert.False(t, compareCodeActions(a, b))
	})

	t.Run("different title", func(t *testing.T) {
		a := protocol.CodeAction{
			Title: "sampleCodeAction2",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		}

		b := protocol.CodeAction{
			Title: "sampleCodeAction0",
			Command: &protocol.Command{
				Command: "sampleCommand1",
			},
		}
		assert.False(t, compareCodeActions(a, b))
	})
}
