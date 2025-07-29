package mapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/model"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/goleak"
)

func TestUlspDaemonToModel(t *testing.T) {
	f := &entity.UlspDaemon{
		Name: "test-daemon",
		UUID: factory.UUID(),
	}
	m := UlspDaemonToModel(f)
	assert.Equal(t, f.Name, m.Name)
	assert.Equal(t, f.UUID, m.UUID)
}

func TestUUIDToModel(t *testing.T) {
	u := factory.UUID()
	m := UUIDToModel(u)
	assert.Equal(t, u, m.UUID)
}

func TestSessionToModel(t *testing.T) {
	conn := jsonrpc2.NewConn(nil)
	f := &entity.Session{
		UUID:             factory.UUID(),
		InitializeParams: &protocol.InitializeParams{},
		Conn:             &conn,
		WorkspaceRoot:    "test/workspace",
		Monorepo:         "test-monorepo",
		Env:              []string{"key=val"},
		UlspEnabled:      true,
	}
	m := SessionToModel(f)
	assert.Equal(t, f.UUID, m.UUID)
	assert.Equal(t, f.InitializeParams, m.InitializeParams)
	assert.Equal(t, f.Conn, m.Conn)
	assert.Equal(t, f.WorkspaceRoot, m.WorkspaceRoot)
	assert.Equal(t, string(f.Monorepo), m.Monorepo)
	assert.Equal(t, f.Env, m.Env)
	assert.Equal(t, f.UlspEnabled, m.UlspEnabled)
}

func TestModelToSession(t *testing.T) {
	t.Run("valid model mapping", func(t *testing.T) {
		conn := jsonrpc2.NewConn(nil)
		m := &model.Session{
			UUID:             factory.UUID(),
			InitializeParams: &protocol.InitializeParams{},
			Conn:             &conn,
			WorkspaceRoot:    "test/workspace",
			Monorepo:         "test-monorepo",
			Env:              []string{"key=val"},
			UlspEnabled:      true,
		}
		f, err := ModelToSession(m)
		assert.NoError(t, err)
		assert.Equal(t, m.UUID, f.UUID)
		assert.Equal(t, m.InitializeParams, f.InitializeParams)
		assert.Equal(t, m.Conn, f.Conn)
		assert.Equal(t, m.WorkspaceRoot, f.WorkspaceRoot)
		assert.Equal(t, m.Monorepo, string(f.Monorepo))
		assert.Equal(t, m.Env, f.Env)
		assert.Equal(t, m.UlspEnabled, f.UlspEnabled)
	})
}

func TestUUIDToSession(t *testing.T) {
	u := factory.UUID()
	m := UUIDToSession(u, nil)
	assert.Equal(t, u, m.UUID)
	assert.True(t, m.UlspEnabled)
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
