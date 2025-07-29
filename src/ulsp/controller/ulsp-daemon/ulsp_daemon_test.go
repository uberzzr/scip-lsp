package ulspdaemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/idl/mock/fxmock"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin/pluginmock"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/protocol"
	"go.uber.org/config"
	"go.uber.org/goleak"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type sampleConfig map[string]interface{}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockShutdowner := fxmock.NewMockShutdowner(ctrl)

	s := &entity.Session{
		UUID: factory.UUID(),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	mockParams := Params{
		Shutdowner: mockShutdowner,
		Logger:     zap.NewNop().Sugar(),
		Sessions:   sessionRepository,
	}

	t.Run("config includes timeout", func(t *testing.T) {
		mockConfig, _ := config.NewStaticProvider(sampleConfig{
			_idleTimeoutMinutesKey: 5,
		})
		mockParams.Config = mockConfig

		assert.NotPanics(t, func() {
			mockShutdowner.EXPECT().Shutdown().Return(nil)
			c, _ := New(mockParams)
			c.RequestFullShutdown(ctx)
			c.Exit(ctx)

			// Small delay to allow shutdown goroutine to complete.
			time.Sleep(100 * time.Millisecond)
		})
	})

	t.Run("config missing timeout", func(t *testing.T) {
		mockConfig, _ := config.NewStaticProvider(sampleConfig{})
		mockParams.Config = mockConfig

		_, err := New(mockParams)
		assert.Error(t, err)
	})
}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestRegisterPlugins(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	// Start with 3 valid plugins.
	availablePlugins := []ulspplugin.Plugin{}
	nameKeys := map[string]bool{}
	for i := 0; i < 3; i++ {
		newPlugin := pluginmock.NewMockPlugin(ctrl)
		info := factory.PluginInfoValid(i)
		nameKeys[info.NameKey] = true
		newPlugin.EXPECT().StartupInfo(gomock.Any()).Return(info, nil).AnyTimes()
		availablePlugins = append(availablePlugins, newPlugin)
	}

	t.Run("valid plugins", func(t *testing.T) {
		c := controller{
			logger:        zap.NewNop().Sugar(),
			sessions:      sessionRepository,
			pluginMethods: map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{},
			pluginConfig:  nameKeys,
			pluginsAll:    availablePlugins,
		}

		err := c.registerSessionPlugins(ctx)
		assert.NoError(t, err)

		for i := 0; i < 3; i++ {
			original, _ := availablePlugins[i].StartupInfo(ctx)
			assert.Equal(t, original.Methods, c.pluginMethods[s.UUID][protocol.MethodTextDocumentDidOpen].Sync[i])
		}
	})

	t.Run("invalid plugin", func(t *testing.T) {
		c := controller{
			logger:        zap.NewNop().Sugar(),
			sessions:      sessionRepository,
			pluginMethods: map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{},
			pluginConfig:  nameKeys,
			pluginsAll:    availablePlugins,
		}

		newPlugin := pluginmock.NewMockPlugin(ctrl)
		info := factory.PluginInfoInvalid(10)
		nameKeys[info.NameKey] = true

		newPlugin.EXPECT().StartupInfo(gomock.Any()).Return(info, nil)
		availablePlugins[1] = newPlugin

		err := c.registerSessionPlugins(ctx)
		assert.Error(t, err)
	})

	t.Run("StartupInfo error", func(t *testing.T) {
		c := controller{
			logger:        zap.NewNop().Sugar(),
			sessions:      sessionRepository,
			pluginMethods: map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{},
			pluginConfig:  nameKeys,
			pluginsAll:    availablePlugins,
		}

		newPlugin := pluginmock.NewMockPlugin(ctrl)
		newPlugin.EXPECT().StartupInfo(gomock.Any()).Return(ulspplugin.PluginInfo{}, errors.New("startup info error"))
		availablePlugins[1] = newPlugin

		err := c.registerSessionPlugins(ctx)
		assert.Error(t, err)
	})
}

func TestExecutePluginMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	c := controller{
		logger:   zap.NewNop().Sugar(),
		sessions: sessionRepository,
	}

	var counter int32
	c.pluginMethods = map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{s.UUID: ulspplugin.RuntimePrioritizedMethods{}}
	c.pluginMethods[s.UUID][protocol.MethodExit] = ulspplugin.MethodLists{
		Sync: []*ulspplugin.Methods{
			{
				Exit: func(ctx context.Context) error {
					atomic.AddInt32(&counter, 1)
					return nil
				},
			},
			{
				Exit: func(ctx context.Context) error {
					atomic.AddInt32(&counter, 1)
					return errors.New("sample")
				},
			},
		},
		Async: []*ulspplugin.Methods{
			{
				Exit: func(ctx context.Context) error {
					atomic.AddInt32(&counter, 1)
					return nil
				},
			},
			{
				Exit: func(ctx context.Context) error {
					atomic.AddInt32(&counter, 1)
					return errors.New("sample")
				},
			},
		},
	}

	sampleFunc := func(ctx context.Context, m *ulspplugin.Methods) {
		m.Exit(ctx)
	}

	t.Run("valid call", func(t *testing.T) {
		counter = 0
		err := c.executePluginMethods(ctx, protocol.MethodExit, sampleFunc, sampleFunc)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 4, int(counter))
	})

	t.Run("missing handler", func(t *testing.T) {
		counter = 0
		err := c.executePluginMethods(ctx, protocol.MethodExit, nil, sampleFunc)
		c.wg.Wait()
		assert.Error(t, err)
		assert.Equal(t, 0, int(counter))
	})

	t.Run("method without registered plugins", func(t *testing.T) {
		counter = 0
		c.executePluginMethods(ctx, protocol.MethodDidCreateFiles, sampleFunc, sampleFunc)
		c.wg.Wait()
		assert.Equal(t, 0, int(counter))
	})

	t.Run("session without registered plugins", func(t *testing.T) {
		s := &entity.Session{
			UUID: factory.UUID(),
		}
		ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)
		err := c.executePluginMethods(ctx, protocol.MethodDidCreateFiles, sampleFunc, sampleFunc)
		assert.NoError(t, err)
	})
}
