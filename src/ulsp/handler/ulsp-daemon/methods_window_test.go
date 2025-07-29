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

func TestWindowsMethods(t *testing.T) {
	t.Skip() // TODO @JamyDev: This is doing some odd type check failure
	tests := []struct {
		name             string
		method           string
		setReturn        func(c *ulspdaemonmock.MockController, result interface{}, err error)
		params           interface{}
		controllerResult interface{}
	}{
		{
			name:   "WorkDoneProgressCancel",
			method: protocol.MethodWorkDoneProgressCancel,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().WorkDoneProgressCancel(gomock.Any(), gomock.Any()).Return(err)
			},
			params:           protocol.WorkDoneProgressCancelParams{},
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

			// Controller error.
			tt.setReturn(c, tt.controllerResult, errors.New("err"))
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), tt.method, tt.params)
			err := r.HandleReq(ctx, replier, req)
			assert.Error(t, err)

			// Invalid params.
			if tt.params != nil {
				req, _ = jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), tt.method, 5)
				err = r.HandleReq(ctx, replier, req)
				assert.Error(t, err)
			}

		})
	}
}
