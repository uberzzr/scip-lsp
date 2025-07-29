package quickactions

import (
	"sort"
	"sync"

	"github.com/gofrs/uuid"
	"go.lsp.dev/protocol"
)

func newActionRangeStore() *actionRangeStore {
	return &actionRangeStore{
		actions:    make(map[uuid.UUID]map[protocol.TextDocumentIdentifier]map[protocol.Range][]protocol.CodeAction),
		codeLenses: make(map[uuid.UUID]map[protocol.TextDocumentIdentifier][]protocol.CodeLens),
		versions:   make(map[uuid.UUID]map[protocol.TextDocumentIdentifier]int32),
	}
}

type actionRangeStore struct {
	actions    map[uuid.UUID]map[protocol.TextDocumentIdentifier]map[protocol.Range][]protocol.CodeAction
	codeLenses map[uuid.UUID]map[protocol.TextDocumentIdentifier][]protocol.CodeLens
	versions   map[uuid.UUID]map[protocol.TextDocumentIdentifier]int32
	mu         sync.Mutex
}

func (a *actionRangeStore) SetVersion(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier, targetVersion int32) (shouldRefresh bool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.versions[sessionUUID] == nil {
		a.versions[sessionUUID] = make(map[protocol.TextDocumentIdentifier]int32)
	}

	if currentVersion, ok := a.versions[sessionUUID][document]; ok && currentVersion >= targetVersion {
		return false
	}

	a.versions[sessionUUID][document] = targetVersion
	return true
}

func (a *actionRangeStore) AddCodeAction(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier, actionRange protocol.Range, action protocol.CodeAction) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.actions[sessionUUID] == nil {
		a.actions[sessionUUID] = make(map[protocol.TextDocumentIdentifier]map[protocol.Range][]protocol.CodeAction)
	}

	if a.versions[sessionUUID] == nil {
		a.versions[sessionUUID] = make(map[protocol.TextDocumentIdentifier]int32)
	}

	if a.actions[sessionUUID][protocol.TextDocumentIdentifier{URI: document.URI}] == nil {
		a.actions[sessionUUID][document] = make(map[protocol.Range][]protocol.CodeAction)
	}
	a.actions[sessionUUID][document][actionRange] = append(a.actions[sessionUUID][document][actionRange], action)
}

func (a *actionRangeStore) AddCodeLens(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier, codeLens protocol.CodeLens) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.codeLenses[sessionUUID] == nil {
		a.codeLenses[sessionUUID] = make(map[protocol.TextDocumentIdentifier][]protocol.CodeLens)
	}

	if a.versions[sessionUUID] == nil {
		a.versions[sessionUUID] = make(map[protocol.TextDocumentIdentifier]int32)
	}

	if a.codeLenses[sessionUUID][protocol.TextDocumentIdentifier{URI: document.URI}] == nil {
		a.codeLenses[sessionUUID][document] = make([]protocol.CodeLens, 0)
	}
	a.codeLenses[sessionUUID][document] = append(a.codeLenses[sessionUUID][document], codeLens)
}

func (a *actionRangeStore) DeleteExistingDocumentRanges(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.actions[sessionUUID], document)
	delete(a.codeLenses[sessionUUID], document)
}

func (a *actionRangeStore) ClearDocument(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.actions[sessionUUID], document)
	delete(a.codeLenses[sessionUUID], document)
	delete(a.versions[sessionUUID], document)
}

func (a *actionRangeStore) DeleteSession(sessionUUID uuid.UUID) {
	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.actions, sessionUUID)
	delete(a.codeLenses, sessionUUID)
}

func (a *actionRangeStore) GetMatchingCodeActions(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier, searchRange protocol.Range) ([]protocol.CodeAction, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.actions[sessionUUID] == nil || a.actions[sessionUUID][document] == nil {
		return []protocol.CodeAction{}, nil
	}

	result := make([]protocol.CodeAction, 0)
	for actionRange, actions := range a.actions[sessionUUID][document] {
		isLineOverlapping := !(actionRange.End.Line < searchRange.Start.Line || actionRange.Start.Line > searchRange.End.Line)
		if isLineOverlapping {
			result = append(result, actions...)
		}
	}

	// Order by Command, then Title
	sort.Slice(result, func(i, j int) bool { return compareCodeActions(result[i], result[j]) })
	return result, nil
}

func (a *actionRangeStore) GetCodeLenses(sessionUUID uuid.UUID, document protocol.TextDocumentIdentifier) ([]protocol.CodeLens, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.codeLenses[sessionUUID] == nil || a.codeLenses[sessionUUID][document] == nil {
		return []protocol.CodeLens{}, nil
	}

	return a.codeLenses[sessionUUID][document], nil
}

func compareCodeActions(a, b protocol.CodeAction) bool {
	if a.Command == nil || b.Command == nil {
		return b.Command == nil && a.Title < b.Title
	}

	if a.Command.Command != b.Command.Command {
		return a.Command.Command <= b.Command.Command
	}

	return a.Title <= b.Title
}
