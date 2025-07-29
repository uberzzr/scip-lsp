package executor

import (
	"bytes"
	"io"
	"os/exec"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides a module to inject using fx.
var Module = fx.Options(
	fx.Supply(
		fx.Annotate(NewExecutor(
			WithExecFunc(func(cmd *exec.Cmd) error { return cmd.Run() }),
		), fx.As(new(Executor))),
	),
)

// Executor wraps the execution of "os/exec".Cmd's to allow adding logs/metrics to
// each exec and makes it easier to test.
type Executor interface {

	// RunCommand - logs and and executes the Cmd specified
	RunCommand(cmd *exec.Cmd, env []string) error
	// Run - logs and and executes the Cmd specified overriding its Stdout/Stderr to return their content
	Run(cmd *exec.Cmd) (stdout string, stderr string, exitCode int, err error)
}

// executorImp implements Executor
type executorImp struct {
	Logger *zap.SugaredLogger
	// ExecFunc may be nil to use executorImp in tests.
	ExecFunc func(e *exec.Cmd) error
}

// Option defines options to customize executorImp's behavior
type Option func(*executorImp)

// WithLogger overrides the default noop logger
func WithLogger(logger *zap.SugaredLogger) Option {
	return func(executor *executorImp) {
		executor.Logger = logger
	}
}

// WithExecFunc provides customized exec behavior for executorImp
func WithExecFunc(execFunc func(e *exec.Cmd) error) Option {
	return func(executor *executorImp) {
		executor.ExecFunc = execFunc
	}
}

// NewExecutor - creates a new executorImp with logger at the level specified and a default executor function
func NewExecutor(opts ...Option) Executor {
	// See ~/go-code/src/code.uber.internal/devexp/ci-jobs/buildkite/internal/module/logger.go
	executor := &executorImp{
		Logger:   zap.NewNop().Sugar(),
		ExecFunc: func(cmd *exec.Cmd) error { return cmd.Run() },
	}
	for _, opt := range opts {
		opt(executor)
	}
	return executor
}

// RunCommand - logs the Path/Args and calls ExecFunc if it is set.
func (l *executorImp) RunCommand(cmd *exec.Cmd, env []string) error {
	if err := l.logCommand(cmd); err != nil {
		return err
	}

	if l.ExecFunc == nil {
		l.Logger.Warn("missing ExecFunc - skipped execution")
		return nil
	}

	cmd.Env = env
	return l.ExecFunc(cmd)
}

// Run - logs the Path/Args and calls ExecFunc if it is set.
func (l *executorImp) Run(cmd *exec.Cmd) (stdout string, stderr string, exitCode int, err error) {
	if err := l.logCommand(cmd); err != nil {
		return "", "", -1, err
	}

	if l.ExecFunc == nil {
		l.Logger.Warn("missing ExecFunc - skipped execution")
		return "", "", 0, nil
	}

	var stdoutB, stderrB bytes.Buffer
	cmd.Stdout = &stdoutB
	cmd.Stderr = &stderrB
	err = l.ExecFunc(cmd)

	return stdoutB.String(), stderrB.String(), cmd.ProcessState.ExitCode(), err
}

// Logs the command specified: Path, Dir, Args, Stdin (if available)
func (l *executorImp) logCommand(cmd *exec.Cmd) error {
	logKeysAndValues := []interface{}{
		"Path", cmd.Path,
		"Dir", cmd.Dir,
		"Args", cmd.Args[1:], // First arg is always the command itself
	}

	if cmd.Stdin != nil {
		stdinBytes, err := io.ReadAll(cmd.Stdin)
		if err != nil {
			return err
		}
		logKeysAndValues = append(logKeysAndValues, "Stdin", string(stdinBytes))
		cmd.Stdin = bytes.NewReader(stdinBytes)
	}

	l.Logger.Infow("Exec", logKeysAndValues...)
	return nil
}
