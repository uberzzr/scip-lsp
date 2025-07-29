package app

import (
	"fmt"
	"os"
	"path"

	"github.com/uber/scip-lsp/src/ulsp/internal/executor"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Context struct {
	Environment        string `yaml:"environment"`
	RuntimeEnvironment string `yaml:"runtimeEnvironment"`
}

const (
	// EnvLocal indicates that the service is running locally.
	EnvLocal = "local"

	// EnvDevelopment indicates that the service is running in a development environment.
	EnvDevelopment = "development"

	// Environment variables
	_envUlspEnvironment = "ULSP_ENVIRONMENT"
)

func decorateEnvContext(env Context) Context {
	envValue := EnvLocal
	if os.Getenv(_envUlspEnvironment) == EnvDevelopment {
		envValue = EnvDevelopment
	} else {
		envValue = EnvLocal
	}

	env.Environment = envValue
	env.RuntimeEnvironment = envValue
	return env
}

// DecorateConfigParams is the set of dependencies required to decorate the config.Provider.
type DecorateConfigParams struct {
	fx.In

	Env      Context
	Executor executor.Executor
	Cfg      config.Provider
	FS       fs.UlspFS
}

// decorateConfigProvider includes any steps that modify the config.Provider before it is used, or use its data for any startup related activities.
func decorateConfigProvider(p DecorateConfigParams) (config.Provider, error) {
	combined, err := ensureLogFolder(p.Cfg, p.FS)
	if err != nil {
		return nil, fmt.Errorf("ensuring log folder: %v", err)
	}

	return combined, nil
}

// Ensure that all configured logging output directories exist or create if necessary.
func ensureLogFolder(cfg config.Provider, fs fs.UlspFS) (config.Provider, error) {
	var c zap.Config
	if err := cfg.Get("logging").Populate(&c); err != nil {
		return nil, fmt.Errorf("loading logging config: %v", err)
	}

	for _, outputPath := range c.OutputPaths {
		dir := path.Dir(outputPath)
		if err := fs.MkdirAll(dir); err != nil {
			return nil, fmt.Errorf("creating logging directory: %v", err)
		}
	}

	return cfg, nil
}
