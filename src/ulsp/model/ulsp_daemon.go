package model

import (
	"github.com/gofrs/uuid"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// UlspDaemon is the repository layer model for a user's UlspDaemon.
type UlspDaemon struct {
	Name string
	UUID uuid.UUID
}

// Session is the repository layer model for an individual IDE session.
type Session struct {
	UUID             uuid.UUID
	InitializeParams *protocol.InitializeParams
	Conn             *jsonrpc2.Conn
	WorkspaceRoot    string
	Monorepo         string
	Env              []string
	UlspEnabled      bool
}
