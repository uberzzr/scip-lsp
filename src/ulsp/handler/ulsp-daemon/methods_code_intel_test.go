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

func TestCodeIntelMethods(t *testing.T) {

	tests := []struct {
		name             string
		method           string
		setReturn        func(c *ulspdaemonmock.MockController, result interface{}, err error)
		params           interface{}
		controllerResult interface{}
	}{
		{
			name:   "CodeAction",
			method: protocol.MethodTextDocumentCodeAction,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.CodeAction)
				c.EXPECT().CodeAction(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.CodeActionParams{},
			controllerResult: []protocol.CodeAction{},
		},
		{
			name:   "CodeLens",
			method: protocol.MethodTextDocumentCodeLens,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.CodeLens)
				c.EXPECT().CodeLens(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.CodeLensParams{},
			controllerResult: []protocol.CodeLens{},
		},
		{
			name:   "CodeLensRefresh",
			method: protocol.MethodCodeLensRefresh,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				c.EXPECT().CodeLensRefresh(gomock.Any()).Return(err)
			},
		},
		{
			name:   "CodeLensResolve",
			method: protocol.MethodCodeLensResolve,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.(*protocol.CodeLens)
				c.EXPECT().CodeLensResolve(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.CodeLens{},
			controllerResult: &protocol.CodeLens{},
		},
		{
			name:   "GotoDeclaration",
			method: protocol.MethodTextDocumentDeclaration,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.LocationLink)
				c.EXPECT().GotoDeclaration(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.DeclarationParams{},
			controllerResult: []protocol.LocationLink{},
		},
		{
			name:   "GotoDefinition",
			method: protocol.MethodTextDocumentDefinition,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.LocationLink)
				c.EXPECT().GotoDefinition(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.DefinitionParams{},
			controllerResult: []protocol.LocationLink{},
		},
		{
			name:   "GotoTypeDefinition",
			method: protocol.MethodTextDocumentTypeDefinition,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.LocationLink)
				c.EXPECT().GotoTypeDefinition(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.TypeDefinitionParams{},
			controllerResult: []protocol.LocationLink{},
		},
		{
			name:   "GotoImplementation",
			method: protocol.MethodTextDocumentImplementation,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.LocationLink)
				c.EXPECT().GotoImplementation(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.ImplementationParams{},
			controllerResult: []protocol.LocationLink{},
		},
		{
			name:   "References",
			method: protocol.MethodTextDocumentReferences,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				r := result.([]protocol.Location)
				c.EXPECT().References(gomock.Any(), gomock.Any()).Return(r, err)
			},
			params:           protocol.ReferenceParams{},
			controllerResult: []protocol.Location{},
		},
		{
			name:   "DocumentSymbol",
			method: protocol.MethodTextDocumentDocumentSymbol,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				h := result.([]protocol.DocumentSymbol)
				c.EXPECT().DocumentSymbol(gomock.Any(), gomock.Any()).Return(h, err)
			},
			params:           protocol.DocumentSymbolParams{},
			controllerResult: []protocol.DocumentSymbol{},
		},
		{
			name:   "Hover",
			method: protocol.MethodTextDocumentHover,
			setReturn: func(c *ulspdaemonmock.MockController, result interface{}, err error) {
				h := result.(*protocol.Hover)
				c.EXPECT().Hover(gomock.Any(), gomock.Any()).Return(h, err)
			},
			params: protocol.HoverParams{},
			controllerResult: &protocol.Hover{
				Contents: protocol.MarkupContent{
					Kind:  protocol.Markdown,
					Value: "test",
				},
			},
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
