package ulspdaemon

import (
	"context"
	"errors"
	"os/exec"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/idl/mock/fxmock"
	"github.com/uber/scip-lsp/idl/mock/jsonrpc2mock"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin/pluginmock"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/gateway/ide-client/ideclientmock"
	"github.com/uber/scip-lsp/src/ulsp/internal/executor/executormock"
	"github.com/uber/scip-lsp/src/ulsp/internal/fs/fsmock"
	workspaceutils "github.com/uber/scip-lsp/src/ulsp/internal/workspace-utils"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitialize(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID:        factory.UUID(),
		UlspEnabled: true,
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	t.Run("initialize success", func(t *testing.T) {
		executorMock := executormock.NewMockExecutor(ctrl)
		sessionRepository := repositorymock.NewMockRepository(ctrl)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()
		sessionRepository.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil)

		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Do(func(cmd *exec.Cmd, env []string) error {
			if cmd.Args[0] == "git" && cmd.Args[1] == "remote" && cmd.Args[2] == "get-url" {
				cmd.Stdout.Write([]byte("sample@code.example.internal:go-code"))
			}

			return nil
		}).AnyTimes()

		core, recorded := observer.New(zap.ErrorLevel)
		logger := zap.New(core)

		samplePlugin1 := pluginmock.NewMockPlugin(ctrl)
		samplePlugin1.EXPECT().StartupInfo(gomock.Any()).Return(ulspplugin.PluginInfo{
			Priorities: map[string]ulspplugin.Priority{
				protocol.MethodInitialize: ulspplugin.PriorityHigh,
			},
			Methods: &ulspplugin.Methods{
				PluginNameKey: "sample1",
				Initialize: func(ctx context.Context, params *protocol.InitializeParams, result *protocol.InitializeResult) error {
					return nil
				},
			},
			NameKey: "sample1",
		}, nil)

		samplePlugin2 := pluginmock.NewMockPlugin(ctrl)
		samplePlugin2.EXPECT().StartupInfo(gomock.Any()).Return(ulspplugin.PluginInfo{
			Priorities: map[string]ulspplugin.Priority{
				protocol.MethodInitialize: ulspplugin.PriorityHigh,
			},
			Methods: &ulspplugin.Methods{
				PluginNameKey: "sample2",
				Initialize: func(ctx context.Context, params *protocol.InitializeParams, result *protocol.InitializeResult) error {
					return errors.New("sample")
				},
			},
			NameKey: "sample2",
		}, nil)

		fsMock := fsmock.NewMockUlspFS(ctrl)
		c := controller{
			logger:        logger.Sugar(),
			sessions:      sessionRepository,
			fs:            fsMock,
			pluginsAll:    []ulspplugin.Plugin{samplePlugin1, samplePlugin2},
			pluginConfig:  map[string]bool{"sample1": true, "sample2": true},
			pluginMethods: map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{s.UUID: ulspplugin.RuntimePrioritizedMethods{}},
			executor:      executorMock,
		}

		c.workspaceUtils = workspaceutils.New(workspaceutils.Params{
			IdeGateway: c.ideGateway,
			Logger:     c.logger,
			FS:         c.fs,
			Executor:   c.executor,
		})

		params := &protocol.InitializeParams{}
		params.WorkspaceFolders = []protocol.WorkspaceFolder{
			{
				URI: "file:///foo/bar",
			},
		}
		fsMock.EXPECT().WorkspaceRoot(gomock.Any()).Return([]byte("sample"), nil).Times(len(params.WorkspaceFolders))

		res, err := c.Initialize(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err, "Unexpected initialize error.")
		assert.Equal(t, res.ServerInfo.Name, "Uber Language Server")
		assert.Equal(t, res.Capabilities.TextDocumentSync, protocol.TextDocumentSyncOptions{
			OpenClose: true,
			Change:    protocol.TextDocumentSyncKindIncremental,
			Save: &protocol.SaveOptions{
				IncludeText: true,
			},
			WillSave:          true,
			WillSaveWaitUntil: true,
		})
		assert.Equal(t, 1, recorded.Len())
	})

	t.Run("missing session uuid in context", func(t *testing.T) {
		ctx := context.Background()

		sessionRepository := repositorymock.NewMockRepository(ctrl)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(nil, errors.New("sample"))
		c := controller{
			sessions: sessionRepository,
		}

		params := &protocol.InitializeParams{}
		_, err := c.Initialize(ctx, params)
		assert.Error(t, err)
	})

	t.Run("get workspace root failure", func(t *testing.T) {
		sessionRepository := repositorymock.NewMockRepository(ctrl)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)

		fsMock := fsmock.NewMockUlspFS(ctrl)
		c := controller{
			logger:   zap.NewNop().Sugar(),
			sessions: sessionRepository,
			fs:       fsMock,
		}

		c.workspaceUtils = workspaceutils.New(workspaceutils.Params{
			IdeGateway: c.ideGateway,
			Logger:     c.logger,
			FS:         c.fs,
			Executor:   c.executor,
		})

		params := &protocol.InitializeParams{}
		params.WorkspaceFolders = []protocol.WorkspaceFolder{
			{
				URI: "file:///foo/bar",
			},
		}
		fsMock.EXPECT().WorkspaceRoot(gomock.Any()).Return(nil, errors.New("sample")).Times(len(params.WorkspaceFolders))
		sessionRepository.EXPECT().Set(gomock.Any(), gomock.Any()).Do(func(ctx context.Context, s *entity.Session) error {
			assert.False(t, s.UlspEnabled)
			return nil
		})

		result, err := c.Initialize(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, protocol.ServerCapabilities{}, result.Capabilities)
	})

	t.Run("get repo name failure", func(t *testing.T) {
		sessionRepository := repositorymock.NewMockRepository(ctrl)
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil)

		executorMock := executormock.NewMockExecutor(ctrl)
		executorMock.EXPECT().RunCommand(gomock.Any(), gomock.Any()).Return(errors.New("sample"))

		fsMock := fsmock.NewMockUlspFS(ctrl)
		c := controller{
			logger:   zap.NewNop().Sugar(),
			sessions: sessionRepository,
			fs:       fsMock,
			executor: executorMock,
		}

		c.workspaceUtils = workspaceutils.New(workspaceutils.Params{
			IdeGateway: c.ideGateway,
			Logger:     c.logger,
			FS:         c.fs,
			Executor:   c.executor,
		})

		params := &protocol.InitializeParams{}
		params.WorkspaceFolders = []protocol.WorkspaceFolder{
			{
				URI: "file:///foo/bar",
			},
		}
		fsMock.EXPECT().WorkspaceRoot(gomock.Any()).Return([]byte("sample"), nil).Times(len(params.WorkspaceFolders))

		_, err := c.Initialize(ctx, params)
		assert.Error(t, err)
	})
}

