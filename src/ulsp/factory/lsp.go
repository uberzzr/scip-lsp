package factory

import (
	"math/rand"

	"go.lsp.dev/protocol"
)

// Range returns a random protocol.Range.
func Range() protocol.Range {
	start := protocol.Position{Line: uint32(rand.Intn(100)), Character: uint32(rand.Intn(100))}
	end := protocol.Position{Line: start.Line + uint32(rand.Intn(100)), Character: uint32(rand.Intn(100))}

	if start.Line == end.Line && start.Character > end.Character {
		end.Character = start.Character + uint32(rand.Intn(100))
	}

	return protocol.Range{
		Start: start,
		End:   end,
	}
}
