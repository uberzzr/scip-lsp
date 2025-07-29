package types

import (
	"go.lsp.dev/protocol"
)

// BreakpointLocation represents a breakpoint location.
type BreakpointLocation struct {
	Line            uint32
	Column          uint32
	ClassName       string
	MethodName      string
	MethodSignature string
}

// ResolveBreakpoints represents a request to resolve breakpoints.
type ResolveBreakpoints struct {
	WorkspaceRoot string
	SourceURI     string
	Breakpoints   []*protocol.Position
}

// ResolveClassToPath represents a request to resolve a class to a path.
type ResolveClassToPath struct {
	WorkspaceRoot      string
	FullyQualifiedName string
	SourceRelativePath string
}
