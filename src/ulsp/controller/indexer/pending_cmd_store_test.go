package indexer

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"go.lsp.dev/protocol"
)

var session1 = uuid.Must(uuid.NewV4())
var session2 = uuid.Must(uuid.NewV4())
var cancelFunc = func() {}

var sampleItems = []struct {
	session uuid.UUID
	doc     protocol.TextDocumentItem
}{
	{
		session: session1,
		doc:     getProtoDoc("file:///foo/bar.go"),
	},
	{
		session: session1,
		doc:     getProtoDoc("file:///foo/bar2.go"),
	},
	{
		session: session1,
		doc:     getProtoDoc("file:///foo/bar3.go"),
	},
	{
		session: session2,
		doc:     getProtoDoc("file:///foo/bar.go"),
	},
	{
		session: session2,
		doc:     getProtoDoc("file:///abc/foo/bar.go"),
	},
	{
		session: session2,
		doc:     getProtoDoc("file:///foo/bar3.go"),
	},
}

func TestSetPendingCmd(t *testing.T) {
	c := pendingCmdStore{}

	for _, item := range sampleItems {
		key := getCmdKeyUtil(item.session, item.doc)
		c.setPendingCmd(key, cancelFunc)
	}

	assert.Equal(t, len(sampleItems), len(c.pendingCmds))
}

func TestGetPendingCmd(t *testing.T) {
	c := pendingCmdStore{}

	cancelF, _, ok := c.getPendingCmd("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, cancelF)

	for _, item := range sampleItems {
		key := getCmdKeyUtil(item.session, item.doc)
		c.setPendingCmd(key, cancelFunc)
	}

	for _, item := range sampleItems {
		key := getCmdKeyUtil(item.session, item.doc)
		cancelF, token, ok := c.getPendingCmd(key)
		assert.True(t, ok)
		assert.NotNil(t, cancelF)
		assert.NotEmpty(t, token)
	}
}

func TestDeletePendingCmd(t *testing.T) {

	c := pendingCmdStore{}
	assert.False(t, c.deletePendingCmd("empty store"))

	for _, item := range sampleItems {
		key := getCmdKeyUtil(item.session, item.doc)
		c.setPendingCmd(key, cancelFunc)
	}

	for _, item := range sampleItems {
		key := getCmdKeyUtil(item.session, item.doc)
		ok := c.deletePendingCmd(key)
		assert.True(t, ok)
	}

	assert.Equal(t, 0, len(c.pendingCmds))
	assert.False(t, c.deletePendingCmd("nonexistent"))
}

func TestCleanSession(t *testing.T) {
	c := pendingCmdStore{}

	c.cleanSession(session1)

	for _, item := range sampleItems {
		key := getCmdKeyUtil(item.session, item.doc)
		c.setPendingCmd(key, cancelFunc)
	}

	c.cleanSession(session1)

	assert.Equal(t, 3, len(c.pendingCmds))
}

func TestGetContainingKey(t *testing.T) {

	t.Run("non existent", func(t *testing.T) {
		c := pendingCmdStore{}
		_, err := c.getContainingKey("nonexistent")
		assert.Error(t, err)
	})

	t.Run("key found", func(t *testing.T) {
		c := pendingCmdStore{}
		key := getCmdKeyUtil(sampleItems[0].session, sampleItems[0].doc)
		token := c.setPendingCmd(key, cancelFunc)

		obtainedKey, err := c.getContainingKey(token)
		assert.NoError(t, err)
		assert.Equal(t, key, obtainedKey)
	})

	t.Run("key not found", func(t *testing.T) {
		c := pendingCmdStore{}
		key := getCmdKeyUtil(sampleItems[0].session, sampleItems[0].doc)
		c.setPendingCmd(key, cancelFunc)

		_, err := c.getContainingKey("non-existent")
		assert.Error(t, err)

	})
}

func TestMarkForReindexing(t *testing.T) {
	t.Run("empty store", func(t *testing.T) {
		c := pendingCmdStore{}
		ok := c.markForReindexing("nonexistent")
		assert.False(t, ok)
	})

	t.Run("key not found", func(t *testing.T) {
		c := pendingCmdStore{}
		c.setPendingCmd("existing-key", cancelFunc)
		ok := c.markForReindexing("nonexistent")
		assert.False(t, ok)
	})

	t.Run("mark for reindexing success", func(t *testing.T) {
		c := pendingCmdStore{}
		key := getCmdKeyUtil(sampleItems[0].session, sampleItems[0].doc)
		c.setPendingCmd(key, cancelFunc)

		ok := c.markForReindexing(key)
		assert.True(t, ok)

		val, _, exists := c.getPendingCmd(key)
		assert.True(t, exists)
		assert.NotNil(t, val)

		assert.True(t, c.needsReindexing(key))
	})
}

func TestNeedsReindexing(t *testing.T) {
	t.Run("empty store", func(t *testing.T) {
		c := pendingCmdStore{}
		needs := c.needsReindexing("nonexistent")
		assert.False(t, needs)
	})

	t.Run("key not found", func(t *testing.T) {
		c := pendingCmdStore{}
		c.setPendingCmd("existing-key", cancelFunc)
		needs := c.needsReindexing("nonexistent")
		assert.False(t, needs)
	})

	t.Run("not marked for reindexing", func(t *testing.T) {
		c := pendingCmdStore{}
		key := getCmdKeyUtil(sampleItems[0].session, sampleItems[0].doc)
		c.setPendingCmd(key, cancelFunc)

		needs := c.needsReindexing(key)
		assert.False(t, needs)
	})

	t.Run("marked for reindexing", func(t *testing.T) {
		c := pendingCmdStore{}
		key := getCmdKeyUtil(sampleItems[0].session, sampleItems[0].doc)
		c.setPendingCmd(key, cancelFunc)
		c.markForReindexing(key)

		needs := c.needsReindexing(key)
		assert.True(t, needs)
	})
}

func TestReindexingFlow(t *testing.T) {
	// This test verifies the flow of the reindexing feature

	// Create a store and add a command
	store := pendingCmdStore{}
	key := "document-key"
	cancelCalled := false
	store.setPendingCmd(key, func() {
		cancelCalled = true
	})

	// Initial state check - no reindexing needed yet
	assert.False(t, store.needsReindexing(key))

	// Mark for reindexing when a new didSave comes in
	assert.True(t, store.markForReindexing(key))

	// Verify reindexing is now needed
	assert.True(t, store.needsReindexing(key))

	// Delete command after indexing completes
	assert.True(t, store.deletePendingCmd(key))

	// Command should now be gone
	_, _, ok := store.getPendingCmd(key)
	assert.False(t, ok)

	// Verify cancel was not called automatically
	assert.False(t, cancelCalled)
}

func getCmdKeyUtil(uuid uuid.UUID, doc protocol.TextDocumentItem) string {
	return fmt.Sprintf("%s-%s", uuid.String(), doc.URI.Filename())
}

func getProtoDoc(text string) protocol.TextDocumentItem {
	return protocol.TextDocumentItem{
		URI: protocol.DocumentURI(text),
	}
}
