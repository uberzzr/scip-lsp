package logfilewriter

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
	"github.com/uber/scip-lsp/src/ulsp/internal/serverinfofile"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const _fmtOutputKey = "output:%s"

// Params define the dependencies for SetupOutputWriter.
type Params struct {
	FS             fs.UlspFS
	Lifecycle      fx.Lifecycle
	ServerInfoFile serverinfofile.ServerInfoFile
}

// SetupOutputWriter creates a writer that will be used to write human readable output to a temporary file for reference by the user.
// This is meant for use in cases where a given plugin may wish to collect its own specific output that is independent of overall server logging.
// The file path will be stored in the server info file for reference by the IDE.
func SetupOutputWriter(p Params, name string) (io.Writer, error) {
	// Output to be stored in a log file under a custom directory in the user's temp directory.
	logsDirPath := filepath.Join(os.TempDir(), name)
	err := p.FS.MkdirAll(logsDirPath)
	if err != nil {
		return nil, err
	}

	logFile, err := p.FS.TempFile(logsDirPath, "")
	if err != nil {
		return nil, err
	}

	// IDE can tail the file by getting the file path from the server info file.
	p.ServerInfoFile.UpdateField(fmt.Sprintf(_fmtOutputKey, name), logFile.Name())

	// Write via a logger for formatting, timestamp, and performance/buffering.
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(logFile),
		zap.InfoLevel,
	)
	bequestLogger := zap.New(core).Sugar()

	// Cleanup on shutdown.
	p.Lifecycle.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			bequestLogger.Sync()
			logFile.Close()
			return p.FS.Remove(logFile.Name())
		},
	})

	return &loggerWriter{logger: bequestLogger}, nil
}

type loggerWriter struct {
	logger *zap.SugaredLogger
}

// Write implements the io.Writer interface by sending data to the given logger.
func (o *loggerWriter) Write(p []byte) (n int, err error) {
	// Incoming data may contain multiple lines, including blank ones.
	// Split and log each line individually.
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		if len(line) > 0 {
			o.logger.Info(line)
		}
	}

	return len(p), nil
}