func TestInitialized(t *testing.T) {
	ctrl := gomock.NewController(t)
	sEnabled := &entity.Session{
		UUID:        factory.UUID(),
		UlspEnabled: true,
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, sEnabled.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)

	mockIdeGateway := ideclientmock.NewMockGateway(ctrl)
	mockIdeGateway.EXPECT().ShowMessage(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	pluginMethods := map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{sEnabled.UUID: ulspplugin.RuntimePrioritizedMethods{}}
	pluginMethods[sEnabled.UUID][protocol.MethodInitialized] = ulspplugin.MethodLists{
		Sync: []*ulspplugin.Methods{
			{
				Initialized: func(ctx context.Context, params *protocol.InitializedParams) error {
					return nil
				},
			},
			{
				Initialized: func(ctx context.Context, params *protocol.InitializedParams) error {
					return errors.New("sample")
				},
			},
		},
		Async: []*ulspplugin.Methods{
			{
				Initialized: func(ctx context.Context, params *protocol.InitializedParams) error {
					return nil
				},
			},
			{
				Initialized: func(ctx context.Context, params *protocol.InitializedParams) error {
					return errors.New("sample")
				},
			},
		},
	}

	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	c := controller{
		logger:        logger.Sugar(),
		pluginMethods: pluginMethods,
		sessions:      sessionRepository,
		ideGateway:    mockIdeGateway,
	}

	t.Run("initialized success", func(t *testing.T) {
		params := &protocol.InitializedParams{}
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(sEnabled, nil)
		err := c.Initialized(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, recorded.Len())
	})

	t.Run("workspace disabled", func(t *testing.T) {
		params := &protocol.InitializedParams{}
		sDisabled := &entity.Session{
			UUID:        factory.UUID(),
			UlspEnabled: false,
		}
		sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(sDisabled, nil)
		err := c.Initialized(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
	})
}

func TestShutdown(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	pluginMethods := map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{s.UUID: ulspplugin.RuntimePrioritizedMethods{}}
	pluginMethods[s.UUID][protocol.MethodShutdown] = ulspplugin.MethodLists{
		Sync: []*ulspplugin.Methods{
			{
				Shutdown: func(ctx context.Context) error {
					return nil
				},
			},
			{
				Shutdown: func(ctx context.Context) error {
					return errors.New("sample")
				},
			},
		},
		Async: []*ulspplugin.Methods{
			{
				Shutdown: func(ctx context.Context) error {
					return nil
				},
			},
			{
				Shutdown: func(ctx context.Context) error {
					return errors.New("sample")
				},
			},
		},
	}

	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	c := controller{
		logger:        logger.Sugar(),
		pluginMethods: pluginMethods,
		sessions:      sessionRepository,
	}
	c.Shutdown(ctx)
	c.wg.Wait()
	assert.Equal(t, 2, recorded.Len())
}

func TestExit(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockShutdowner := fxmock.NewMockShutdowner(ctrl)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	sessionRepository.EXPECT().SessionCount(gomock.Any()).Return(1, nil)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	mockIdeGateway := ideclientmock.NewMockGateway(ctrl)
	mockIdeGateway.EXPECT().DeregisterClient(gomock.Any(), gomock.Any()).Return(nil)

	pluginMethods := map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{s.UUID: ulspplugin.RuntimePrioritizedMethods{}}
	pluginMethods[s.UUID][protocol.MethodExit] = ulspplugin.MethodLists{
		Sync: []*ulspplugin.Methods{
			{
				Exit: func(ctx context.Context) error {
					return nil
				},
			},
			{
				Exit: func(ctx context.Context) error {
					return errors.New("sample")
				},
			},
		},
		Async: []*ulspplugin.Methods{
			{
				Exit: func(ctx context.Context) error {
					return nil
				},
			},
			{
				Exit: func(ctx context.Context) error {
					return errors.New("sample")
				},
			},
		},
	}

	t.Run("full shutdown enabled", func(t *testing.T) {
		c := controller{
			shutdowner:         mockShutdowner,
			fullShutdown:       true,
			sessions:           sessionRepository,
			idleTimeoutMinutes: time.Duration(5) * time.Minute,
			pluginMethods:      pluginMethods,
			ideGateway:         mockIdeGateway,
		}
		c.refreshIdleTimer(ctx)

		core, recorded := observer.New(zap.ErrorLevel)
		c.logger = zap.New(core).Sugar()

		mockShutdowner.EXPECT().Shutdown().Return(nil).Times(1)
		c.Exit(ctx)
		c.wg.Wait()
		assert.Equal(t, 2, recorded.Len())
	})

	t.Run("full shutdown disabled", func(t *testing.T) {
		c := controller{
			shutdowner:         mockShutdowner,
			fullShutdown:       false,
			sessions:           sessionRepository,
			idleTimeoutMinutes: time.Duration(5) * time.Minute,
			pluginMethods:      pluginMethods,
			ideGateway:         mockIdeGateway,
		}
		c.refreshIdleTimer(ctx)

		core, recorded := observer.New(zap.ErrorLevel)
		c.logger = zap.New(core).Sugar()

		sessionRepository.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil)

		c.Exit(ctx)
		c.wg.Wait()
		assert.Equal(t, 2, recorded.Len())

		// Ensure proper cleanup of running goroutine by calling again with full shutdown enabled.
		mockShutdowner.EXPECT().Shutdown().Return(nil).Times(1)
		c.fullShutdown = true
		c.Exit(ctx)
		time.Sleep(100 * time.Millisecond)
	})
}

func TestRequestFullShutdown(t *testing.T) {
	c := controller{}

	// fullShutdown is set to true
	assert.False(t, c.fullShutdown)
	c.RequestFullShutdown(context.Background())
	assert.True(t, c.fullShutdown)

	// Duplicate calls have no effect
	c.RequestFullShutdown(context.Background())
	assert.True(t, c.fullShutdown)
}

func TestInitSession(t *testing.T) {
	ctrl := gomock.NewController(t)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	mockShutdowner := fxmock.NewMockShutdowner(ctrl)
	mockShutdowner.EXPECT().Shutdown().Return(nil).AnyTimes()

	mockIdeGateway := ideclientmock.NewMockGateway(ctrl)
	mockIdeGateway.EXPECT().RegisterClient(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	c := controller{
		sessions:           sessionRepository,
		shutdowner:         mockShutdowner,
		logger:             zap.NewNop().Sugar(),
		idleTimer:          time.NewTimer(time.Hour),
		idleTimeoutMinutes: time.Hour,
		ideGateway:         mockIdeGateway,
	}

	mockConn := jsonrpc2mock.NewMockConn(ctrl)
	var conn jsonrpc2.Conn = mockConn

	t.Run("value set successfully", func(t *testing.T) {
		sessionRepository.EXPECT().Set(gomock.Any(), gomock.Any()).Return(nil)
		sessionRepository.EXPECT().SessionCount(gomock.Any()).Return(1, nil)
		id, err := c.InitSession(ctx, &conn)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.UUID{}, id)

		// Timer should be stopped when a value is set.
		assert.False(t, c.idleTimer.Stop())
	})

	t.Run("error setting value", func(t *testing.T) {
		sessionRepository.EXPECT().Set(gomock.Any(), gomock.Any()).Return(errors.New("error"))
		sessionRepository.EXPECT().SessionCount(gomock.Any()).Return(0, nil)
		_, err := c.InitSession(ctx, &conn)
		assert.Error(t, err)

		// Timer should be running when no sessions are active.
		assert.True(t, c.idleTimer.Stop())
	})
}

func TestEndSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	sessionRepository := repositorymock.NewMockRepository(ctrl)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	sessionRepository.EXPECT().SessionCount(gomock.Any()).Return(1, nil).AnyTimes()
	sessionRepository.EXPECT().Delete(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	mockIdeGateway := ideclientmock.NewMockGateway(ctrl)
	mockIdeGateway.EXPECT().DeregisterClient(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	pluginMethods := map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{s.UUID: ulspplugin.RuntimePrioritizedMethods{}}
	pluginMethods[s.UUID][ulspplugin.MethodEndSession] = ulspplugin.MethodLists{
		Sync: []*ulspplugin.Methods{
			{
				EndSession: func(ctx context.Context, uuid uuid.UUID) error {
					return nil
				},
			},
			{
				EndSession: func(ctx context.Context, uuid uuid.UUID) error {
					return errors.New("sample")
				},
			},
		},
		Async: []*ulspplugin.Methods{
			{
				EndSession: func(ctx context.Context, uuid uuid.UUID) error {
					return nil
				},
			},
			{
				EndSession: func(ctx context.Context, uuid uuid.UUID) error {
					return errors.New("sample")
				},
			},
		},
	}

	t.Run("plugins registered", func(t *testing.T) {
		c := controller{
			sessions:           sessionRepository,
			idleTimeoutMinutes: time.Duration(5) * time.Minute,
			pluginMethods:      pluginMethods,
			ideGateway:         mockIdeGateway,
			idleTimer:          time.NewTimer(time.Hour),
		}
		c.refreshIdleTimer(ctx)

		core, recorded := observer.New(zap.ErrorLevel)
		c.logger = zap.New(core).Sugar()

		c.EndSession(ctx, s.UUID)
		c.wg.Wait()
		assert.Equal(t, 2, recorded.Len())
	})

	t.Run("no plugins registered", func(t *testing.T) {
		c := controller{
			sessions:           sessionRepository,
			idleTimeoutMinutes: time.Duration(5) * time.Minute,
			pluginMethods:      map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{},
			ideGateway:         mockIdeGateway,
			idleTimer:          time.NewTimer(time.Hour),
		}
		c.refreshIdleTimer(ctx)

		core, recorded := observer.New(zap.ErrorLevel)
		c.logger = zap.New(core).Sugar()

		err := c.EndSession(ctx, s.UUID)
		c.wg.Wait()
		assert.Equal(t, 0, recorded.Len())
		assert.NoError(t, err)
	})

}
