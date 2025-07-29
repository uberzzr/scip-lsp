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

func TestDocumentMethods(t *testing.T) {

	tests := []struct {
		name             string
		method           string
		setReturn        func(c *ulspdaemonmock.MockController, result interface{}, err error)
		params           interface{}
		controllerResult interface{}
	}{
		{
			name:   "DidChange",
			method: protocol.MethodTextDocumentDidChange,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidChange(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.DidChangeTextDocumentParams{},
		},
		{
			name:   "DidChangeWatchedFiles",
			method: protocol.MethodWorkspaceDidChangeWatchedFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidChangeWatchedFiles(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.DidChangeWatchedFilesParams{},
		},
		{
			name:   "DidOpen",
			method: protocol.MethodTextDocumentDidOpen,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidOpen(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.DidOpenTextDocumentParams{},
		},
		{
			name:   "DidClose",
			method: protocol.MethodTextDocumentDidClose,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidClose(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.DidCloseTextDocumentParams{},
		},
		{
			name:   "WillSave",
			method: protocol.MethodTextDocumentWillSave,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().WillSave(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.WillSaveTextDocumentParams{},
		},
		{
			name:   "WillSaveWaitUntil",
			method: protocol.MethodTextDocumentWillSaveWaitUntil,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.TextEdit)
				c.EXPECT().WillSaveWaitUntil(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.WillSaveTextDocumentParams{},
			controllerResult: []protocol.TextEdit{{}, {}, {}},
		},
		{
			name:   "DidSave",
			method: protocol.MethodTextDocumentDidSave,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidSave(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.DidSaveTextDocumentParams{},
		},
		{
			name:   "WillRenameFiles",
			method: protocol.MethodWillRenameFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.(*protocol.WorkspaceEdit)
				c.EXPECT().WillRenameFiles(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.RenameFilesParams{},
			controllerResult: &protocol.WorkspaceEdit{},
		},
		{
			name:   "DidRenameFiles",
			method: protocol.MethodDidRenameFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidRenameFiles(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.RenameFilesParams{},
		},
		{
			name:   "WillCreateFiles",
			method: protocol.MethodWillCreateFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.(*protocol.WorkspaceEdit)
				c.EXPECT().WillCreateFiles(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.CreateFilesParams{},
			controllerResult: &protocol.WorkspaceEdit{},
		},
		{
			name:   "DidCreateFiles",
			method: protocol.MethodDidCreateFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidCreateFiles(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.CreateFilesParams{},
		},
		{
			name:   "WillDeleteFiles",
			method: protocol.MethodWillDeleteFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.(*protocol.WorkspaceEdit)
				c.EXPECT().WillDeleteFiles(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.DeleteFilesParams{},
			controllerResult: &protocol.WorkspaceEdit{},
		},
		{
			name:   "DidDeleteFiles",
			method: protocol.MethodDidDeleteFiles,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().DidDeleteFiles(gomock.Any(), gomock.Any()).Return(err)
			},
			params: protocol.DeleteFilesParams{},
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
