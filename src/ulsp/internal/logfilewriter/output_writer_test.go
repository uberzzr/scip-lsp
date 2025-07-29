package logfilewriter

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs/fsmock"
	"github.com/uber/scip-lsp/src/ulsp/internal/serverinfofile/serverinfofilemock"
	"go.uber.org/fx/fxtest"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSetupOutputWriter(t *testing.T) {
	lifecycleMock := fxtest.NewLifecycle(t)
	ctrl := gomock.NewController(t)
	serverInfoFileMock := serverinfofilemock.NewMockServerInfoFile(ctrl)
	fsMock := fsmock.NewMockUlspFS(ctrl)

	p := Params{
		Lifecycle:      lifecycleMock,
		ServerInfoFile: serverInfoFileMock,
		FS:             fsMock,
	}

	t.Run("success", func(t *testing.T) {
		fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
		file, err := os.CreateTemp(t.TempDir(), "")
		assert.NoError(t, err)
		fsMock.EXPECT().TempFile(gomock.Any(), gomock.Any()).Return(file, nil)
		serverInfoFileMock.EXPECT().UpdateField(fmt.Sprintf(_fmtOutputKey, "sample-key"), file.Name()).Return(nil)

		writer, err := SetupOutputWriter(p, "sample-key")
		assert.NoError(t, err)

		_, err = writer.Write([]byte("sample log message"))
		assert.NoError(t, err)
	})

	t.Run("mkdir fail", func(t *testing.T) {
		fsMock.EXPECT().MkdirAll(gomock.Any()).Return(errors.New("sample"))
		_, err := SetupOutputWriter(p, "sample-key")
		assert.Error(t, err)
	})

	t.Run("tempfile fail", func(t *testing.T) {
		fsMock.EXPECT().MkdirAll(gomock.Any()).Return(nil)
		fsMock.EXPECT().TempFile(gomock.Any(), gomock.Any()).Return(nil, errors.New("sample"))
		_, err := SetupOutputWriter(p, "sample-key")
		assert.Error(t, err)
	})

}

func TestWrite(t *testing.T) {
	// For testing purposes, collect logger results in a buffer.
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(&buf),
		zap.InfoLevel,
	)
	logger := zap.New(core).Sugar()
	sampleWriter := loggerWriter{logger}

	sampleMessage := "sample log message"

	_, err := sampleWriter.Write([]byte(sampleMessage + "\n" + sampleMessage + "\n\n"))
	assert.NoError(t, err)
	assert.True(t, strings.Contains(buf.String(), "sample log message"))
	assert.Len(t, strings.Split(strings.TrimSpace(buf.String()), "\n"), 2)
}
