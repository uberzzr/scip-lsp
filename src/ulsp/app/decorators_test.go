package app

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uber/scip-lsp/src/ulsp/internal/executor"
	executormock "github.com/uber/scip-lsp/src/ulsp/internal/executor/executormock"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs"
	fsmock "github.com/uber/scip-lsp/src/ulsp/internal/fs/fsmock"
	"go.uber.org/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/mock/gomock"
)

func TestEnv(t *testing.T) {

	tests := []struct {
		name      string
		setEnvKey string
		setEnvVal string
		expectVal string
	}{
		{
			name:      "local",
			expectVal: EnvLocal,
		},
		{
			name:      "development",
			setEnvKey: _envUlspEnvironment,
			setEnvVal: "development",
			expectVal: EnvDevelopment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnvKey != "" {
				os.Setenv(tt.setEnvKey, tt.setEnvVal)
				defer os.Unsetenv(tt.setEnvKey)
			}

			fxtest.New(
				t,
				fx.Provide(func() Context {
					return Context{
						Environment:        "local",
						RuntimeEnvironment: "local",
					}
				}),
				fx.Decorate(decorateEnvContext),
				fx.Invoke(func(ctx Context) {
					require.Equal(t, tt.expectVal, ctx.Environment, "unexpected environment")
					require.Equal(t, tt.expectVal, ctx.RuntimeEnvironment, "unexpected runtime environment")
				}),
			).RequireStart().RequireStop()
		})
	}
}

func TestDecorateConfigProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("no errors", func(t *testing.T) {
		fsMock := fsmock.NewMockUlspFS(ctrl)
		fsMock.EXPECT().MkdirAll("/tmp/foo").Return(nil)
		executorMock := executormock.NewMockExecutor(ctrl)

		fxtest.New(
			t,
			fx.Provide(func() fs.UlspFS {
				return fsMock
			}),
			fx.Provide(func() config.Provider {
				p, _ := config.NewStaticProvider(map[string]interface{}{
					"logging": map[string]interface{}{
						"outputPaths": []string{
							"/tmp/foo/myfile1.log",
						},
					},
				})
				return p
			}),
			fx.Provide(func() Context {
				return Context{
					RuntimeEnvironment: EnvDevelopment,
				}
			}),
			fx.Provide(func() executor.Executor {
				return executorMock
			}),
			fx.Decorate(decorateConfigProvider),
			fx.Invoke(func(cfg config.Provider) {
			}),
		).RequireStart().RequireStop()
	})
}

func TestEnsureLogFolder(t *testing.T) {
	ctrl := gomock.NewController(t)
	t.Run("no errors", func(t *testing.T) {
		fsMock := fsmock.NewMockUlspFS(ctrl)

		fsMock.EXPECT().MkdirAll("/tmp/foo").Return(nil)
		fsMock.EXPECT().MkdirAll("/tmp/bar").Return(nil)

		fxtest.New(
			t,
			fx.Provide(func() fs.UlspFS {
				return fsMock
			}),
			fx.Provide(func() config.Provider {
				p, _ := config.NewStaticProvider(map[string]interface{}{
					"logging": map[string]interface{}{
						"outputPaths": []string{
							"/tmp/foo/myfile1.log",
							"/tmp/bar/myfile2.log",
						},
					},
				})
				return p
			}),
			fx.Decorate(ensureLogFolder),
			fx.Invoke(func(cfg config.Provider) {
			}),
		).RequireStart().RequireStop()
	})

	t.Run("error creating directory", func(t *testing.T) {
		fsMock := fsmock.NewMockUlspFS(ctrl)
		fsMock.EXPECT().MkdirAll("/tmp/foo").Return(errors.New("error creating directory"))
		p, _ := config.NewStaticProvider(map[string]interface{}{
			"logging": map[string]interface{}{
				"outputPaths": []string{
					"/tmp/foo/myfile1.log",
					"/tmp/bar/myfile2.log",
				},
			},
		})
		_, err := ensureLogFolder(p, fsMock)
		assert.Error(t, err)
	})
}
