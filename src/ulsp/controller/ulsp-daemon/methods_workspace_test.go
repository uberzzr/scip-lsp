package ulspdaemon

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestWorkspaceMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID: uuid.Must(uuid.NewV4()),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	c := controller{
		logger:        logger.Sugar(),
		pluginMethods: sampleWorkspaceMethods(s.UUID),
		sessions:      sessionRepository,
	}

	t.Run("ExecuteCommand", func(t *testing.T) {
		params := &protocol.ExecuteCommandParams{}
		result, err := c.ExecuteCommand(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
		assert.Nil(t, result)
	})
}

// sampleDocumentMethods a sample of RuntimePrioritizedMethods to be used for testing.
// For each method, simulates two assigned plugins: the first returns nil and the second returns an error.
func sampleWorkspaceMethods(id uuid.UUID) map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods {
	err := errors.New("sample")
	m := []*ulspplugin.Methods{
		{
			ExecuteCommand: func(ctx context.Context, params *protocol.ExecuteCommandParams) error {
				return nil
			},
		},
		{
			ExecuteCommand: func(ctx context.Context, params *protocol.ExecuteCommandParams) error {
				return err
			},
		},
	}

	methodLists := ulspplugin.MethodLists{
		Sync:  m,
		Async: m,
	}

	result := make(ulspplugin.RuntimePrioritizedMethods)
	for _, val := range []string{
		protocol.MethodWorkspaceExecuteCommand,
	} {
		result[val] = methodLists
	}

	return map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{id: result}

}
