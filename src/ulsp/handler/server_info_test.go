package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
)

func TestOutputYarpcConnectionInfo(t *testing.T) {
	t.Skip()
	// serverInfoFile := serverinfofilemock.NewMockServerInfoFile(ctrl)

	tests := []struct {
		name       string
		cfg        interface{}
		setupMocks func()
		wantErr    bool
	}{
		{
			name: "valid config",
			cfg: map[string]interface{}{
				"inbounds": map[interface{}]interface{}{
					"foo": map[interface{}]interface{}{
						"address": "sample:1234",
					},
					"bar": map[interface{}]interface{}{
						"address": "other:5678",
					},
				},
			},
			setupMocks: func() {
				// serverInfoFile.EXPECT().UpdateField(gomock.Any(), gomock.Any()).With(fmt.Sprintf(_fmtInfoFileKey, "foo", _configKeyAddress), "sample:1234").Return(nil)
				// serverInfoFile.EXPECT().UpdateField(gomock.Any(), gomock.Any()).With(fmt.Sprintf(_fmtInfoFileKey, "bar", _configKeyAddress), "other:5678").Return(nil)
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			cfg: map[string]interface{}{
				"inbounds": map[interface{}]interface{}{
					"bar": map[interface{}]interface{}{
						"address": map[interface{}]interface{}{},
					},
				},
			},
			setupMocks: func() {},
			wantErr:    true,
		},
		{
			name: "invalid inbounds",
			cfg: map[string]interface{}{
				"inbounds": "foo",
			},
			setupMocks: func() {},
			wantErr:    true,
		},
		{
			name: "invalid inbound contents",
			cfg: map[string]interface{}{
				"inbounds": map[interface{}]interface{}{
					"bar": "sample",
				},
			},
			setupMocks: func() {},
			wantErr:    true,
		},
		{
			name:       "invalid top level entry",
			cfg:        "sample",
			setupMocks: func() {},
			wantErr:    true,
		},
		{
			name: "file update error",
			cfg: map[string]interface{}{
				"inbounds": map[interface{}]interface{}{
					"sample": map[interface{}]interface{}{
						"address": "sample:1234",
					},
				},
			},
			setupMocks: func() {
				// serverInfoFile.EXPECT().UpdateField(gomock.Any(), gomock.Any()).With(fmt.Sprintf(_fmtInfoFileKey, "sample", _configKeyAddress), "sample:1234").Return(fmt.Errorf("sample"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip()
			_, err := config.NewStaticProvider(map[string]interface{}{"yarpc": tt.cfg})
			assert.NoError(t, err)
			tt.setupMocks()
			// err = outputYARPCConnectionInfo(cfg, serverInfoFile.Build())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
