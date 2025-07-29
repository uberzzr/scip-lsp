package ulspdaemon

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/controller/ulsp-daemon/ulspdaemonmock"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
)

func TestWorkspaceMethods(t *testing.T) {

	tests := []struct {
		name             string
		method           string
		setReturn        func(c *ulspdaemonmock.MockController, result interface{}, err error)
		params           interface{}
		controllerResult interface{}
	}{
		{
			name:   "ExecuteCommand",
			method: protocol.MethodWorkspaceExecuteCommand,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().ExecuteCommand(gomock.Any(), gomock.Any()).Return(nil, err)
			},
			params:           protocol.ExecuteCommandParams{},
			controllerResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			replier := newMockReplier()

			c := ulspdaemonmock.NewMockController(ctrl)
			r := jsonRPCRouter{ulspdaemon: c}

			// Valid params.
			tt.setReturn(c, tt.controllerResult, nil)
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), tt.method, tt.params)
			err := r.HandleReq(ctx, replier, req)
			assert.NoError(t, err)

			// Invalid params.
			if tt.params != nil {
				req, _ = jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), tt.method, 5)
				err = r.HandleReq(ctx, replier, req)
				assert.Error(t, err)
			}

			// Controller error.
			tt.setReturn(c, tt.controllerResult, errors.New("err"))
			req, _ = jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), tt.method, tt.params)
			err = r.HandleReq(ctx, replier, req)
			assert.Error(t, err)
		})
	}
}
