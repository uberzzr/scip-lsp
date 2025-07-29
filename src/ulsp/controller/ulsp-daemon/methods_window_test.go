package ulspdaemon

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/uber/scip-lsp/src/ulsp/entity"
	ulspplugin "github.com/uber/scip-lsp/src/ulsp/entity/ulsp-plugin"
	"github.com/uber/scip-lsp/src/ulsp/factory"
	"github.com/uber/scip-lsp/src/ulsp/repository/session/repositorymock"
	"go.lsp.dev/protocol"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestWindowMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	s := &entity.Session{
		UUID: factory.UUID(),
	}
	ctx := context.WithValue(context.Background(), entity.SessionContextKey, s.UUID)

	sessionRepository := repositorymock.NewMockRepository(ctrl)
	sessionRepository.EXPECT().GetFromContext(gomock.Any()).Return(s, nil).AnyTimes()

	core, recorded := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	c := controller{
		logger:        logger.Sugar(),
		pluginMethods: sampleWindowMethods(s.UUID),
		sessions:      sessionRepository,
	}

	t.Run("WorkDoneProgressCancel", func(t *testing.T) {
		params := &protocol.WorkDoneProgressCancelParams{}
		err := c.WorkDoneProgressCancel(ctx, params)
		c.wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(recorded.TakeAll()))
	})
}

// sampleWindowMethods a sample of RuntimePrioritizedMethods to be used for testing.
// For each method, simulates two assigned plugins: the first returns nil and the second returns an error.
func sampleWindowMethods(id uuid.UUID) map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods {
	err := errors.New("sample")
	m := []*ulspplugin.Methods{
		{
			WorkDoneProgressCancel: func(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
				return nil
			},
		},
		{
			WorkDoneProgressCancel: func(ctx context.Context, params *protocol.WorkDoneProgressCancelParams) error {
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
		protocol.MethodWorkDoneProgressCancel,
	} {
		result[val] = methodLists
	}

	return map[uuid.UUID]ulspplugin.RuntimePrioritizedMethods{id: result}

}
