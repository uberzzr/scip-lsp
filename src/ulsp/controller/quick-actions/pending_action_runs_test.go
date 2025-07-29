package quickactions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"go.lsp.dev/protocol"
)

func TestAddInProgressAction(t *testing.T) {
	store := newInProgressActionStore()
	sessionUUID := factory.UUID()
	actionName := "testAction"
	token := "token1"

	store.AddInProgressAction(sessionUUID, actionName, token)

	exists := store.TokenExists(sessionUUID, actionName, token)
	assert.True(t, exists, "Token should exist after being added")
}

func TestDeleteInProgressAction(t *testing.T) {
	store := newInProgressActionStore()
	validUUID := factory.UUID()
	nonExistentUUID := factory.UUID()
	actionName := "testAction"
	token := "token1"

	store.AddInProgressAction(validUUID, actionName, token)

	store.DeleteInProgressAction(nonExistentUUID, "dummy", "dummy")
	store.DeleteInProgressAction(validUUID, "dummy", "dummy")
	store.DeleteInProgressAction(validUUID, actionName, "dummy")

	exists := store.TokenExists(validUUID, actionName, token)
	assert.True(t, exists, "Token should exist after invalid delete attempts")

	store.DeleteInProgressAction(validUUID, actionName, token)

	exists = store.TokenExists(validUUID, actionName, token)
	assert.False(t, exists, "Token should not exist after being deleted")
}

func TestTokenExists(t *testing.T) {
	store := newInProgressActionStore()
	sessionUUID := factory.UUID()
	actionName := "testAction"
	token := "token1"

	store.AddInProgressAction(sessionUUID, actionName, token)

	exists := store.TokenExists(sessionUUID, actionName, token)
	assert.True(t, exists, "Token should exist after being added")

	invalidSessionUUID := factory.UUID()
	exists = store.TokenExists(invalidSessionUUID, "dummy", "dummy")
	assert.False(t, exists, "Token should not for invalid session")

	exists = store.TokenExists(sessionUUID, "dummyAction", "dummy")
	assert.False(t, exists, "Token should not exist actions was not added")

	exists = store.TokenExists(sessionUUID, actionName, "nonExistentToken")
	assert.False(t, exists, "Token should not exist if token was not added")

}

func TestGetInProgressActionRunToken(t *testing.T) {
	store := newInProgressActionStore()
	sessionUUID := factory.UUID()
	actionName := "testAction"
	token := "token1"

	actualToken := store.AddInProgressAction(sessionUUID, actionName, token)

	retrievedToken := store.GetInProgressActionRunToken(sessionUUID, actionName, token)
	assert.NotNil(t, retrievedToken, "Retrieved token should not be nil")
	assert.Equal(t, actualToken, retrievedToken, "Retrieved token should match the added token")

	retrievedToken = store.GetInProgressActionRunToken(sessionUUID, actionName, "nonExistentToken")
	assert.Nil(t, retrievedToken, "Retrieved token should be nil for a non-existent token")

	retrievedToken = store.GetInProgressActionRunToken(sessionUUID, "nonexistentAction", "asad")
	assert.Nil(t, retrievedToken, "Retrieved token should be nil for a non-existent action")

	retrievedToken = store.GetInProgressActionRunToken(factory.UUID(), "dasd", "asad")
	assert.Nil(t, retrievedToken, "Retrieved token should be nil for a non-existent sessionID")
}

func TestInProgressRunDeleteSession(t *testing.T) {
	store := newInProgressActionStore()
	session1 := factory.UUID()
	session2 := factory.UUID()

	store.AddInProgressAction(session1, "action", "token1")
	store.AddInProgressAction(session1, "action", "token2")
	store.AddInProgressAction(session1, "action1", "token1")

	store.AddInProgressAction(session2, "action", "token1")

	assert.Len(t, store.inProgressActions, 4, "Store should have 4 in-progress actions")

	store.DeleteSession(session1)
	assert.Len(t, store.inProgressActions, 1, "Store should have 1 in-progress actions after ending session1")
	exist := store.TokenExists(session2, "action", "token1")
	assert.True(t, exist, "Token for session2 should exist after ending session1")

	store.DeleteSession(session2)
	assert.Len(t, store.inProgressActions, 0, "store should be empty")
}

func TestGetInProgressActionRunTokens(t *testing.T) {
	store := newInProgressActionStore()
	sessionID1 := factory.UUID()
	sessionID2 := factory.UUID()

	t1Token := store.AddInProgressAction(sessionID1, "sadasd", "t1")
	t2Token := store.AddInProgressAction(sessionID1, "das", "t2")
	t3Token := store.AddInProgressAction(sessionID2, "action2", "t3")

	tokens := store.GetInProgressActionRunTokens(sessionID1)
	assert.Len(t, tokens, 2)
	assert.ElementsMatch(t, tokens, []*protocol.ProgressToken{t1Token, t2Token})

	tokens = store.GetInProgressActionRunTokens(sessionID2)
	assert.Len(t, tokens, 1)
	assert.ElementsMatch(t, tokens, []*protocol.ProgressToken{t3Token})

	tokens = store.GetInProgressActionRunTokens(factory.UUID())
	assert.Len(t, tokens, 0)
}
