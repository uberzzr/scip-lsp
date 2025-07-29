package core

import (
	"fmt"
	"os"
	"path/filepath"

	uber_config "go.uber.org/config"
	"go.uber.org/fx"
)

var ConfigModule = fx.Options(
	fx.Provide(NewConfig),
)

type Config struct {
	provider uber_config.Provider
}

func (c Config) Get(path string) uber_config.Value {
	return c.provider.Get(path)
}

func (c Config) Name() string {
	return "config"
}

func NewConfig() (uber_config.Provider, error) {
	// Get the config directory path
	configDir := getConfigDir()

	// First, load meta.yaml to get the list of configuration files
	metaPath := filepath.Join(configDir, "meta.yaml")
	metaProvider, err := uber_config.NewYAML(
		uber_config.File(metaPath),
		uber_config.Expand(os.LookupEnv),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load meta configuration: %w", err)
	}

	// Get the files list from meta.yaml
	var configFiles []string
	if err := metaProvider.Get("files").Populate(&configFiles); err != nil {
		return nil, fmt.Errorf("failed to read files list from meta.yaml: %w", err)
	}

	// Create a slice to hold valid config files
	var validFiles []string

	// Check which files exist and add them to the list
	for _, file := range configFiles {
		fullPath := filepath.Join(configDir, file)
		if _, err := os.Stat(fullPath); err == nil {
			validFiles = append(validFiles, fullPath)
		}
	}

	if len(validFiles) == 0 {
		return nil, fmt.Errorf("no configuration files found in %s", configDir)
	}

	// Create options for loading all valid files
	var options []uber_config.YAMLOption
	for _, file := range validFiles {
		options = append(options, uber_config.File(file))
	}
	options = append(options, uber_config.Expand(os.LookupEnv))

	// Create the provider with all files and environment variable substitution
	provider, err := uber_config.NewYAML(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	return Config{provider: provider}, nil
}

// getConfigDir returns the path to the configuration directory
func getConfigDir() string {
	// Try to get from environment variable first
	if configDir := os.Getenv("ULSP_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	// Default to the config directory relative to the current working directory
	// This assumes the binary is run from the workspace root
	return "src/ulsp/config"
}
