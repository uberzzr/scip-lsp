package ulspdaemon

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"go.lsp.dev/jsonrpc2"
)

func TestHandleReq(t *testing.T) {
	ctx := context.Background()
	m := jsonRPCRouter{}

	request, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), "sampleMethod", []string{"val1", "val2"})
	err := m.HandleReq(ctx, newMockReplier(), request)
	assert.Error(t, err)
}

func TestUUID(t *testing.T) {
	sampleUUID := factory.UUID()
	m := jsonRPCRouter{uuid: sampleUUID}
	assert.Equal(t, sampleUUID, m.UUID())
}
