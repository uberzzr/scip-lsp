package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/config"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	// t.Skip() // TODO: @JamyDev look into config safety
	tests := []struct {
		name           string
		loggingConfig  string
		expectedLevel  zapcore.Level
		expectedFormat string
		expectError    bool
	}{
		{
			name: "info level json encoding",
			loggingConfig: `
logging:
  level: info
  development: false
  encoding: json
  outputPaths:
    - stdout
`,
			expectedLevel:  zapcore.InfoLevel,
			expectedFormat: "json",
			expectError:    false,
		},
		{
			name: "debug level console encoding",
			loggingConfig: `
logging:
  level: debug
  development: true
  encoding: console
  outputPaths:
    - stdout
`,
			expectedLevel:  zapcore.DebugLevel,
			expectedFormat: "console",
			expectError:    false,
		},
		{
			name: "error level default encoding",
			loggingConfig: `
logging:
  level: error
  development: false
  outputPaths:
    - stdout
`,
			expectedLevel:  zapcore.ErrorLevel,
			expectedFormat: "json", // default
			expectError:    false,
		},
		{
			name: "invalid level",
			loggingConfig: `
logging:
  level: invalid
  development: false
  encoding: json
  outputPaths:
    - stdout
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config provider with the test config
			provider, err := config.NewYAML(
				config.Source(strings.NewReader(tt.loggingConfig)),
			)
			require.NoError(t, err)

			sugared, err := NewSugaredLogger(provider)
			if tt.expectError {
				assert.Error(t, err, "Expected error for invalid config")
				return
			}
			// Create the logger
			logger := NewLogger(sugared)

			require.NoError(t, err)
			require.NotNil(t, logger)

			// Test that the logger can be used
			logger.Info("test message")
		})
	}
}

func TestLoggingConfig_Populate(t *testing.T) {
	configYAML := strings.NewReader(`
logging:
  level: warn
  development: true
  encoding: console
  outputPaths:
    - stdout
    - stderr
`)

	provider, err := config.NewYAML(config.Source(configYAML))
	require.NoError(t, err)

	var loggingConfig LoggingConfig
	err = provider.Get("logging").Populate(&loggingConfig)
	require.NoError(t, err)

	assert.Equal(t, "warn", loggingConfig.Level)
	assert.True(t, loggingConfig.Development)
	assert.Equal(t, "console", loggingConfig.Encoding)
	assert.Equal(t, []string{"stdout", "stderr"}, loggingConfig.OutputPaths)
}

func TestLogger_Integration(t *testing.T) {
	// Create a test config with development mode
	configYAML := strings.NewReader(`
logging:
  level: debug
  development: true
  encoding: console
  outputPaths:
    - stdout
`)

	provider, err := config.NewYAML(config.Source(configYAML))
	require.NoError(t, err)

	logger, err := NewSugaredLogger(provider)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test various log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Test structured logging
	logger.Infow("structured message", "key1", "value1", "key2", 42)

	// Test error logging
	logger.Errorw("error with context", "error", "test error", "code", 500)
}
