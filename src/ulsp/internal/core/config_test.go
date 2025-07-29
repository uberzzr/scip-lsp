package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	t.Skip() // TODO: @JamyDev look into config resolve safety
	tests := []struct {
		name        string
		setupEnv    func()
		expectError bool
	}{
		{
			name: "loads config from default directory",
			setupEnv: func() {
				os.Unsetenv("ULSP_CONFIG_DIR")
			},
			expectError: false,
		},
		{
			name: "loads config from custom directory via env var",
			setupEnv: func() {
				os.Setenv("ULSP_CONFIG_DIR", "src/ulsp/config")
			},
			expectError: false,
		},
		{
			name: "fails when config directory doesn't exist",
			setupEnv: func() {
				os.Setenv("ULSP_CONFIG_DIR", "/nonexistent/path")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			t.Cleanup(func() {
				os.Unsetenv("ULSP_CONFIG_DIR")
			})

			provider, err := NewConfig()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)

				// Test that we can get configuration values
				config := provider.(Config)

				// Test getting a known configuration value
				serviceName := config.Get("service.name")
				assert.True(t, serviceName.HasValue())
				assert.Equal(t, "ulsp-daemon", serviceName.String())

				// Test getting a nested configuration value
				loggingLevel := config.Get("logging.level")
				assert.True(t, loggingLevel.HasValue())
			}
		})
	}
}

func TestConfig_Get(t *testing.T) {
	t.Skip() // TODO: @JamyDev look into config resolve safety
	provider, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, provider)

	config := provider.(Config)

	tests := []struct {
		name     string
		path     string
		expected string
		hasValue bool
	}{
		{
			name:     "gets service name",
			path:     "service.name",
			expected: "ulsp-daemon",
			hasValue: true,
		},
		{
			name:     "gets logging level",
			path:     "logging.level",
			expected: "info",
			hasValue: true,
		},
		{
			name:     "gets nested configuration",
			path:     "yarpc.inbounds.http.address",
			expected: ":${ULSP_PORT_HTTP:27881}",
			hasValue: true,
		},
		{
			name:     "returns empty value for non-existent path",
			path:     "nonexistent.path",
			expected: "",
			hasValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := config.Get(tt.path)
			assert.Equal(t, tt.hasValue, value.HasValue())
			if tt.hasValue {
				assert.Equal(t, tt.expected, value.String())
			}
		})
	}
}

func TestConfig_Name(t *testing.T) {
	t.Skip() // TODO: @JamyDev look into config resolve safety
	provider, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, provider)

	config := provider.(Config)
	assert.Equal(t, "config", config.Name())
}

func TestGetConfigDir(t *testing.T) {
	tests := []struct {
		name           string
		setupEnv       func()
		expectedResult string
	}{
		{
			name: "returns environment variable when set",
			setupEnv: func() {
				os.Setenv("ULSP_CONFIG_DIR", "/custom/config/path")
			},
			expectedResult: "/custom/config/path",
		},
		{
			name: "returns default path when environment variable not set",
			setupEnv: func() {
				os.Unsetenv("ULSP_CONFIG_DIR")
			},
			expectedResult: "src/ulsp/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			t.Cleanup(func() {
				os.Unsetenv("ULSP_CONFIG_DIR")
			})

			result := getConfigDir()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestConfigWithEnvironmentVariables(t *testing.T) {
	t.Skip() // TODO: @JamyDev look into config resolve safety
	// Set up environment variables that are referenced in the config
	t.Setenv("ULSP_PORT_HTTP", "8080")
	t.Setenv("ULSP_PORT_GRPC", "9090")
	t.Setenv("HOME", "/test/home")

	provider, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, provider)

	config := provider.(Config)

	// Test that environment variables are properly substituted
	httpAddress := config.Get("yarpc.inbounds.http.address")
	assert.True(t, httpAddress.HasValue())
	assert.Equal(t, ":8080", httpAddress.String())

	grpcAddress := config.Get("yarpc.inbounds.grpc.address")
	assert.True(t, grpcAddress.HasValue())
	assert.Equal(t, ":9090", grpcAddress.String())

	// Test that default values work when environment variables are not set
	t.Setenv("ULSP_PORT_HTTP", "")
	httpAddressDefault := config.Get("yarpc.inbounds.http.address")
	assert.True(t, httpAddressDefault.HasValue())
	assert.Equal(t, ":27881", httpAddressDefault.String())
}

func TestConfigFilePriority(t *testing.T) {
	t.Skip() // TODO: @JamyDev look into config resolve safety
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create test config files with different values
	baseConfig := `service:
  name: base-service
logging:
  level: info`

	devConfig := `service:
  name: dev-service
logging:
  level: debug`

	localConfig := `logging:
  level: warn`

	// Write config files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "base.yaml"), []byte(baseConfig), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "development.yaml"), []byte(devConfig), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "local.yaml"), []byte(localConfig), 0644))

	// Set the config directory to our temp directory
	t.Setenv("ULSP_CONFIG_DIR", tempDir)

	provider, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, provider)

	config := provider.(Config)

	// Test that later files override earlier ones
	serviceName := config.Get("service.name")
	assert.True(t, serviceName.HasValue())
	assert.Equal(t, "dev-service", serviceName.String()) // Should be from development.yaml

	loggingLevel := config.Get("logging.level")
	assert.True(t, loggingLevel.HasValue())
	assert.Equal(t, "warn", loggingLevel.String()) // Should be from local.yaml
}
