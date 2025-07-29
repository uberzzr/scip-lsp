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

func TestInitialize(t *testing.T) {

	tests := []struct {
		name             string
		method           string
		params           interface{}
		controllerResult *protocol.InitializeResult
		controllerError  error
		wantErr          bool
	}{
		{
			name:             "error from controller",
			params:           protocol.InitializeParams{},
			controllerResult: nil,
			controllerError:  errors.New("controller error"),
			wantErr:          true,
		},
		{
			name:             "no error from controller",
			params:           protocol.InitializeParams{},
			controllerResult: &protocol.InitializeResult{},
			controllerError:  nil,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			replier := newMockReplier()

			c := ulspdaemonmock.NewMockController(ctrl)
			c.EXPECT().Initialize(gomock.Any(), gomock.Any()).Return(tt.controllerResult, tt.controllerError)

			r := jsonRPCRouter{ulspdaemon: c}
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), protocol.MethodInitialize, tt.params)
			err := r.HandleReq(ctx, replier, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInitialized(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		params         interface{}
		initializedErr error
		wantErr        bool
	}{
		{
			name:           "error from controller",
			params:         protocol.InitializedParams{},
			initializedErr: errors.New("initialized error"),
			wantErr:        true,
		},
		{
			name:           "no error from controller",
			params:         protocol.InitializedParams{},
			initializedErr: nil,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			replier := newMockReplier()

			c := ulspdaemonmock.NewMockController(ctrl)
			c.EXPECT().Initialized(gomock.Any(), gomock.Any()).Return(tt.initializedErr)

			r := jsonRPCRouter{ulspdaemon: c}
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), protocol.MethodInitialized, tt.params)
			err := r.HandleReq(ctx, replier, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestShutdown(t *testing.T) {
	tests := []struct {
		name            string
		controllerError error
		wantErr         bool
	}{
		{
			name:            "error from controller",
			controllerError: errors.New("controller error"),
			wantErr:         true,
		},
		{
			name:            "no error from controller",
			controllerError: nil,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			replier := newMockReplier()

			c := ulspdaemonmock.NewMockController(ctrl)
			c.EXPECT().Shutdown(gomock.Any()).Return(tt.controllerError)

			r := jsonRPCRouter{ulspdaemon: c}
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), protocol.MethodShutdown, nil)
			err := r.HandleReq(ctx, replier, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestExit(t *testing.T) {
	tests := []struct {
		name            string
		controllerError error
		wantErr         bool
	}{
		{
			name:            "error from controller",
			controllerError: errors.New("controller error"),
			wantErr:         true,
		},
		{
			name:            "no error from controller",
			controllerError: nil,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			replier := newMockReplier()

			c := ulspdaemonmock.NewMockController(ctrl)
			c.EXPECT().Exit(gomock.Any()).Return(tt.controllerError)

			r := jsonRPCRouter{ulspdaemon: c}
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), protocol.MethodExit, nil)
			err := r.HandleReq(ctx, replier, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRequestFullShutdown(t *testing.T) {
	tests := []struct {
		name            string
		controllerError error
		wantErr         bool
	}{
		{
			name:            "error from controller",
			controllerError: errors.New("controller error"),
			wantErr:         true,
		},
		{
			name:            "no error from controller",
			controllerError: nil,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			ctx := context.Background()
			replier := newMockReplier()

			c := ulspdaemonmock.NewMockController(ctrl)
			c.EXPECT().RequestFullShutdown(gomock.Any()).Return(tt.controllerError)

			r := jsonRPCRouter{ulspdaemon: c}
			req, _ := jsonrpc2.NewCall(jsonrpc2.NewNumberID(5), MethodRequestFullShutdown, nil)
			err := r.HandleReq(ctx, replier, req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
