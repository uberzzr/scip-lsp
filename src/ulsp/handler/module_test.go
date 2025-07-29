package handler

import (
	"testing"

	_ "github.com/uber/scip-lsp/idl/ulsp/service"
	"go.uber.org/goleak"
)

func TestModule(t *testing.T) {
	// assert.NotPanics(t, func() { pb.NewFxUlspDaemonYARPCProcedures() })
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
