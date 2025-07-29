package serverinfofile

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/uber/scip-lsp/idl/mock/configmock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/config"
	_ "go.uber.org/fx"
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
			name: "all required params are present",
			params: Params{
				Lifecycle: lifecycleMock,
				Config:    newMockConfigProvider(ctrl, "valid"),
			},
			wantErr: false,
		},
		{
			name: "config processing error",
			params: Params{
				Lifecycle: lifecycleMock,
				Config:    newMockConfigProvider(ctrl, "missingKey"),
			},
			wantErr: true,
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

func TestOnStop(t *testing.T) {

	t.Run("file removed", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())

		m := module{
			logger:   zap.NewNop().Sugar(),
			infofile: tempFile.Name(),
		}

		_, err = os.Stat(tempFile.Name())
		assert.NoError(t, err)

		// Ensure no error return and file no longer present on disk.
		err = m.OnStop(context.Background())
		assert.NoError(t, err)
		_, err = os.Stat(tempFile.Name())
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("file removal error", func(t *testing.T) {
		// Create a temporary file in a read only directory, to force an error from os.Remove
		tempDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err)

		tempFile, err := os.CreateTemp(tempDir, "test")
		assert.NoError(t, err)
		tempFile.Close()

		err = os.Chmod(tempDir, 0555)
		assert.NoError(t, err)

		defer func() {
			os.Chmod(tempDir, 0755)
			os.RemoveAll(tempDir)
		}()

		m := module{
			logger:   zap.NewNop().Sugar(),
			infofile: tempDir,
		}

		err = m.OnStop(context.Background())
		assert.Error(t, err)
	})

}

func TestUpdateField(t *testing.T) {
	t.Run("multiple successful updates", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test")
		assert.NoError(t, err)
		defer os.Remove(tempFile.Name())

		m := module{
			infofile:     tempFile.Name(),
			logger:       zap.NewNop().Sugar(),
			fileContents: make(map[string]string),
		}

		// Make several step by step updates and confirm file contents are as expected
		steps := []struct {
			key        string
			value      string
			expectJSON string
		}{
			{
				key:        "key1",
				value:      "value1",
				expectJSON: "{\"key1\":\"value1\"}",
			},
			{
				key:        "key1",
				value:      "value2",
				expectJSON: "{\"key1\":\"value2\"}",
			},
			{
				key:        "key2",
				value:      "value2",
				expectJSON: "{\"key1\":\"value2\",\"key2\":\"value2\"}",
			},
			{
				key:        "key1",
				value:      "value3",
				expectJSON: "{\"key1\":\"value3\",\"key2\":\"value2\"}",
			},
		}

		for _, step := range steps {
			err = m.UpdateField(step.key, step.value)
			assert.NoError(t, err)
			assert.Equal(t, step.value, m.fileContents[step.key])
			contents, err := os.ReadFile(tempFile.Name())
			assert.NoError(t, err)
			assert.Equal(t, step.expectJSON, string(contents))
		}
	})

	t.Run("file write failure", func(t *testing.T) {
		// Create a directory instead of a file, to force a write failure
		tempDir, err := os.MkdirTemp("", "test")
		assert.NoError(t, err)
		defer os.RemoveAll(tempDir)

		m := module{
			infofile:     tempDir,
			logger:       zap.NewNop().Sugar(),
			fileContents: make(map[string]string),
		}
		err = m.UpdateField("key", "value")
		assert.Error(t, err)
	})
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
			name:        "missing path key",
			configKey:   "missingKey",
			wantErr:     true,
			errorString: "missing field \"serverInfoFilePath\" in config",
		},
		{
			name:        "missing path value",
			configKey:   "missingValue",
			wantErr:     true,
			errorString: "missing field \"serverInfoFilePath\" in config",
		},
		{
			name:        "incorrectly formatted entry",
			configKey:   "formatProblem",
			wantErr:     true,
			errorString: "getting config field \"serverInfoFilePath\": yaml: unmarshal errors:\n  line 1: cannot unmarshal !!map into string",
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

func newMockConfigProvider(ctrl *gomock.Controller, configKey string) config.Provider {
	configs := map[string]string{
		"valid": `
serverInfoFilePath: /my/sample/path/.ulspd
`,
		"missingKey": `
otherKey: /my/sample/path/.ulspd
`,
		"missingValue": `
serverInfoFilePath:
otherKey: sample
`,
		"formatProblem": `
serverInfoFilePath:
  infofile: /sample/.file
  address:
    key: val`,
	}

	yamlProv, _ := config.NewYAML(config.Source(strings.NewReader(configs[configKey])))
	configProviderMock := configmock.NewMockProvider(ctrl)
	configProviderMock.EXPECT().Get(_configKeyInfoFile).Return(yamlProv.Get(_configKeyInfoFile)).AnyTimes()
	return configProviderMock
}
