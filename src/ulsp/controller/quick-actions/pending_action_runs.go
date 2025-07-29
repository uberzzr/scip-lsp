package quickactions

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"go.lsp.dev/protocol"
)

type taskIdentifier struct {
	session    uuid.UUID
	actionName string
	argsToken  string
}

type pendingActionRunStore struct {
	mu                sync.Mutex
	inProgressActions map[taskIdentifier]*protocol.ProgressToken
}

func newInProgressActionStore() *pendingActionRunStore {
	return &pendingActionRunStore{
		inProgressActions: make(map[taskIdentifier]*protocol.ProgressToken),
	}
}

// AddInProgressAction adds an in-progress action to the store.
func (a *pendingActionRunStore) AddInProgressAction(sessionUUID uuid.UUID, actionName string, argsToken string) *protocol.ProgressToken {
	a.mu.Lock()
	defer a.mu.Unlock()

	curTask := newTaskIdentifier(sessionUUID, actionName, argsToken)
	progressToken := protocol.NewProgressToken(factory.UUID().String())
	a.inProgressActions[curTask] = progressToken
	return progressToken
}

// DeleteInProgressAction deletes an in-progress action from the store.
func (a *pendingActionRunStore) DeleteInProgressAction(sessionUUID uuid.UUID, actionName string, argsToken string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	curTask := newTaskIdentifier(sessionUUID, actionName, argsToken)
	if _, exist := a.inProgressActions[curTask]; !exist {
		return
	}
	delete(a.inProgressActions, curTask)
}

// TokenExists checks if a token exists for a given action and token
func (a *pendingActionRunStore) TokenExists(sessionUUID uuid.UUID, actionName string, argsToken string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	curTask := newTaskIdentifier(sessionUUID, actionName, argsToken)
	_, exist := a.inProgressActions[curTask]
	return exist
}

// GetInProgressActionRunToken returns the in-progress action token.
func (a *pendingActionRunStore) GetInProgressActionRunToken(sessionUUID uuid.UUID, actionName string, argsToken string) *protocol.ProgressToken {
	a.mu.Lock()
	defer a.mu.Unlock()

	curTask := newTaskIdentifier(sessionUUID, actionName, argsToken)
	if _, exist := a.inProgressActions[curTask]; !exist {
		return nil
	}

	return a.inProgressActions[curTask]
}

// GetInProgressActionRunTokens returns all in-progress action progress token associated with the session.
func (a *pendingActionRunStore) GetInProgressActionRunTokens(sessionUUID uuid.UUID) []*protocol.ProgressToken {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := make([]*protocol.ProgressToken, 0)
	for k, v := range a.inProgressActions {
		if k.session == sessionUUID {
			result = append(result, v)
		}
	}
	return result
}

// DeleteSession deletes all in-progress actions associated with the session.
func (a *pendingActionRunStore) DeleteSession(sessionUUID uuid.UUID) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for k := range a.inProgressActions {
		if k.session == sessionUUID {
			delete(a.inProgressActions, k)
		}
	}
}

func newTaskIdentifier(sessionUUID uuid.UUID, actionName string, argsToken string) taskIdentifier {
	return taskIdentifier{
		session:    sessionUUID,
		actionName: actionName,
		argsToken:  argsToken,
	}
}
