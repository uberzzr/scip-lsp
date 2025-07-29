package executor

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// Instantiates the new Executor through fx provider
func fxExecutor(t *testing.T) (Executor, *observer.ObservedLogs) {
	var e Executor
	core, recorded := observer.New(zap.InfoLevel)
	logger := zap.New(core).Sugar()

	fxtest.New(t,
		fx.Provide(
			func() Executor {
				return NewExecutor(WithLogger(logger))
			},
		),
		fx.Populate(&e),
	).RequireStart().RequireStop()

	return e, recorded
}

func TestRunCommand(t *testing.T) {
	e, recorded := fxExecutor(t)

	t.Run("RunCommandWithoutStdin", func(t *testing.T) {
		binPath, err := exec.LookPath("true")
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("no true available")
		}
		require.NoError(t, err)

		cmd := exec.Command("true", "1", "2")
		cmd.Dir = "/"
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err = e.RunCommand(cmd, env)
		assert.NoError(t, err)
		logs := recorded.TakeAll()
		require.Len(t, logs, 1)
		assert.Equal(t, map[string]interface{}{
			"Path": binPath,
			"Dir":  "/",
			"Args": []interface{}{"1", "2"},
		}, logs[0].ContextMap())
	})

	t.Run("RunCommandWithStdin", func(t *testing.T) {
		binPath, err := exec.LookPath("true")
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("no true available")
		}
		require.NoError(t, err)

		cmd := exec.Command("true", "1", "2")
		cmd.Dir = "/"
		cmd.Stdin = strings.NewReader("SomeInput")
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err = e.RunCommand(cmd, env)
		assert.NoError(t, err)
		logs := recorded.TakeAll()
		require.Len(t, logs, 1)
		assert.Equal(t, map[string]interface{}{
			"Path":  binPath,
			"Dir":   "/",
			"Args":  []interface{}{"1", "2"},
			"Stdin": "SomeInput",
		}, logs[0].ContextMap())
	})

	t.Run("fail", func(t *testing.T) {
		binPath, err := exec.LookPath("false")
		if errors.Is(err, exec.ErrNotFound) {
			t.Skip("no false available")
		}
		require.NoError(t, err)

		cmd := exec.Command("false", "3", "4")
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err = e.RunCommand(cmd, env)
		assert.Error(t, err)
		logs := recorded.TakeAll()
		require.Len(t, logs, 1)
		assert.Equal(t, map[string]interface{}{
			"Path": binPath,
			"Dir":  "",
			"Args": []interface{}{"3", "4"},
		}, logs[0].ContextMap())
	})
}

func TestRun(t *testing.T) {
	tempDir := t.TempDir()
	e, _ := fxExecutor(t)

	t.Run("ls", func(t *testing.T) {
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err := e.RunCommand(exec.Command("ls", tempDir), env)
		assert.NoError(t, err)

		cmd := exec.Command("ls")
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Equal(t, "", stdOut)
		assert.Empty(t, stdErr)
		assert.Equal(t, 0, exitCode)
		assert.NoError(t, err)
	})

	t.Run("touch", func(t *testing.T) {
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err := e.RunCommand(exec.Command("touch", filepath.Join(tempDir, "1.txt")), env)
		assert.NoError(t, err)

		cmd := exec.Command("touch", "2.txt")
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Equal(t, "", stdOut)
		assert.Empty(t, stdErr)
		assert.Equal(t, 0, exitCode)
		assert.NoError(t, err)
	})

	t.Run("ls", func(t *testing.T) {
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err := e.RunCommand(exec.Command("ls", tempDir), env)
		assert.NoError(t, err)

		cmd := exec.Command("ls")
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Equal(t, "1.txt\n2.txt\n", stdOut)
		assert.Empty(t, stdErr)
		assert.Equal(t, 0, exitCode)
		assert.NoError(t, err)
	})

	t.Run("rm", func(t *testing.T) {
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err := e.RunCommand(exec.Command("rm", filepath.Join(tempDir, "1.txt")), env)
		assert.NoError(t, err)

		cmd := exec.Command("rm", "2.txt")
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Equal(t, "", stdOut)
		assert.Empty(t, stdErr)
		assert.Equal(t, 0, exitCode)
		assert.NoError(t, err)
	})

	t.Run("ls", func(t *testing.T) {
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err := e.RunCommand(exec.Command("ls", tempDir), env)
		assert.NoError(t, err)

		cmd := exec.Command("ls")
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Equal(t, "", stdOut)
		assert.Empty(t, stdErr)
		assert.Equal(t, 0, exitCode)
		assert.NoError(t, err)
	})
}

func TestRunFails(t *testing.T) {
	tempDir := t.TempDir()
	e, _ := fxExecutor(t)

	t.Run("rm dir", func(t *testing.T) {
		env := []string{"KEY1=VAL1", "KEY2=VAL2"}
		err := e.RunCommand(exec.Command("rm", tempDir), env)
		assert.Error(t, err)
		assert.Equal(t, "exit status 1", err.Error())

		cmd := exec.Command("rm", tempDir)
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Empty(t, string(stdOut))
		assert.Contains(t, strings.ToLower(stdErr), "is a directory")
		assert.Equal(t, 1, exitCode)
		assert.Error(t, err)
		assert.Equal(t, "exit status 1", err.Error())
	})

	t.Run("Unknown Command", func(t *testing.T) {
		cmd := exec.Command("no_valid_command_")
		cmd.Dir = tempDir
		cmd.Env = os.Environ()
		stdOut, stdErr, exitCode, err := e.Run(cmd)

		assert.Empty(t, stdOut)
		assert.Empty(t, stdErr)
		assert.Equal(t, -1, exitCode)
		assert.Error(t, err)
		assert.Equal(t, `exec: "no_valid_command_": executable file not found in $PATH`, err.Error())
	})
}
