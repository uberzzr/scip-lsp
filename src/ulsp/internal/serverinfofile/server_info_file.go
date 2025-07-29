package serverinfofile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const _configKeyInfoFile = "serverInfoFilePath"

// Module is the Fx module for this package.
var Module = fx.Provide(New)

// ServerInfoFile is an interface to manage contents of a single server info file.
// It is intended to be used to store connection info for reference by the IDE and other tools, and written to at service launch.
type ServerInfoFile interface {
	UpdateField(key string, value string) error
}

type module struct {
	infofile     string
	logger       *zap.SugaredLogger
	fileContents map[string]string
	mu           sync.Mutex
}

// Params define values to be used by ServerInfoFile.
type Params struct {
	fx.In

	Config    config.Provider
	Lifecycle fx.Lifecycle
	Logger    *zap.SugaredLogger
}

// New creates a new ServerInfoFile which manages contents of a single server info file.
func New(p Params) (ServerInfoFile, error) {
	m := module{
		logger:       p.Logger,
		fileContents: make(map[string]string),
	}

	if err := m.processConfig(p.Config); err != nil {
		return nil, err
	}

	p.Lifecycle.Append(fx.Hook{
		OnStop: m.OnStop,
	})

	return &m, nil
}

func (m *module) OnStop(ctx context.Context) error {
	if m.infofile != "" {
		if err := os.Remove(m.infofile); err != nil {
			return err
		}
	}

	return nil
}

func (m *module) UpdateField(key string, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fileContents[key] = value
	jsonOutput, err := json.Marshal(m.fileContents)
	if err != nil {
		return fmt.Errorf("marshalling json: %w", err)
	}

	if err := os.WriteFile(m.infofile, jsonOutput, 0644); err != nil {
		return fmt.Errorf("creating info file: %w", err)
	}
	m.logger.Infow("connection info saved", zap.String("file", m.infofile), zap.String(key, value))
	return nil
}

func (m *module) processConfig(cfg config.Provider) error {
	val := cfg.Get(_configKeyInfoFile)
	if err := val.Populate(&m.infofile); err != nil {
		// incorrectly formatted config
		return fmt.Errorf("getting config field %q: %w", _configKeyInfoFile, err)
	}

	if m.infofile == "" {
		// yaml is missing either the key or value
		return fmt.Errorf("missing field %q in config", _configKeyInfoFile)
	}

	return nil
}
