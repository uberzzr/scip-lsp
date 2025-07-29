package jsonrpcfx

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/uber/scip-lsp/idl/mock/configmock"
	"github.com/uber/scip-lsp/idl/mock/jsonrpc2mock"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/scip-lsp/src/ulsp/internal/serverinfofile/serverinfofilemock"
	"go.uber.org/config"
	"go.uber.org/fx/fxtest"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	lifecycleMock := fxtest.NewLifecycle(t)

	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			name:    "missing required params",
			params:  Params{},
			wantErr: true,
		},
		{
			name: "all required params are present",
			params: Params{
				Lifecycle: lifecycleMock,
				Config:    newMockConfigProvider(ctrl, "valid"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterRouter(t *testing.T) {
	ctrl := gomock.NewController(t)
	m := module{}

	mockConnectionManager := NewMockConnectionManager(ctrl)

	// first call should return no error
	err := m.RegisterConnectionManager(mockConnectionManager)
	assert.NoError(t, err)

	// duplicate call should return error
	err = m.RegisterConnectionManager(mockConnectionManager)
	assert.Error(t, err)
}

func TestServeStream(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	mockServer := module{
		logger: zap.NewNop().Sugar(),
	}

	mockUUID, _ := uuid.NewV4()
	mockRouter := NewMockRouter(ctrl)
	mockRouter.EXPECT().UUID().Return(mockUUID).AnyTimes()

	mockConnectionManager := NewMockConnectionManager(ctrl)
	mockConnectionManager.EXPECT().RemoveConnection(ctx, mockUUID)

	conn := jsonrpc2mock.NewMockConn(ctrl)
	conn.EXPECT().Go(gomock.Any(), gomock.Any())

	// Return a channel and immediately close it.
	c := make(chan struct{})
	conn.EXPECT().Done().Return(c)
	go func() {
		c <- struct{}{}
		close(c)
	}()

	conn.EXPECT().Err()

	tests := []struct {
		name                        string
		connectionManagerRegistered bool
		wantErr                     bool

		// Return values from NewConnection
		routerReturnVal Router
		errReturnVal    error
	}{
		{
			name:    "no connection manager registered",
			wantErr: true,

			connectionManagerRegistered: false,
			routerReturnVal:             nil,
			errReturnVal:                nil,
		},
		{
			name:    "failed NewConnection",
			wantErr: true,

			connectionManagerRegistered: true,
			routerReturnVal:             nil,
			errReturnVal:                errors.New("sample error"),
		},
		{
			name:    "successful NewConnection",
			wantErr: false,

			connectionManagerRegistered: true,
			routerReturnVal:             mockRouter,
			errReturnVal:                nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.connectionManagerRegistered {
				mockServer.RegisterConnectionManager(mockConnectionManager)
			}

			if tt.routerReturnVal != nil || tt.errReturnVal != nil {

				mockConnectionManager.EXPECT().NewConnection(gomock.Any(), gomock.Any()).Return(tt.routerReturnVal, tt.errReturnVal)
			}

			err := mockServer.ServeStream(ctx, conn)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	m := module{
		logger: zap.NewNop().Sugar(),
	}
	err := m.setup()
	assert.Error(t, err)

	m = module{Address: ":1234"}
	err = m.setup()
	assert.NoError(t, err)
}

func TestProcessConfig(t *testing.T) {

	tests := []struct {
		name        string
		configKey   string
		wantErr     bool
		errorString string
	}{
		{
			name:      "valid configuration",
			configKey: "valid",
			wantErr:   false,
		},
		{
			name:        "missing address key",
			configKey:   "missingKey",
			wantErr:     true,
			errorString: "missing field \"jsonrpc.address\" in config",
		},
		{
			name:        "missing address value",
			configKey:   "missingKey",
			wantErr:     true,
			errorString: "missing field \"jsonrpc.address\" in config",
		},
		{
			name:        "incorrectly formatted entry",
			configKey:   "formatProblem",
			wantErr:     true,
			errorString: "getting config field \"jsonrpc.address\": yaml: unmarshal errors:\n  line 1: cannot unmarshal !!map into string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gomockCtrl := gomock.NewController(t)
			cfg := newMockConfigProvider(gomockCtrl, tt.configKey)

			m := module{
				logger: zap.NewNop().Sugar(),
			}
			err := m.processConfig(cfg)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errorString, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	tmpDir, err := os.MkdirTemp(os.Getenv("TEST_TMPDIR"), t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	infoFileMock := serverinfofilemock.NewMockServerInfoFile(ctrl)

	m := module{
		Address:        ":1234",
		serverInfoFile: infoFileMock,
		logger:         zap.NewNop().Sugar(),
	}

	infoFileMock.EXPECT().UpdateField(_outputKey, m.Address).Return(nil)
	assert.Panics(t, func() { m.start() })
}

func TestOnStart(t *testing.T) {
	tmpDir, err := os.MkdirTemp(os.Getenv("TEST_TMPDIR"), t.Name())
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	ctx := context.Background()
	m := module{
		logger: zap.NewNop().Sugar(),
	}

	err = m.OnStart(ctx)
	assert.Error(t, err)
}

func newMockConfigProvider(ctrl *gomock.Controller, configKey string) config.Provider {
	configs := map[string]string{
		"valid": `
jsonrpc:
  address: :5859
  infofile: /sample/.file`,
		"missingKey": `
jsonrpc:
  infofile: /sample/.file`,
		"missingValue": `
jsonrpc:
  address: :5859
  infofile:`,
		"formatProblem": `
jsonrpc:
  infofile: /sample/.file
  address:
    key: val`,
	}

	yamlProv, _ := config.NewYAML(config.Source(strings.NewReader(configs[configKey])))
	configProviderMock := configmock.NewMockProvider(ctrl)
	configProviderMock.EXPECT().Get(_configKeyAddress).Return(yamlProv.Get(_configKeyAddress)).AnyTimes()
	return configProviderMock
}
