package core

import (
	"os"

	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig represents the logging configuration from the config files
type LoggingConfig struct {
	Level       string   `yaml:"level"`
	Development bool     `yaml:"development"`
	Encoding    string   `yaml:"encoding"`
	OutputPaths []string `yaml:"outputPaths"`
}

// Module provides the logger dependencies
var LoggerModule = fx.Options(
	fx.Provide(NewSugaredLogger),
	fx.Provide(NewLogger),
)

func NewLogger(sugar *zap.SugaredLogger) *zap.Logger {
	return sugar.Desugar()
}

// NewSugaredLogger creates a new zap.SugaredLogger based on the configuration
func NewSugaredLogger(provider config.Provider) (*zap.SugaredLogger, error) {
	var loggingConfig LoggingConfig

	// Get logging configuration from the provider
	if err := provider.Get("logging").Populate(&loggingConfig); err != nil {
		return nil, err
	}

	// Parse the log level
	level, err := zapcore.ParseLevel(loggingConfig.Level)
	if err != nil {
		return nil, err
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	if loggingConfig.Development {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	// Create the encoder based on the configuration
	var encoder zapcore.Encoder
	switch loggingConfig.Encoding {
	case "json":
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	case "console":
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	default:
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create the core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout), // Default to stdout, can be enhanced to support multiple outputs
		level,
	)

	// Create the logger
	var logger *zap.Logger
	if loggingConfig.Development {
		logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	} else {
		logger = zap.New(core)
	}

	return logger.Sugar(), nil
}
